package discovery

import (
	"reflect"
	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/youtube/vitess/go/vt/topo"
	"github.com/youtube/vitess/go/vt/zktopo/zktestserver"

	querypb "github.com/youtube/vitess/go/vt/proto/query"
	topodatapb "github.com/youtube/vitess/go/vt/proto/topodata"
)

func TestFindAllKeyspaceShards(t *testing.T) {
	ctx := context.Background()
	ts := zktestserver.New(t, []string{"cell1", "cell2"})

	// no keyspace / shards
	ks, err := findAllKeyspaceShards(ctx, ts, "cell1")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(ks) > 0 {
		t.Errorf("why did I get anything? %v", ks)
	}

	// add one
	if err := ts.UpdateSrvKeyspace(ctx, "cell1", "test_keyspace", &topodatapb.SrvKeyspace{
		Partitions: []*topodatapb.SrvKeyspace_KeyspacePartition{
			{
				ServedType: topodatapb.TabletType_MASTER,
				ShardReferences: []*topodatapb.ShardReference{
					{
						Name: "test_shard0",
					},
				},
			},
		},
	}); err != nil {
		t.Fatalf("can't add srvKeyspace: %v", err)
	}

	// get it
	ks, err = findAllKeyspaceShards(ctx, ts, "cell1")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(ks, map[keyspaceShard]bool{
		keyspaceShard{
			keyspace: "test_keyspace",
			shard:    "test_shard0",
		}: true,
	}) {
		t.Errorf("got wrong value: %v", ks)
	}

	// add another one
	if err := ts.UpdateSrvKeyspace(ctx, "cell1", "test_keyspace2", &topodatapb.SrvKeyspace{
		Partitions: []*topodatapb.SrvKeyspace_KeyspacePartition{
			{
				ServedType: topodatapb.TabletType_MASTER,
				ShardReferences: []*topodatapb.ShardReference{
					{
						Name: "test_shard1",
					},
				},
			},
			{
				ServedType: topodatapb.TabletType_MASTER,
				ShardReferences: []*topodatapb.ShardReference{
					{
						Name: "test_shard2",
					},
				},
			},
		},
	}); err != nil {
		t.Fatalf("can't add srvKeyspace: %v", err)
	}

	// get it
	ks, err = findAllKeyspaceShards(ctx, ts, "cell1")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(ks, map[keyspaceShard]bool{
		keyspaceShard{
			keyspace: "test_keyspace",
			shard:    "test_shard0",
		}: true,
		keyspaceShard{
			keyspace: "test_keyspace2",
			shard:    "test_shard1",
		}: true,
		keyspaceShard{
			keyspace: "test_keyspace2",
			shard:    "test_shard2",
		}: true,
	}) {
		t.Errorf("got wrong value: %v", ks)
	}
}

func TestWaitForTablets(t *testing.T) {
	shortCtx, shortCancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer shortCancel()
	waitAvailableTabletInterval = 20 * time.Millisecond

	ep := topo.NewTablet(0, "a")
	ep.PortMap["vt"] = 1
	input := make(chan *querypb.StreamHealthResponse)
	createFakeConn(ep, input)

	hc := NewHealthCheck(1*time.Millisecond, 1*time.Millisecond, 1*time.Hour, "" /* statsSuffix */)
	hc.AddTablet("cell", "", ep)

	// this should time out
	if err := WaitForTablets(shortCtx, hc, "cell", "keyspace", "shard", []topodatapb.TabletType{topodatapb.TabletType_REPLICA}); err != ErrWaitForTabletsTimeout {
		t.Errorf("got wrong error: %v", err)
	}

	// this should fail, but return a non-timeout error
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := WaitForTablets(cancelledCtx, hc, "cell", "keyspace", "shard", []topodatapb.TabletType{topodatapb.TabletType_REPLICA}); err == nil || err == ErrWaitForTabletsTimeout {
		t.Errorf("want: non-timeout error, got: %v", err)
	}

	// send the tablet in
	shr := &querypb.StreamHealthResponse{
		Target: &querypb.Target{
			Keyspace:   "keyspace",
			Shard:      "shard",
			TabletType: topodatapb.TabletType_REPLICA,
		},
		Serving:       true,
		RealtimeStats: &querypb.RealtimeStats{SecondsBehindMaster: 1, CpuUsage: 0.2},
	}
	input <- shr

	// and ask again, with longer time outs so it's not flaky
	longCtx, longCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer longCancel()
	waitAvailableTabletInterval = 10 * time.Millisecond
	if err := WaitForTablets(longCtx, hc, "cell", "keyspace", "shard", []topodatapb.TabletType{topodatapb.TabletType_REPLICA}); err != nil {
		t.Errorf("got error: %v", err)
	}
}
