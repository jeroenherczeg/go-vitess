// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vtgate

import (
	"reflect"
	"runtime"
	"testing"
	"time"

	tproto "github.com/youtube/vitess/go/vt/tabletserver/proto"
	"github.com/youtube/vitess/go/vt/vtgate/proto"
)

// This file uses the sandbox_test framework.

func init() {
	Init(new(sandboxTopo), "aa", 1*time.Second, 10)
}

func TestVTGateExecuteShard(t *testing.T) {
	resetSandbox()
	sbc := &sandboxConn{}
	testConns[0] = sbc
	q := proto.QueryShard{
		Sql:    "query",
		Shards: []string{"0"},
	}
	qr := new(proto.QueryResult)
	err := RpcVTGate.ExecuteShard(nil, &q, qr)
	if err != nil {
		t.Errorf("want nil, got %v", err)
	}
	if !reflect.DeepEqual(singleRowResult, qr) {
		t.Errorf("want \n%#v, got \n%#v", singleRowResult, qr)
	}
	if qr.Session != nil {
		t.Errorf("want nil, got %#v\n", qr.Session)
	}

	q.Session = new(proto.Session)
	RpcVTGate.Begin(nil, nil, q.Session)
	if !q.Session.InTransaction {
		t.Errorf("want true, got false")
	}
	RpcVTGate.ExecuteShard(nil, &q, qr)
	want := &proto.Session{
		InTransaction: true,
		ShardSessions: []*proto.ShardSession{{
			Shard:         "0",
			TransactionId: 1,
		}},
	}
	if !reflect.DeepEqual(want, q.Session) {
		t.Errorf("want \n%#v, got \n%#v", want, q.Session)
	}

	RpcVTGate.Commit(nil, q.Session, nil)
	if sbc.CommitCount != 1 {
		t.Errorf("want 1, got %d", sbc.CommitCount)
	}

	q.Session = new(proto.Session)
	RpcVTGate.Begin(nil, nil, q.Session)
	RpcVTGate.ExecuteShard(nil, &q, qr)
	RpcVTGate.Rollback(nil, q.Session, nil)
	runtime.Gosched()
	if sbc.RollbackCount != 1 {
		t.Errorf("want 1, got %d", sbc.RollbackCount)
	}
}

func TestVTGateExecuteBatchShard(t *testing.T) {
	resetSandbox()
	testConns[0] = &sandboxConn{}
	testConns[1] = &sandboxConn{}
	q := proto.BatchQueryShard{
		Queries: []tproto.BoundQuery{{
			"query",
			nil,
		}, {
			"query",
			nil,
		}},
		Shards: []string{"0", "1"},
	}
	qrl := new(proto.QueryResultList)
	err := RpcVTGate.ExecuteBatchShard(nil, &q, qrl)
	if err != nil {
		t.Errorf("want nil, got %v", err)
	}
	if len(qrl.List) != 2 {
		t.Errorf("want 2, got %v", len(qrl.List))
	}
	if qrl.List[0].RowsAffected != 2 {
		t.Errorf("want 2, got %v", qrl.List[0].RowsAffected)
	}
	if qrl.Session != nil {
		t.Errorf("want nil, got %#v\n", qrl.Session)
	}

	q.Session = new(proto.Session)
	RpcVTGate.Begin(nil, nil, q.Session)
	err = RpcVTGate.ExecuteBatchShard(nil, &q, qrl)
	if len(q.Session.ShardSessions) != 2 {
		t.Errorf("want 2, got %d", len(q.Session.ShardSessions))
	}
}

func TestVTGateStreamExecuteShard(t *testing.T) {
	resetSandbox()
	sbc := &sandboxConn{}
	testConns[0] = sbc
	q := proto.QueryShard{
		Sql:    "query",
		Shards: []string{"0"},
	}
	var qrs []proto.QueryResult
	err := RpcVTGate.StreamExecuteShard(nil, &q, func(r interface{}) error {
		qrs = append(qrs, *(r.(*proto.QueryResult)))
		return nil
	})
	if err != nil {
		t.Errorf("want nil, got %v", err)
	}
	want := []proto.QueryResult{{QueryResult: singleRowResult.QueryResult}}
	if !reflect.DeepEqual(want, qrs) {
		t.Errorf("want \n%#v, got \n%#v", want, qrs)
	}

	q.Session = new(proto.Session)
	qrs = nil
	RpcVTGate.Begin(nil, nil, q.Session)
	err = RpcVTGate.StreamExecuteShard(nil, &q, func(r interface{}) error {
		qrs = append(qrs, *(r.(*proto.QueryResult)))
		return nil
	})
	want = []proto.QueryResult{{
		QueryResult: singleRowResult.QueryResult,
	}, {
		Session: &proto.Session{
			InTransaction: true,
			ShardSessions: []*proto.ShardSession{{
				Shard:         "0",
				TransactionId: 1,
			}},
		},
	}}
	if !reflect.DeepEqual(want, qrs) {
		t.Errorf("want \n%#v, got \n%#v", want, qrs)
	}
}
