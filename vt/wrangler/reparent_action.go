package wrangler

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"code.google.com/p/vitess/go/relog"
	"code.google.com/p/vitess/go/vt/hook"
	"code.google.com/p/vitess/go/vt/mysqlctl"
	tm "code.google.com/p/vitess/go/vt/tabletmanager"
)

// helper struct to queue up results
type rpcContext struct {
	tablet   *tm.TabletInfo
	position *mysqlctl.ReplicationPosition
	err      error
}

// These functions reimplement a few actions that were originally
// implemented as direct RPCs.  This gives a consistent, if not slower
// mechanism for performing critical actions. It also leaves more
// centralize debug information in zk when a failure occurs.

func (wr *Wrangler) getMasterPosition(ti *tm.TabletInfo) (*mysqlctl.ReplicationPosition, error) {
	actionPath, err := wr.ai.MasterPosition(ti.Path())
	if err != nil {
		return nil, err
	}
	result, err := wr.ai.WaitForCompletionReply(actionPath, wr.actionTimeout())
	if err != nil {
		return nil, err
	}
	return result.(*mysqlctl.ReplicationPosition), nil
}

// Check all the tablets replication positions to find if some
// will have a problem, and suggest a fix for them.
func (wr *Wrangler) checkSlaveReplication(tabletMap map[string]*tm.TabletInfo, masterTabletUid uint32) error {
	relog.Info("Checking all replication positions will allow the transition:")
	masterIsDead := masterTabletUid == tm.NO_TABLET

	// Check everybody has the right master. If there is no master
	// (crash) just check that everyone has the same parent.
	for _, tablet := range tabletMap {
		if masterTabletUid == tm.NO_TABLET {
			masterTabletUid = tablet.Parent.Uid
		}
		if tablet.Parent.Uid != masterTabletUid {
			return fmt.Errorf("tablet %v not slaved correctly, expected %v, found %v", tablet.Path(), masterTabletUid, tablet.Parent.Uid)
		}
	}

	// now check all the replication positions will allow us to proceed
	if masterIsDead {
		relog.Debug("  master is dead, not checking Seconds Behind Master value")
	}
	var lastError error
	mutex := sync.Mutex{}
	wg := sync.WaitGroup{}
	for _, tablet := range tabletMap {
		wg.Add(1)
		go func(tablet *tm.TabletInfo) {
			defer wg.Done()

			if tablet.Type == tm.TYPE_LAG {
				relog.Info("  skipping slave position check for %v tablet %v", tablet.Type, tablet.Path())
				return
			}

			actionPath, err := wr.ai.SlavePosition(tablet.Path())
			if err != nil {
				mutex.Lock()
				lastError = err
				mutex.Unlock()
				relog.Error("  error asking tablet %v for slave position: %v", tablet.Path(), err)
				return
			}
			result, err := wr.ai.WaitForCompletionReply(actionPath, wr.actionTimeout())
			if err != nil {
				mutex.Lock()
				lastError = err
				mutex.Unlock()
				if tablet.Type == tm.TYPE_BACKUP {
					relog.Warning("  failed to get slave position from backup tablet %v, either wait for backup to finish or scrap tablet (%v)", tablet.Path(), err)
				} else {
					relog.Warning("  failed to get slave position from %v: %v", tablet.Path(), err)
				}
				return
			}

			if !masterIsDead {
				replPos := result.(*mysqlctl.ReplicationPosition)
				var dur time.Duration = time.Duration(uint(time.Second) * replPos.SecondsBehindMaster)
				if dur > wr.actionTimeout() {
					err = fmt.Errorf("slave is too far behind to complete reparent in time (%v>%v), either increase timeout using 'vtctl -wait-time XXX ReparentShard ...' or scrap tablet %v", dur, wr.actionTimeout(), tablet.Path())
					relog.Error("  %v", err)
					mutex.Lock()
					lastError = err
					mutex.Unlock()
					return
				}

				relog.Debug("  slave is %v behind master (<%v), reparent should work for %v", dur, wr.actionTimeout(), tablet.Path())
			}
		}(tablet)
	}
	wg.Wait()
	return lastError
}

// Check all the tablets to see if we can proceed with reparenting.
// masterPosition is supplied from the demoted master if we are doing
// this gracefully.
func (wr *Wrangler) checkSlaveConsistency(tabletMap map[uint32]*tm.TabletInfo, masterPosition *mysqlctl.ReplicationPosition) error {
	relog.Debug("checkSlaveConsistency %v %#v", mapKeys(tabletMap), masterPosition)

	// FIXME(msolomon) Something still feels clumsy here and I can't put my finger on it.
	calls := make(chan *rpcContext, len(tabletMap))
	f := func(ti *tm.TabletInfo) {
		ctx := &rpcContext{tablet: ti}
		defer func() {
			calls <- ctx
		}()

		var args *tm.SlavePositionReq
		if masterPosition != nil {
			// If the master position is known, do our best to wait for replication to catch up.
			args = &tm.SlavePositionReq{*masterPosition, int(wr.actionTimeout().Seconds())}
		} else {
			// In the case where a master is down, look for the last bit of data copied and wait
			// for that to apply. That gives us a chance to wait for all data.
			actionPath, err := wr.ai.SlavePosition(ti.Path())
			if err != nil {
				ctx.err = err
				return
			}
			result, err := wr.ai.WaitForCompletionReply(actionPath, wr.actionTimeout())
			if err != nil {
				ctx.err = err
				return
			}
			replPos := result.(*mysqlctl.ReplicationPosition)
			lastDataPos := mysqlctl.ReplicationPosition{MasterLogFile: replPos.MasterLogFileIo,
				MasterLogPositionIo: replPos.MasterLogPositionIo}
			args = &tm.SlavePositionReq{lastDataPos, int(wr.actionTimeout().Seconds())}
		}

		// This option waits for the SQL thread to apply all changes to this instance.
		actionPath, err := wr.ai.WaitSlavePosition(ti.Path(), args)
		if err != nil {
			ctx.err = err
			return
		}
		result, err := wr.ai.WaitForCompletionReply(actionPath, wr.actionTimeout())
		if err != nil {
			ctx.err = err
			return
		}
		ctx.position = result.(*mysqlctl.ReplicationPosition)
	}

	for _, tablet := range tabletMap {
		// Pass loop variable explicitly so we don't have a concurrency issue.
		go f(tablet)
	}

	// map positions to tablets
	positionMap := make(map[string][]uint32)
	for i := 0; i < len(tabletMap); i++ {
		ctx := <-calls
		mapKey := "unavailable-tablet-error"
		if ctx.err == nil {
			mapKey = ctx.position.MapKey()
		}
		if _, ok := positionMap[mapKey]; !ok {
			positionMap[mapKey] = make([]uint32, 0, 32)
		}
		positionMap[mapKey] = append(positionMap[mapKey], ctx.tablet.Uid)
	}

	if len(positionMap) == 1 {
		// great, everyone agrees
		// demotedMasterReplicationState is nil if demotion failed
		if masterPosition != nil {
			demotedMapKey := masterPosition.MapKey()
			if _, ok := positionMap[demotedMapKey]; !ok {
				for slaveMapKey, _ := range positionMap {
					return fmt.Errorf("slave position doesn't match demoted master: %v != %v", demotedMapKey,
						slaveMapKey)
				}
			}
		}
	} else {
		// FIXME(msolomon) in the event of a crash, do you pick replica that is
		// furthest along or do you promote the majority? data loss vs availability
		// sounds like you pick the latest group and reclone.
		items := make([]string, 0, 32)
		for slaveMapKey, uids := range positionMap {
			tabletPaths := make([]string, len(uids))
			for i, uid := range uids {
				tabletPaths[i] = tabletMap[uid].Path()
			}
			items = append(items, fmt.Sprintf("  %v\n    %v", slaveMapKey, strings.Join(tabletPaths, "\n    ")))
		}
		sort.Strings(items)
		return fmt.Errorf("inconsistent slaves, mark some offline with vtctl ScrapTablet\n%v", strings.Join(items, "\n"))
	}
	return nil
}

// Shut off all replication.
func (wr *Wrangler) stopSlaves(tabletMap map[string]*tm.TabletInfo) error {
	errs := make(chan error, len(tabletMap))
	f := func(ti *tm.TabletInfo) {
		actionPath, err := wr.ai.StopSlave(ti.Path())
		if err == nil {
			err = wr.ai.WaitForCompletion(actionPath, wr.actionTimeout())
		}
		if err != nil {
			relog.Debug("StopSlave failed: %v", err)
		}
		errs <- err
	}

	for _, tablet := range tabletMap {
		// Pass loop variable explicitly so we don't have a concurrency issue.
		go f(tablet)
	}

	// wait for responses
	for i := 0; i < len(tabletMap); i++ {
		if err := <-errs; err != nil {
			return err
		}
	}

	return nil
}

// Return a list of corresponding replication positions.
// Handles masters and slaves, but it's up to the caller to guarantee
// all tablets are in the same shard.
func (wr *Wrangler) tabletReplicationPositions(tablets []*tm.TabletInfo) ([]*mysqlctl.ReplicationPosition, error) {
	relog.Debug("tabletReplicationPositions %v", tablets)
	calls := make([]*rpcContext, len(tablets))
	wg := sync.WaitGroup{}

	f := func(idx int) {
		defer wg.Done()
		ti := tablets[idx]
		ctx := &rpcContext{tablet: ti}
		calls[idx] = ctx

		var actionPath string
		if ti.Type == tm.TYPE_MASTER {
			actionPath, ctx.err = wr.ai.MasterPosition(ti.Path())
		} else if ti.IsSlaveType() {
			actionPath, ctx.err = wr.ai.SlavePosition(ti.Path())
		}

		if ctx.err != nil {
			return
		}

		var result interface{}
		if result, ctx.err = wr.ai.WaitForCompletionReply(actionPath, wr.actionTimeout()); ctx.err != nil {
			return
		}
		ctx.position = result.(*mysqlctl.ReplicationPosition)
	}

	for i, tablet := range tablets {
		// Don't scan tablets that won't return something useful. Otherwise, you'll
		// end up waiting for a timeout.
		if tablet.Type == tm.TYPE_MASTER || tablet.IsSlaveType() {
			wg.Add(1)
			go f(i)
		} else {
			relog.Info("tabletReplicationPositions: skipping tablet %v type %v", tablet.Path(), tablet.Type)
		}
	}
	wg.Wait()

	someErrors := false
	positions := make([]*mysqlctl.ReplicationPosition, len(tablets))
	for i, ctx := range calls {
		if ctx == nil {
			continue
		}
		if ctx.err != nil {
			relog.Warning("could not get replication position for tablet %v %v", ctx.tablet.Path(), ctx.err)
			someErrors = true
		} else {
			positions[i] = ctx.position
		}
	}
	if someErrors {
		return positions, fmt.Errorf("partial position map, some errors")
	}
	return positions, nil
}

func (wr *Wrangler) demoteMaster(ti *tm.TabletInfo) (*mysqlctl.ReplicationPosition, error) {
	relog.Info("demote master %v", ti.Path())
	actionPath, err := wr.ai.DemoteMaster(ti.Path())
	if err != nil {
		return nil, err
	}
	err = wr.ai.WaitForCompletion(actionPath, wr.actionTimeout())
	if err != nil {
		return nil, err
	}
	return wr.getMasterPosition(ti)
}

func (wr *Wrangler) promoteSlave(ti *tm.TabletInfo) (rsd *tm.RestartSlaveData, err error) {
	relog.Info("promote slave %v", ti.Path())
	actionPath, err := wr.ai.PromoteSlave(ti.Path())
	if err != nil {
		return
	}
	result, err := wr.ai.WaitForCompletionReply(actionPath, wr.actionTimeout())
	if err != nil {
		return
	}
	rsd = result.(*tm.RestartSlaveData)
	return
}

func (wr *Wrangler) slaveWasPromoted(ti *tm.TabletInfo) error {
	relog.Info("slave was promoted %v", ti.Path())
	actionPath, err := wr.ai.SlaveWasPromoted(ti.Path())
	if err != nil {
		return err
	}
	err = wr.ai.WaitForCompletion(actionPath, wr.actionTimeout())
	if err != nil {
		return err
	}
	return nil
}

func (wr *Wrangler) restartSlaves(slaveTabletMap map[string]*tm.TabletInfo, rsd *tm.RestartSlaveData) (majorityRestart bool, err error) {
	wg := new(sync.WaitGroup)
	slaves := CopyMapValues(slaveTabletMap, []*tm.TabletInfo{}).([]*tm.TabletInfo)
	errs := make([]error, len(slaveTabletMap))

	f := func(i int) {
		errs[i] = wr.restartSlave(slaves[i], rsd)
		if errs[i] != nil {
			// FIXME(msolomon) Don't bail early, just mark this phase as
			// failed. We might decide to proceed if enough of these
			// succeed.
			//
			// FIXME(msolomon) This is a somewhat delicate retry - have to
			// figure out why it failed on the tablet end. This could lead
			// to a nasty case of having to recompute where to start
			// replication. Practically speaking, that chance is pretty low.
			relog.Warning("restart slave failed: %v %v", slaves[i].Path(), errs[i])
		}
		wg.Done()
	}

	for i, _ := range slaves {
		wg.Add(1)
		go f(i)
	}
	wg.Wait()

	errCount := 0
	badTablets := make([]string, 0, 16)
	for i, err := range errs {
		if err != nil {
			errCount++
			badTablets = append(badTablets, slaves[i].Path())
		}
	}
	// Phrase the question with multiplication so we don't get caught by int
	// division rounding.
	majorityRestart = errCount*2 < len(slaveTabletMap)

	if errCount > 0 {
		err = fmt.Errorf("restart slave failed on some tablets (%v): %v", errCount, strings.Join(badTablets, ", "))
	}
	return
}

func (wr *Wrangler) restartSlave(ti *tm.TabletInfo, rsd *tm.RestartSlaveData) (err error) {
	relog.Info("restart slave %v", ti.Path())
	actionPath, err := wr.ai.RestartSlave(ti.Path(), rsd)
	if err != nil {
		return err
	}
	return wr.ai.WaitForCompletion(actionPath, wr.actionTimeout())
}

func (wr *Wrangler) checkMasterElect(ti *tm.TabletInfo) error {
	// Check the master-elect is fit for duty - call out for hardware checks.
	// if the server was already serving live traffic, it's probably good
	if tm.IsServingType(ti.Type) {
		return nil
	}
	return wr.ExecuteOptionalTabletInfoHook(ti, hook.NewSimpleHook("preflight_serving_type"))
}

func (wr *Wrangler) finishReparent(masterElect *tm.TabletInfo, majorityRestart, leaveMasterReadOnly bool) error {
	// If the majority of slaves restarted, move ahead.
	if majorityRestart {
		if leaveMasterReadOnly {
			relog.Warning("leaving master-elect read-only, change with: vtctl SetReadWrite %v", masterElect.Path())
		} else {
			relog.Info("marking master-elect read-write %v", masterElect.Path())
			actionPath, err := wr.ai.SetReadWrite(masterElect.Path())
			if err == nil {
				err = wr.ai.WaitForCompletion(actionPath, wr.actionTimeout())
			}
			if err != nil {
				relog.Warning("master master-elect read-write failed, leaving master-elect read-only, change with: vtctl SetReadWrite %v", masterElect.Path())
			}
		}
	} else {
		relog.Warning("minority reparent, manual fixes are needed, leaving master-elect read-only, change with: vtctl SetReadWrite %v", masterElect.Path())
	}

	relog.Info("rebuilding shard serving graph data in zk")
	return wr.rebuildShard(masterElect.ShardPath())
}

func (wr *Wrangler) breakReplication(slaveMap map[string]*tm.TabletInfo, masterElect *tm.TabletInfo) error {
	// We are forcing a reparenting. Make sure that all slaves stop so
	// no data is accidentally replicated through before we call RestartSlave.
	relog.Info("stop slaves %v", masterElect.Path())
	err := wr.stopSlaves(slaveMap)
	if err != nil {
		return err
	}

	// Force slaves to break, just in case they were not advertised in
	// the replication graph.
	relog.Info("break slaves %v", masterElect.Path())
	actionPath, err := wr.ai.BreakSlaves(masterElect.Path())
	if err == nil {
		err = wr.ai.WaitForCompletion(actionPath, wr.actionTimeout())
	}
	return err
}

func restartableTabletMap(slaves map[string]*tm.TabletInfo) map[uint32]*tm.TabletInfo {
	// Under normal circumstances, prune out lag as not restartable.
	// These types are explicitly excluded from reparenting since you
	// will just wait forever for them to catch up.  A possible
	// improvement is waiting for the io thread to reach the same
	// position as the sql thread on a normal slave.
	tabletMap := make(map[uint32]*tm.TabletInfo)
	for _, ti := range slaves {
		if ti.Type != tm.TYPE_LAG {
			tabletMap[ti.Uid] = ti
		} else {
			relog.Info("skipping reparent action for tablet %v %v", ti.Type, ti.Path())
		}
	}
	return tabletMap
}

func slaveTabletMap(tabletMap map[string]*tm.TabletInfo) (slaveMap map[string]*tm.TabletInfo, master *tm.TabletInfo, err error) {
	slaveMap = make(map[string]*tm.TabletInfo)
	for zkPath, ti := range tabletMap {
		if ti.Type != tm.TYPE_MASTER && ti.Type != tm.TYPE_SCRAP {
			slaveMap[zkPath] = ti
		} else if ti.Parent.Uid == tm.NO_TABLET {
			if master != nil {
				return nil, nil, fmt.Errorf("master tablet conflict in shard %v: %v, %v", master.ShardPath(), master.Path(), ti.Path())
			}
			master = ti
		}
	}
	return
}
