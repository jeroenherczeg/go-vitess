// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package zkwrangler

import (
	"code.google.com/p/vitess/go/relog"
	"code.google.com/p/vitess/go/vt/key"
	tm "code.google.com/p/vitess/go/vt/tabletmanager"
)

// forceMasterSnapshot: Normally a master is not a viable tablet to snapshot.
// However, there are degenerate cases where you need to override this, for
// instance the initial clone of a new master.
func (wr *Wrangler) PartialSnapshot(zkTabletPath, keyName string, startKey, endKey key.HexKeyspaceId, forceMasterSnapshot bool, concurrency int) (manifest, parent string, err error) {
	var ti *tm.TabletInfo
	ti, err = tm.ReadTablet(wr.zconn, zkTabletPath)
	if err != nil {
		return
	}

	originalType := ti.Tablet.Type

	if ti.Tablet.Type == tm.TYPE_MASTER && forceMasterSnapshot {
		// In this case, we don't bother recomputing the serving graph.
		// All queries will have to fail anyway.
		relog.Info("force change type master -> backup: %v", zkTabletPath)
		// There is a legitimate reason to force in the case of a single
		// master.
		ti.Tablet.Type = tm.TYPE_BACKUP
		err = tm.UpdateTablet(wr.zconn, zkTabletPath, ti)
	} else {
		err = wr.ChangeType(zkTabletPath, tm.TYPE_BACKUP, false)
	}

	if err != nil {
		return
	}

	actionPath, err := wr.ai.PartialSnapshot(zkTabletPath, &tm.PartialSnapshotArgs{keyName, startKey, endKey, concurrency})
	if err != nil {
		return
	}

	results, actionErr := wr.ai.WaitForCompletionReply(actionPath, wr.actionTimeout())
	var reply *tm.SnapshotReply
	if actionErr != nil {
		relog.Error("PartialSnapshot failed, still restoring tablet type: %v", actionErr)
		reply = &tm.SnapshotReply{}
	} else {
		reply = results.(*tm.SnapshotReply)
	}

	// Restore type
	relog.Info("change type after snapshot: %v %v", zkTabletPath, originalType)
	if ti.Tablet.Parent.Uid == tm.NO_TABLET && forceMasterSnapshot {
		relog.Info("force change type backup -> master: %v", zkTabletPath)
		ti.Tablet.Type = tm.TYPE_MASTER
		err = tm.UpdateTablet(wr.zconn, zkTabletPath, ti)
	} else {
		err = wr.ChangeType(zkTabletPath, originalType, false)
	}
	if err != nil {
		// failure in changing the zk type is probably worse,
		// so returning that (we logged actionErr anyway)
		return
	}
	return reply.ManifestPath, reply.ZkParentPath, actionErr
}

func (wr *Wrangler) PartialRestore(zkSrcTabletPath, srcFilePath, zkDstTabletPath, zkParentPath string, fetchConcurrency, fetchRetryCount int) error {
	err := wr.ChangeType(zkDstTabletPath, tm.TYPE_RESTORE, false)
	if err != nil {
		return err
	}

	actionPath, err := wr.ai.PartialRestore(zkDstTabletPath, &tm.RestoreArgs{zkSrcTabletPath, srcFilePath, zkParentPath, fetchConcurrency, fetchRetryCount, false})
	if err != nil {
		return err
	}

	if err = wr.ai.WaitForCompletion(actionPath, wr.actionTimeout()); err != nil {
		return err
	}

	// Restore moves us into the replication graph as a spare. There are no
	// consequences to the replication or serving graphs, so no rebuild required.
	return nil
}

func (wr *Wrangler) PartialClone(zkSrcTabletPath, zkDstTabletPath, keyName string, startKey, endKey key.HexKeyspaceId, forceMasterSnapshot bool, concurrency, fetchConcurrency, fetchRetryCount int) error {
	srcFilePath, zkParentPath, err := wr.PartialSnapshot(zkSrcTabletPath, keyName, startKey, endKey, forceMasterSnapshot, concurrency)
	if err != nil {
		return err
	}
	if err := wr.PartialRestore(zkSrcTabletPath, srcFilePath, zkDstTabletPath, zkParentPath, fetchConcurrency, fetchRetryCount); err != nil {
		return err
	}
	return nil
}
