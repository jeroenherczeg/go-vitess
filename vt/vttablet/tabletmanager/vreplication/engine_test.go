/*
Copyright 2018 The Vitess Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package vreplication

import (
	"reflect"
	"testing"

	"golang.org/x/net/context"
	"vitess.io/vitess/go/sqltypes"
	"vitess.io/vitess/go/vt/binlog/binlogplayer"
	"vitess.io/vitess/go/vt/mysqlctl/fakemysqldaemon"

	topodatapb "vitess.io/vitess/go/vt/proto/topodata"
)

func TestEngineExec(t *testing.T) {
	ts := createTopo()
	_ = addTablet(ts, 100, "0", topodatapb.TabletType_REPLICA, true, true)
	_ = newFakeBinlogClient()
	dbClient := binlogplayer.NewVtClientMock()
	dbClient.CommitChannel = make(chan []string, 10)

	// Test Insert

	// insert into _vt.vreplication
	dbClient.AddResult(&sqltypes.Result{InsertID: 1})
	// select * from _vt.vreplication
	dbClient.AddResult(sqltypes.MakeTestResult(
		sqltypes.MakeTestFields(
			"id|state|source|pos",
			"int64|varchar|varchar|varchar",
		),
		`1|Running|keyspace:"ks" shard:"0" key_range:<end:"\200" > |MariaDB/0-1-1083`,
	))
	// select tps
	dbClient.AddResult(testTPSResponse)
	// insert into t
	dbClient.AddResult(testDMLResponse)
	// update _vt.vreplication
	dbClient.AddResult(testDMLResponse)

	dbClientFactory := func() binlogplayer.VtClient { return dbClient }
	mysqld := &fakemysqldaemon.FakeMysqlDaemon{MysqlPort: 3306}

	vre := NewEngine(ts, testCell, mysqld, dbClientFactory)
	_ = vre.Open(context.Background())
	defer vre.Close()

	qr, err := vre.Exec("insert into _vt.vreplication values(null)")
	if err != nil {
		t.Fatal(err)
	}
	wantqr := &sqltypes.Result{InsertID: 1}
	if !reflect.DeepEqual(qr, wantqr) {
		t.Errorf("Exec: %v, want %v", qr, wantqr)
	}

	ct := vre.controllers[1]
	if ct == nil || ct.id != 1 {
		t.Errorf("ct: %v, id should be 1", ct)
	}
	if ct == nil || ct.startPos != "MariaDB/0-1-1083" {
		t.Errorf("ct: %v, startPos should be 'MariaDB/0-1-1083'", ct)
	}

	expectCommit(t, dbClient, []string{
		"insert into _vt.vreplication values (null)",
		"select * from _vt.vreplication where id = 1",
		"SELECT max_tps, max_replication_lag FROM _vt.vreplication WHERE id=1",
		"BEGIN",
		"insert into t values(1)",
		"UPDATE _vt.vreplication SET pos='MariaDB/0-1-1235', time_updated=",
		"COMMIT",
	})

	// Test Update

	// update _vt.vreplication
	dbClient.AddResult(testDMLResponse)
	// select * from _vt.vreplication
	dbClient.AddResult(sqltypes.MakeTestResult(
		sqltypes.MakeTestFields(
			"id|state|source|pos",
			"int64|varchar|varchar|varchar",
		),
		`1|Running|keyspace:"ks" shard:"0" key_range:<end:"\200" > |MariaDB/0-1-1084`,
	))
	// select tps
	dbClient.AddResult(testTPSResponse)
	// insert into t
	dbClient.AddResult(testDMLResponse)
	// update _vt.vreplication
	dbClient.AddResult(testDMLResponse)

	qr, err = vre.Exec("update _vt.vreplication set pos='MariaDB/0-1-1084' where id = 1")
	if err != nil {
		t.Fatal(err)
	}
	wantqr = &sqltypes.Result{RowsAffected: 1}
	if !reflect.DeepEqual(qr, wantqr) {
		t.Errorf("Exec: %v, want %v", qr, wantqr)
	}

	ct = vre.controllers[1]
	if ct == nil || ct.startPos != "MariaDB/0-1-1084" {
		t.Errorf("ct: %v, startPos should be 'MariaDB/0-1-1084'", ct)
	}

	expectCommit(t, dbClient, []string{
		"update _vt.vreplication set pos = 'MariaDB/0-1-1084' where id = 1",
		"select * from _vt.vreplication where id = 1",
		"SELECT max_tps, max_replication_lag FROM _vt.vreplication WHERE id=1",
		"BEGIN",
		"insert into t values(1)",
		"UPDATE _vt.vreplication SET pos='MariaDB/0-1-1235', time_updated=",
		"COMMIT",
	})

	// Test Delete

	// delete _vt.vreplication
	dbClient.AddResult(testDMLResponse)

	qr, err = vre.Exec("delete from _vt.vreplication where id = 1")
	if err != nil {
		t.Fatal(err)
	}
	wantqr = &sqltypes.Result{RowsAffected: 1}
	if !reflect.DeepEqual(qr, wantqr) {
		t.Errorf("Exec: %v, want %v", qr, wantqr)
	}

	ct = vre.controllers[1]
	if ct != nil {
		t.Errorf("ct: %v, want nil", ct)
	}
}

func TestEngineBadInsert(t *testing.T) {
	ts := createTopo()
	_ = addTablet(ts, 100, "0", topodatapb.TabletType_REPLICA, true, true)
	_ = newFakeBinlogClient()

	dbClient := binlogplayer.NewVtClientMock()
	// insert into _vt.vreplication
	dbClient.AddResult(&sqltypes.Result{})
	dbClientFactory := func() binlogplayer.VtClient { return dbClient }
	mysqld := &fakemysqldaemon.FakeMysqlDaemon{MysqlPort: 3306}

	vre := NewEngine(ts, testCell, mysqld, dbClientFactory)
	_ = vre.Open(context.Background())
	defer vre.Close()

	want := "insert failed to generate an id"
	_, err := vre.Exec("insert into _vt.vreplication values(null)")
	if err == nil || err.Error() != want {
		t.Errorf("vre.Exec err: %v, want %v", err, want)
	}
}

func TestEngineSelect(t *testing.T) {
	ts := createTopo()
	_ = addTablet(ts, 100, "0", topodatapb.TabletType_REPLICA, true, true)
	_ = newFakeBinlogClient()
	dbClient := binlogplayer.NewVtClientMock()

	wantResult := sqltypes.MakeTestResult(
		sqltypes.MakeTestFields(
			"id|state|source|pos",
			"int64|varchar|varchar|varchar",
		),
		`1|Running|keyspace:"ks" shard:"0" key_range:<end:"\200" > |MariaDB/0-1-1083`,
	)
	dbClient.AddResult(wantResult)

	dbClientFactory := func() binlogplayer.VtClient { return dbClient }
	mysqld := &fakemysqldaemon.FakeMysqlDaemon{MysqlPort: 3306}

	vre := NewEngine(ts, testCell, mysqld, dbClientFactory)
	_ = vre.Open(context.Background())
	defer vre.Close()

	wantQuery := "select * from _vt.vreplication where workflow = 'x'"
	qr, err := vre.Exec(wantQuery)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(qr, wantResult) {
		t.Errorf("Exec: %v, want %v", qr, wantResult)
	}

	gotQuery := dbClient.Stdout[len(dbClient.Stdout)-1]
	if gotQuery != wantQuery {
		t.Errorf("Query: %v, want %v", gotQuery, wantQuery)
	}
}
