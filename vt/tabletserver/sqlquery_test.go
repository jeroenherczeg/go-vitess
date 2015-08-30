// Copyright 2015, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tabletserver

import (
	"expvar"
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"

	mproto "github.com/youtube/vitess/go/mysql/proto"
	"github.com/youtube/vitess/go/sqltypes"
	pb "github.com/youtube/vitess/go/vt/proto/query"
	"github.com/youtube/vitess/go/vt/proto/topodata"
	"github.com/youtube/vitess/go/vt/proto/vtrpc"
	"github.com/youtube/vitess/go/vt/tabletserver/proto"
	"github.com/youtube/vitess/go/vt/vttest/fakesqldb"
	"golang.org/x/net/context"
)

func TestSqlQueryGetState(t *testing.T) {
	states := []int64{
		StateNotConnected,
		StateNotServing,
		StateServing,
		StateTransitioning,
		StateShuttingDown,
	}
	// Don't reuse stateName.
	names := []string{
		"NOT_SERVING",
		"NOT_SERVING",
		"SERVING",
		"NOT_SERVING",
		"SHUTTING_DOWN",
	}
	setUpSQLQueryTest()
	testUtils := newTestUtils()
	config := testUtils.newQueryServiceConfig()
	sqlQuery := NewSqlQuery(config)
	for i, state := range states {
		sqlQuery.setState(state)
		if stateName := sqlQuery.GetState(); stateName != names[i] {
			t.Errorf("GetState: %s, want %s", stateName, names[i])
		}
	}
}

func TestSqlQueryAllowQueriesFailBadConn(t *testing.T) {
	db := setUpSQLQueryTest()
	db.EnableConnFail()
	testUtils := newTestUtils()
	config := testUtils.newQueryServiceConfig()
	sqlQuery := NewSqlQuery(config)
	checkSQLQueryState(t, sqlQuery, StateNotConnected)
	dbconfigs := testUtils.newDBConfigs()
	err := sqlQuery.StartService(nil, &dbconfigs, []SchemaOverride{}, testUtils.newMysqld(&dbconfigs))
	if err == nil {
		t.Fatalf("SqlQuery.StartService should fail")
	}
	checkSQLQueryState(t, sqlQuery, StateNotConnected)
}

func TestSqlQueryAllowQueriesFailStrictModeConflictWithRowCache(t *testing.T) {
	setUpSQLQueryTest()
	testUtils := newTestUtils()
	config := testUtils.newQueryServiceConfig()
	// disable strict mode
	config.StrictMode = false
	sqlQuery := NewSqlQuery(config)
	checkSQLQueryState(t, sqlQuery, StateNotConnected)
	dbconfigs := testUtils.newDBConfigs()
	// enable rowcache
	dbconfigs.App.EnableRowcache = true
	err := sqlQuery.StartService(nil, &dbconfigs, []SchemaOverride{}, testUtils.newMysqld(&dbconfigs))
	if err == nil {
		t.Fatalf("SqlQuery.StartService should fail because strict mode is disabled while rowcache is enabled.")
	}
	checkSQLQueryState(t, sqlQuery, StateNotConnected)
}

func TestSqlQueryAllowQueries(t *testing.T) {
	setUpSQLQueryTest()
	testUtils := newTestUtils()
	config := testUtils.newQueryServiceConfig()
	sqlQuery := NewSqlQuery(config)
	checkSQLQueryState(t, sqlQuery, StateNotConnected)
	dbconfigs := testUtils.newDBConfigs()
	sqlQuery.setState(StateServing)
	err := sqlQuery.StartService(nil, &dbconfigs, []SchemaOverride{}, testUtils.newMysqld(&dbconfigs))
	sqlQuery.StopService()
	want := "cannot start tabletserver"
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("SqlQuery.StartService: %v, must contain %s", err, want)
	}
	sqlQuery.setState(StateShuttingDown)
	err = sqlQuery.StartService(nil, &dbconfigs, []SchemaOverride{}, testUtils.newMysqld(&dbconfigs))
	if err == nil {
		t.Fatalf("SqlQuery.StartService should fail")
	}
	sqlQuery.StopService()
}

func TestSqlQueryInitDBConfig(t *testing.T) {
	setUpSQLQueryTest()
	testUtils := newTestUtils()
	config := testUtils.newQueryServiceConfig()
	sqlQuery := NewSqlQuery(config)
	sqlQuery.setState(StateServing)
	err := sqlQuery.InitDBConfig(nil, nil, nil, nil)
	want := "InitDBConfig failed"
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Errorf("SqlQuery.StartService: %v, must contain %s", err, want)
	}
	sqlQuery.setState(StateNotConnected)
	err = sqlQuery.InitDBConfig(nil, nil, nil, nil)
	if err != nil {
		t.Error(err)
	}
}

func TestDecideAction(t *testing.T) {
	setUpSQLQueryTest()
	testUtils := newTestUtils()
	config := testUtils.newQueryServiceConfig()
	sqlQuery := NewSqlQuery(config)
	err := sqlQuery.InitDBConfig(nil, nil, nil, nil)
	if err != nil {
		t.Error(err)
	}

	_, err = sqlQuery.decideAction(topodata.TabletType_MASTER, false)
	want := "cannot SetServingType"
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Errorf("SqlQuery.StartService: %v, must contain %s", err, want)
	}

	target := &pb.Target{}
	err = sqlQuery.InitDBConfig(target, nil, nil, nil)
	if err != nil {
		t.Error(err)
	}

	sqlQuery.setState(StateNotConnected)
	action, err := sqlQuery.decideAction(topodata.TabletType_MASTER, false)
	if err != nil {
		t.Error(err)
	}
	if action != actionNone {
		t.Errorf("decideAction: %v, want %v", action, actionNone)
	}

	sqlQuery.setState(StateNotConnected)
	action, err = sqlQuery.decideAction(topodata.TabletType_MASTER, true)
	if err != nil {
		t.Error(err)
	}
	if action != actionFullStart {
		t.Errorf("decideAction: %v, want %v", action, actionFullStart)
	}
	if sqlQuery.state != StateTransitioning {
		t.Errorf("sqlQuery.state: %v, want %v", sqlQuery.state, StateTransitioning)
	}

	sqlQuery.setState(StateNotServing)
	action, err = sqlQuery.decideAction(topodata.TabletType_MASTER, false)
	if err != nil {
		t.Error(err)
	}
	if action != actionNone {
		t.Errorf("decideAction: %v, want %v", action, actionNone)
	}

	sqlQuery.setState(StateNotServing)
	action, err = sqlQuery.decideAction(topodata.TabletType_MASTER, true)
	if err != nil {
		t.Error(err)
	}
	if action != actionServeNewType {
		t.Errorf("decideAction: %v, want %v", action, actionServeNewType)
	}
	if sqlQuery.state != StateTransitioning {
		t.Errorf("sqlQuery.state: %v, want %v", sqlQuery.state, StateTransitioning)
	}

	sqlQuery.setState(StateServing)
	action, err = sqlQuery.decideAction(topodata.TabletType_MASTER, false)
	if err != nil {
		t.Error(err)
	}
	if action != actionGracefulStop {
		t.Errorf("decideAction: %v, want %v", action, actionGracefulStop)
	}
	if sqlQuery.state != StateShuttingDown {
		t.Errorf("sqlQuery.state: %v, want %v", sqlQuery.state, StateShuttingDown)
	}

	sqlQuery.setState(StateServing)
	action, err = sqlQuery.decideAction(topodata.TabletType_MASTER, true)
	if err != nil {
		t.Error(err)
	}
	if action != actionServeNewType {
		t.Errorf("decideAction: %v, want %v", action, actionServeNewType)
	}
	if sqlQuery.state != StateTransitioning {
		t.Errorf("sqlQuery.state: %v, want %v", sqlQuery.state, StateTransitioning)
	}

	sqlQuery.setState(StateTransitioning)
	action, err = sqlQuery.decideAction(topodata.TabletType_MASTER, false)
	want = "cannot SetServingType"
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Errorf("SqlQuery.StartService: %v, must contain %s", err, want)
	}

	sqlQuery.setState(StateShuttingDown)
	action, err = sqlQuery.decideAction(topodata.TabletType_MASTER, false)
	want = "cannot SetServingType"
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Errorf("SqlQuery.StartService: %v, must contain %s", err, want)
	}
}

func TestSetServingType(t *testing.T) {
	setUpSQLQueryTest()
	testUtils := newTestUtils()
	config := testUtils.newQueryServiceConfig()
	sqlQuery := NewSqlQuery(config)
	dbconfigs := testUtils.newDBConfigs()

	err := sqlQuery.InitDBConfig(nil, &dbconfigs, []SchemaOverride{}, testUtils.newMysqld(&dbconfigs))
	if err != nil {
		t.Error(err)
	}

	err = sqlQuery.SetServingType(topodata.TabletType_REPLICA, true)
	want := "cannot SetServingType"
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Errorf("SqlQuery.StartService: %v, must contain %s", err, want)
	}

	target := &pb.Target{}
	err = sqlQuery.InitDBConfig(target, &dbconfigs, []SchemaOverride{}, testUtils.newMysqld(&dbconfigs))
	if err != nil {
		t.Error(err)
	}

	err = sqlQuery.SetServingType(topodata.TabletType_REPLICA, false)
	if err != nil {
		t.Error(err)
	}
	checkSQLQueryState(t, sqlQuery, StateNotConnected)

	err = sqlQuery.SetServingType(topodata.TabletType_REPLICA, true)
	if err != nil {
		t.Error(err)
	}
	checkSQLQueryState(t, sqlQuery, StateServing)

	err = sqlQuery.SetServingType(topodata.TabletType_RDONLY, true)
	if err != nil {
		t.Error(err)
	}
	checkSQLQueryState(t, sqlQuery, StateServing)

	err = sqlQuery.SetServingType(topodata.TabletType_SPARE, false)
	if err != nil {
		t.Error(err)
	}
	checkSQLQueryState(t, sqlQuery, StateNotServing)

	sqlQuery.StopService()
	checkSQLQueryState(t, sqlQuery, StateNotConnected)
}

func TestSqlQueryCheckMysql(t *testing.T) {
	setUpSQLQueryTest()
	testUtils := newTestUtils()
	config := testUtils.newQueryServiceConfig()
	sqlQuery := NewSqlQuery(config)
	dbconfigs := testUtils.newDBConfigs()
	target := &pb.Target{}
	err := sqlQuery.StartService(target, &dbconfigs, []SchemaOverride{}, testUtils.newMysqld(&dbconfigs))
	defer sqlQuery.StopService()
	if err != nil {
		t.Fatal(err)
	}
	if !sqlQuery.CheckMySQL() {
		t.Error("CheckMySQL should return true")
	}
	err = sqlQuery.SetServingType(topodata.TabletType_SPARE, false)
	if err != nil {
		t.Fatal(err)
	}
	if !sqlQuery.CheckMySQL() {
		t.Error("CheckMySQL should return true")
	}
	checkSQLQueryState(t, sqlQuery, StateNotServing)
}

func TestSqlQueryCheckMysqlFailInvalidConn(t *testing.T) {
	db := setUpSQLQueryTest()
	testUtils := newTestUtils()
	config := testUtils.newQueryServiceConfig()
	sqlQuery := NewSqlQuery(config)
	dbconfigs := testUtils.newDBConfigs()
	err := sqlQuery.StartService(nil, &dbconfigs, []SchemaOverride{}, testUtils.newMysqld(&dbconfigs))
	defer sqlQuery.StopService()
	if err != nil {
		t.Fatalf("SqlQuery.StartService should success but get error: %v", err)
	}
	// make mysql conn fail
	db.EnableConnFail()
	if sqlQuery.CheckMySQL() {
		t.Fatalf("CheckMySQL should return false")
	}
}

func TestSqlQueryCheckMysqlFailUninitializedQueryEngine(t *testing.T) {
	setUpSQLQueryTest()
	testUtils := newTestUtils()
	config := testUtils.newQueryServiceConfig()
	sqlQuery := NewSqlQuery(config)
	dbconfigs := testUtils.newDBConfigs()
	// this causes QueryEngine not being initialized properly
	sqlQuery.setState(StateServing)
	err := sqlQuery.StartService(nil, &dbconfigs, []SchemaOverride{}, testUtils.newMysqld(&dbconfigs))
	defer sqlQuery.StopService()
	want := "cannot start tabletserver"
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("SqlQuery.StartService: %v, must contain %s", err, want)
	}
	// QueryEngine.CheckMySQL shoudl panic and CheckMySQL should return false
	if sqlQuery.CheckMySQL() {
		t.Fatalf("CheckMySQL should return false")
	}
}

func TestSqlQueryCheckMysqlInUnintialized(t *testing.T) {
	setUpSQLQueryTest()
	testUtils := newTestUtils()
	config := testUtils.newQueryServiceConfig()
	config.EnablePublishStats = true
	sqlQuery := NewSqlQuery(config)
	// sqlquery start request fail because we are in StateNotConnected;
	// however, CheckMySQL should return true. Here, we always assume
	// MySQL is healthy unless we've verified it is not.
	if !sqlQuery.CheckMySQL() {
		t.Fatalf("CheckMySQL should return true")
	}
	tabletState := expvar.Get(config.StatsPrefix + "TabletState")
	if tabletState == nil {
		t.Fatalf("%sTabletState should be exposed", config.StatsPrefix)
	}
	varzState, err := strconv.Atoi(tabletState.String())
	if err != nil {
		t.Fatalf("invalid state reported by expvar, should be a valid state code, but got: %s", tabletState.String())
	}
	if varzState != StateNotConnected {
		t.Fatalf("queryservice should be in %d state, but exposed varz reports: %s", StateNotConnected, varzState)
	}
}

func TestSqlQueryGetSessionId(t *testing.T) {
	setUpSQLQueryTest()
	testUtils := newTestUtils()
	config := testUtils.newQueryServiceConfig()
	sqlQuery := NewSqlQuery(config)
	if err := sqlQuery.GetSessionId(nil, nil); err == nil {
		t.Fatalf("call GetSessionId should get an error")
	}
	keyspace := "test_keyspace"
	shard := "0"
	dbconfigs := testUtils.newDBConfigs()
	err := sqlQuery.StartService(nil, &dbconfigs, []SchemaOverride{}, testUtils.newMysqld(&dbconfigs))
	if err != nil {
		t.Fatalf("StartService failed: %v", err)
	}
	defer sqlQuery.StopService()
	sessionInfo := proto.SessionInfo{}
	err = sqlQuery.GetSessionId(
		&proto.SessionParams{Keyspace: keyspace, Shard: shard},
		&sessionInfo,
	)
	if err != nil {
		t.Fatalf("got GetSessionId error: %v", err)
	}
	if sessionInfo.SessionId != sqlQuery.sessionID {
		t.Fatalf("call GetSessionId returns an unexpected session id, "+
			"expect seesion id: %d but got %d", sqlQuery.sessionID,
			sessionInfo.SessionId)
	}
	err = sqlQuery.GetSessionId(
		&proto.SessionParams{Keyspace: keyspace},
		&sessionInfo,
	)
	if err == nil {
		t.Fatalf("call GetSessionId should fail because of missing shard in request")
	}
	err = sqlQuery.GetSessionId(
		&proto.SessionParams{Shard: shard},
		&sessionInfo,
	)
	if err == nil {
		t.Fatalf("call GetSessionId should fail because of missing keyspace in request")
	}
}

func TestSqlQueryCommandFailUnMatchedSessionId(t *testing.T) {
	setUpSQLQueryTest()
	testUtils := newTestUtils()
	config := testUtils.newQueryServiceConfig()
	sqlQuery := NewSqlQuery(config)
	dbconfigs := testUtils.newDBConfigs()
	err := sqlQuery.StartService(nil, &dbconfigs, []SchemaOverride{}, testUtils.newMysqld(&dbconfigs))
	if err != nil {
		t.Fatalf("StartService failed: %v", err)
	}
	defer sqlQuery.StopService()
	ctx := context.Background()
	session := proto.Session{
		SessionId:     0,
		TransactionId: 0,
	}
	txInfo := proto.TransactionInfo{TransactionId: 0}
	if err = sqlQuery.Begin(ctx, nil, &session, &txInfo); err == nil {
		t.Fatalf("call SqlQuery.Begin should fail because of an invalid session id: 0")
	}

	if err = sqlQuery.Commit(ctx, nil, &session); err == nil {
		t.Fatalf("call SqlQuery.Commit should fail because of an invalid session id: 0")
	}

	if err = sqlQuery.Rollback(ctx, nil, &session); err == nil {
		t.Fatalf("call SqlQuery.Rollback should fail because of an invalid session id: 0")
	}

	query := proto.Query{
		Sql:           "select * from test_table limit 1000",
		BindVariables: nil,
		SessionId:     session.SessionId,
		TransactionId: session.TransactionId,
	}
	reply := mproto.QueryResult{}
	if err := sqlQuery.Execute(ctx, nil, &query, &reply); err == nil {
		t.Fatalf("call SqlQuery.Execute should fail because of an invalid session id: 0")
	}

	streamSendReply := func(*mproto.QueryResult) error { return nil }
	if err = sqlQuery.StreamExecute(ctx, nil, &query, streamSendReply); err == nil {
		t.Fatalf("call SqlQuery.StreamExecute should fail because of an invalid session id: 0")
	}

	batchQuery := proto.QueryList{
		Queries: []proto.BoundQuery{
			proto.BoundQuery{
				Sql:           "noquery",
				BindVariables: nil,
			},
		},
		SessionId: session.SessionId,
	}

	batchReply := proto.QueryResultList{
		List: []mproto.QueryResult{
			mproto.QueryResult{},
			mproto.QueryResult{},
		},
	}
	if err = sqlQuery.ExecuteBatch(ctx, nil, &batchQuery, &batchReply); err == nil {
		t.Fatalf("call SqlQuery.ExecuteBatch should fail because of an invalid session id: 0")
	}

	splitQuery := proto.SplitQueryRequest{
		Query: proto.BoundQuery{
			Sql:           "select * from test_table where count > :count",
			BindVariables: nil,
		},
		SplitCount: 10,
		SessionID:  session.SessionId,
	}

	splitQueryReply := proto.SplitQueryResult{
		Queries: []proto.QuerySplit{
			proto.QuerySplit{
				Query: proto.BoundQuery{
					Sql:           "",
					BindVariables: nil,
				},
				RowCount: 10,
			},
		},
	}
	if err = sqlQuery.SplitQuery(ctx, nil, &splitQuery, &splitQueryReply); err == nil {
		t.Fatalf("call SqlQuery.SplitQuery should fail because of an invalid session id: 0")
	}
}

func TestSqlQueryCommitTransaciton(t *testing.T) {
	db := setUpSQLQueryTest()
	testUtils := newTestUtils()
	// sql that will be executed in this test
	executeSQL := "select * from test_table limit 1000"
	executeSQLResult := &mproto.QueryResult{
		RowsAffected: 1,
		Rows: [][]sqltypes.Value{
			[]sqltypes.Value{sqltypes.MakeString([]byte("row01"))},
		},
	}
	db.AddQuery(executeSQL, executeSQLResult)
	config := testUtils.newQueryServiceConfig()
	sqlQuery := NewSqlQuery(config)
	dbconfigs := testUtils.newDBConfigs()
	err := sqlQuery.StartService(nil, &dbconfigs, []SchemaOverride{}, testUtils.newMysqld(&dbconfigs))
	if err != nil {
		t.Fatalf("StartService failed: %v", err)
	}
	defer sqlQuery.StopService()
	ctx := context.Background()
	session := proto.Session{
		SessionId:     sqlQuery.sessionID,
		TransactionId: 0,
	}
	txInfo := proto.TransactionInfo{TransactionId: 0}
	if err = sqlQuery.Begin(ctx, nil, &session, &txInfo); err != nil {
		t.Fatalf("call SqlQuery.Begin failed")
	}
	session.TransactionId = txInfo.TransactionId
	query := proto.Query{
		Sql:           executeSQL,
		BindVariables: nil,
		SessionId:     session.SessionId,
		TransactionId: session.TransactionId,
	}
	reply := mproto.QueryResult{}
	if err := sqlQuery.Execute(ctx, nil, &query, &reply); err != nil {
		t.Fatalf("failed to execute query: %s", query.Sql)
	}
	if err := sqlQuery.Commit(ctx, nil, &session); err != nil {
		t.Fatalf("call SqlQuery.Commit failed")
	}
}

func TestSqlQueryRollback(t *testing.T) {
	db := setUpSQLQueryTest()
	testUtils := newTestUtils()
	// sql that will be executed in this test
	executeSQL := "select * from test_table limit 1000"
	executeSQLResult := &mproto.QueryResult{
		RowsAffected: 1,
		Rows: [][]sqltypes.Value{
			[]sqltypes.Value{sqltypes.MakeString([]byte("row01"))},
		},
	}
	db.AddQuery(executeSQL, executeSQLResult)
	config := testUtils.newQueryServiceConfig()
	sqlQuery := NewSqlQuery(config)
	dbconfigs := testUtils.newDBConfigs()
	err := sqlQuery.StartService(nil, &dbconfigs, []SchemaOverride{}, testUtils.newMysqld(&dbconfigs))
	if err != nil {
		t.Fatalf("StartService failed: %v", err)
	}
	defer sqlQuery.StopService()
	ctx := context.Background()
	session := proto.Session{
		SessionId:     sqlQuery.sessionID,
		TransactionId: 0,
	}
	txInfo := proto.TransactionInfo{TransactionId: 0}
	if err = sqlQuery.Begin(ctx, nil, &session, &txInfo); err != nil {
		t.Fatalf("call SqlQuery.Begin failed")
	}
	session.TransactionId = txInfo.TransactionId
	query := proto.Query{
		Sql:           executeSQL,
		BindVariables: nil,
		SessionId:     session.SessionId,
		TransactionId: session.TransactionId,
	}
	reply := mproto.QueryResult{}
	if err := sqlQuery.Execute(ctx, nil, &query, &reply); err != nil {
		t.Fatalf("failed to execute query: %s", query.Sql)
	}
	if err := sqlQuery.Rollback(ctx, nil, &session); err != nil {
		t.Fatalf("call SqlQuery.Rollback failed")
	}
}

func TestSqlQueryStreamExecute(t *testing.T) {
	db := setUpSQLQueryTest()
	testUtils := newTestUtils()
	// sql that will be executed in this test
	executeSQL := "select * from test_table limit 1000"
	executeSQLResult := &mproto.QueryResult{
		RowsAffected: 1,
		Rows: [][]sqltypes.Value{
			[]sqltypes.Value{sqltypes.MakeString([]byte("row01"))},
		},
	}
	db.AddQuery(executeSQL, executeSQLResult)

	config := testUtils.newQueryServiceConfig()
	sqlQuery := NewSqlQuery(config)
	dbconfigs := testUtils.newDBConfigs()
	err := sqlQuery.StartService(nil, &dbconfigs, []SchemaOverride{}, testUtils.newMysqld(&dbconfigs))
	if err != nil {
		t.Fatalf("StartService failed: %v", err)
	}
	defer sqlQuery.StopService()
	ctx := context.Background()
	session := proto.Session{
		SessionId:     sqlQuery.sessionID,
		TransactionId: 0,
	}
	txInfo := proto.TransactionInfo{TransactionId: 0}
	if err = sqlQuery.Begin(ctx, nil, &session, &txInfo); err != nil {
		t.Fatalf("call SqlQuery.Begin failed")
	}
	session.TransactionId = txInfo.TransactionId
	query := proto.Query{
		Sql:           executeSQL,
		BindVariables: nil,
		SessionId:     session.SessionId,
		TransactionId: session.TransactionId,
	}
	sendReply := func(*mproto.QueryResult) error { return nil }
	if err := sqlQuery.StreamExecute(ctx, nil, &query, sendReply); err == nil {
		t.Fatalf("SqlQuery.StreamExecute should fail: %s", query.Sql)
	}
	if err := sqlQuery.Rollback(ctx, nil, &session); err != nil {
		t.Fatalf("call SqlQuery.Rollback failed")
	}
	query.TransactionId = 0
	if err := sqlQuery.StreamExecute(ctx, nil, &query, sendReply); err != nil {
		t.Fatalf("SqlQuery.StreamExecute should success: %s, but get error: %v",
			query.Sql, err)
	}
}

func TestSqlQueryExecuteBatch(t *testing.T) {
	db := setUpSQLQueryTest()
	testUtils := newTestUtils()
	sql := "insert into test_table values (1, 2)"
	sqlResult := &mproto.QueryResult{}
	expanedSQL := "insert into test_table values (1, 2) /* _stream test_table (pk ) (1 ); */"

	db.AddQuery(sql, sqlResult)
	db.AddQuery(expanedSQL, sqlResult)
	config := testUtils.newQueryServiceConfig()
	sqlQuery := NewSqlQuery(config)
	dbconfigs := testUtils.newDBConfigs()
	err := sqlQuery.StartService(nil, &dbconfigs, []SchemaOverride{}, testUtils.newMysqld(&dbconfigs))
	if err != nil {
		t.Fatalf("StartService failed: %v", err)
	}
	defer sqlQuery.StopService()
	ctx := context.Background()
	query := proto.QueryList{
		Queries: []proto.BoundQuery{
			proto.BoundQuery{
				Sql:           sql,
				BindVariables: nil,
			},
		},
		AsTransaction: true,
		SessionId:     sqlQuery.sessionID,
	}

	reply := proto.QueryResultList{
		List: []mproto.QueryResult{
			mproto.QueryResult{},
			*sqlResult,
			mproto.QueryResult{},
		},
	}
	if err := sqlQuery.ExecuteBatch(ctx, nil, &query, &reply); err != nil {
		t.Fatalf("SqlQuery.ExecuteBatch should success: %v, but get error: %v",
			query, err)
	}
}

func TestSqlQueryExecuteBatchFailEmptyQueryList(t *testing.T) {
	setUpSQLQueryTest()
	testUtils := newTestUtils()
	config := testUtils.newQueryServiceConfig()
	sqlQuery := NewSqlQuery(config)
	dbconfigs := testUtils.newDBConfigs()
	err := sqlQuery.StartService(nil, &dbconfigs, []SchemaOverride{}, testUtils.newMysqld(&dbconfigs))
	if err != nil {
		t.Fatalf("StartService failed: %v", err)
	}
	defer sqlQuery.StopService()
	ctx := context.Background()
	query := proto.QueryList{
		Queries:   []proto.BoundQuery{},
		SessionId: sqlQuery.sessionID,
	}

	reply := proto.QueryResultList{
		List: []mproto.QueryResult{},
	}
	err = sqlQuery.ExecuteBatch(ctx, nil, &query, &reply)
	verifyTabletError(t, err, ErrFail)
}

func TestSqlQueryExecuteBatchFailAsTransaction(t *testing.T) {
	setUpSQLQueryTest()
	testUtils := newTestUtils()
	config := testUtils.newQueryServiceConfig()
	sqlQuery := NewSqlQuery(config)
	dbconfigs := testUtils.newDBConfigs()
	err := sqlQuery.StartService(nil, &dbconfigs, []SchemaOverride{}, testUtils.newMysqld(&dbconfigs))
	if err != nil {
		t.Fatalf("StartService failed: %v", err)
	}
	defer sqlQuery.StopService()
	ctx := context.Background()
	query := proto.QueryList{
		Queries: []proto.BoundQuery{
			proto.BoundQuery{
				Sql:           "begin",
				BindVariables: nil,
			},
		},
		SessionId:     sqlQuery.sessionID,
		AsTransaction: true,
		TransactionId: 1,
	}

	reply := proto.QueryResultList{
		List: []mproto.QueryResult{},
	}
	err = sqlQuery.ExecuteBatch(ctx, nil, &query, &reply)
	verifyTabletError(t, err, ErrFail)
}

func TestSqlQueryExecuteBatchBeginFail(t *testing.T) {
	db := setUpSQLQueryTest()
	testUtils := newTestUtils()
	// make "begin" query fail
	db.AddRejectedQuery("begin", errRejected)
	config := testUtils.newQueryServiceConfig()
	sqlQuery := NewSqlQuery(config)
	dbconfigs := testUtils.newDBConfigs()
	err := sqlQuery.StartService(nil, &dbconfigs, []SchemaOverride{}, testUtils.newMysqld(&dbconfigs))
	if err != nil {
		t.Fatalf("StartService failed: %v", err)
	}
	defer sqlQuery.StopService()
	ctx := context.Background()
	query := proto.QueryList{
		Queries: []proto.BoundQuery{
			proto.BoundQuery{
				Sql:           "begin",
				BindVariables: nil,
			},
		},
		SessionId: sqlQuery.sessionID,
	}

	reply := proto.QueryResultList{
		List: []mproto.QueryResult{
			mproto.QueryResult{},
		},
	}
	if err := sqlQuery.ExecuteBatch(ctx, nil, &query, &reply); err == nil {
		t.Fatalf("SqlQuery.ExecuteBatch should fail")
	}
}

func TestSqlQueryExecuteBatchCommitFail(t *testing.T) {
	db := setUpSQLQueryTest()
	testUtils := newTestUtils()
	// make "commit" query fail
	db.AddRejectedQuery("commit", errRejected)
	config := testUtils.newQueryServiceConfig()
	sqlQuery := NewSqlQuery(config)
	dbconfigs := testUtils.newDBConfigs()
	err := sqlQuery.StartService(nil, &dbconfigs, []SchemaOverride{}, testUtils.newMysqld(&dbconfigs))
	if err != nil {
		t.Fatalf("StartService failed: %v", err)
	}
	defer sqlQuery.StopService()
	ctx := context.Background()
	query := proto.QueryList{
		Queries: []proto.BoundQuery{
			proto.BoundQuery{
				Sql:           "begin",
				BindVariables: nil,
			},
			proto.BoundQuery{
				Sql:           "commit",
				BindVariables: nil,
			},
		},
		SessionId: sqlQuery.sessionID,
	}

	reply := proto.QueryResultList{
		List: []mproto.QueryResult{
			mproto.QueryResult{},
			mproto.QueryResult{},
		},
	}
	if err := sqlQuery.ExecuteBatch(ctx, nil, &query, &reply); err == nil {
		t.Fatalf("SqlQuery.ExecuteBatch should fail")
	}
}

func TestSqlQueryExecuteBatchSqlExecFailInTransaction(t *testing.T) {
	db := setUpSQLQueryTest()
	testUtils := newTestUtils()
	sql := "insert into test_table values (1, 2)"
	sqlResult := &mproto.QueryResult{}
	expanedSQL := "insert into test_table values (1, 2) /* _stream test_table (pk ) (1 ); */"

	db.AddQuery(sql, sqlResult)
	db.AddQuery(expanedSQL, sqlResult)

	// make this query fail
	db.AddRejectedQuery(sql, errRejected)
	db.AddRejectedQuery(expanedSQL, errRejected)

	config := testUtils.newQueryServiceConfig()
	sqlQuery := NewSqlQuery(config)
	dbconfigs := testUtils.newDBConfigs()
	err := sqlQuery.StartService(nil, &dbconfigs, []SchemaOverride{}, testUtils.newMysqld(&dbconfigs))
	if err != nil {
		t.Fatalf("StartService failed: %v", err)
	}
	defer sqlQuery.StopService()
	ctx := context.Background()
	query := proto.QueryList{
		Queries: []proto.BoundQuery{
			proto.BoundQuery{
				Sql:           sql,
				BindVariables: nil,
			},
		},
		AsTransaction: true,
		SessionId:     sqlQuery.sessionID,
	}

	reply := proto.QueryResultList{
		List: []mproto.QueryResult{
			mproto.QueryResult{},
			*sqlResult,
			mproto.QueryResult{},
		},
	}

	if db.GetQueryCalledNum("rollback") != 0 {
		t.Fatalf("rollback should not be executed.")
	}

	if err := sqlQuery.ExecuteBatch(ctx, nil, &query, &reply); err == nil {
		t.Fatalf("SqlQuery.ExecuteBatch should fail")
	}

	if db.GetQueryCalledNum("rollback") != 1 {
		t.Fatalf("rollback should be executed only once.")
	}
}

func TestSqlQueryExecuteBatchSqlSucceedInTransaction(t *testing.T) {
	db := setUpSQLQueryTest()
	testUtils := newTestUtils()
	sql := "insert into test_table values (1, 2)"
	sqlResult := &mproto.QueryResult{}
	expanedSQL := "insert into test_table values (1, 2) /* _stream test_table (pk ) (1 ); */"

	db.AddQuery(sql, sqlResult)
	db.AddQuery(expanedSQL, sqlResult)

	// cause execution error for this particular sql query
	db.AddRejectedQuery(sql, errRejected)

	config := testUtils.newQueryServiceConfig()
	config.EnableAutoCommit = true
	sqlQuery := NewSqlQuery(config)
	dbconfigs := testUtils.newDBConfigs()
	err := sqlQuery.StartService(nil, &dbconfigs, []SchemaOverride{}, testUtils.newMysqld(&dbconfigs))
	if err != nil {
		t.Fatalf("StartService failed: %v", err)
	}
	defer sqlQuery.StopService()
	ctx := context.Background()
	query := proto.QueryList{
		Queries: []proto.BoundQuery{
			proto.BoundQuery{
				Sql:           sql,
				BindVariables: nil,
			},
		},
		SessionId: sqlQuery.sessionID,
	}

	reply := proto.QueryResultList{
		List: []mproto.QueryResult{
			*sqlResult,
		},
	}
	if err := sqlQuery.ExecuteBatch(ctx, nil, &query, &reply); err != nil {
		t.Fatalf("SqlQuery.ExecuteBatch should succeed")
	}
}

func TestSqlQueryExecuteBatchCallCommitWithoutABegin(t *testing.T) {
	setUpSQLQueryTest()
	testUtils := newTestUtils()
	config := testUtils.newQueryServiceConfig()
	sqlQuery := NewSqlQuery(config)
	dbconfigs := testUtils.newDBConfigs()
	err := sqlQuery.StartService(nil, &dbconfigs, []SchemaOverride{}, testUtils.newMysqld(&dbconfigs))
	if err != nil {
		t.Fatalf("StartService failed: %v", err)
	}
	defer sqlQuery.StopService()
	ctx := context.Background()
	query := proto.QueryList{
		Queries: []proto.BoundQuery{
			proto.BoundQuery{
				Sql:           "commit",
				BindVariables: nil,
			},
		},
		SessionId: sqlQuery.sessionID,
	}

	reply := proto.QueryResultList{
		List: []mproto.QueryResult{
			mproto.QueryResult{},
		},
	}
	if err := sqlQuery.ExecuteBatch(ctx, nil, &query, &reply); err == nil {
		t.Fatalf("SqlQuery.ExecuteBatch should fail")
	}
}

func TestExecuteBatchNestedTransaction(t *testing.T) {
	db := setUpSQLQueryTest()
	testUtils := newTestUtils()
	sql := "insert into test_table values (1, 2)"
	sqlResult := &mproto.QueryResult{}
	expanedSQL := "insert into test_table values (1, 2) /* _stream test_table (pk ) (1 ); */"

	db.AddQuery(sql, sqlResult)
	db.AddQuery(expanedSQL, sqlResult)
	config := testUtils.newQueryServiceConfig()
	sqlQuery := NewSqlQuery(config)
	dbconfigs := testUtils.newDBConfigs()
	err := sqlQuery.StartService(nil, &dbconfigs, []SchemaOverride{}, testUtils.newMysqld(&dbconfigs))
	if err != nil {
		t.Fatalf("StartService failed: %v", err)
	}
	defer sqlQuery.StopService()
	ctx := context.Background()
	query := proto.QueryList{
		Queries: []proto.BoundQuery{
			proto.BoundQuery{
				Sql:           "begin",
				BindVariables: nil,
			},
			proto.BoundQuery{
				Sql:           "begin",
				BindVariables: nil,
			},
			proto.BoundQuery{
				Sql:           sql,
				BindVariables: nil,
			},
			proto.BoundQuery{
				Sql:           "commit",
				BindVariables: nil,
			},
			proto.BoundQuery{
				Sql:           "commit",
				BindVariables: nil,
			},
		},
		SessionId: sqlQuery.sessionID,
	}

	reply := proto.QueryResultList{
		List: []mproto.QueryResult{
			mproto.QueryResult{},
			mproto.QueryResult{},
			*sqlResult,
			mproto.QueryResult{},
			mproto.QueryResult{},
		},
	}
	if err := sqlQuery.ExecuteBatch(ctx, nil, &query, &reply); err == nil {
		t.Fatalf("SqlQuery.Execute should fail because of nested transaction")
	}
	sqlQuery.qe.txPool.SetTimeout(10)
}

func TestSqlQuerySplitQuery(t *testing.T) {
	db := setUpSQLQueryTest()
	db.AddQuery("SELECT MIN(pk), MAX(pk) FROM test_table", &mproto.QueryResult{
		Fields: []mproto.Field{
			mproto.Field{Name: "pk", Type: mproto.VT_LONG},
		},
		RowsAffected: 1,
		Rows: [][]sqltypes.Value{
			[]sqltypes.Value{
				sqltypes.MakeNumeric([]byte("1")),
				sqltypes.MakeNumeric([]byte("100")),
			},
		},
	})
	db.AddQuery("SELECT pk FROM test_table LIMIT 0", &mproto.QueryResult{
		Fields: []mproto.Field{
			mproto.Field{Name: "pk", Type: mproto.VT_LONG},
		},
		RowsAffected: 1,
		Rows: [][]sqltypes.Value{
			[]sqltypes.Value{
				sqltypes.MakeNumeric([]byte("1")),
			},
		},
	})
	testUtils := newTestUtils()
	config := testUtils.newQueryServiceConfig()
	sqlQuery := NewSqlQuery(config)
	dbconfigs := testUtils.newDBConfigs()
	err := sqlQuery.StartService(nil, &dbconfigs, []SchemaOverride{}, testUtils.newMysqld(&dbconfigs))
	if err != nil {
		t.Fatalf("StartService failed: %v", err)
	}
	defer sqlQuery.StopService()
	ctx := context.Background()
	query := proto.SplitQueryRequest{
		Query: proto.BoundQuery{
			Sql:           "select * from test_table where count > :count",
			BindVariables: nil,
		},
		SplitCount: 10,
		SessionID:  sqlQuery.sessionID,
	}

	reply := proto.SplitQueryResult{
		Queries: []proto.QuerySplit{
			proto.QuerySplit{
				Query: proto.BoundQuery{
					Sql:           "",
					BindVariables: nil,
				},
				RowCount: 10,
			},
		},
	}
	if err := sqlQuery.SplitQuery(ctx, nil, &query, &reply); err != nil {
		t.Fatalf("SqlQuery.SplitQuery should success: %v, but get error: %v",
			query, err)
	}
}

func TestSqlQuerySplitQueryInvalidQuery(t *testing.T) {
	db := setUpSQLQueryTest()
	db.AddQuery("SELECT MIN(pk), MAX(pk) FROM test_table", &mproto.QueryResult{
		Fields: []mproto.Field{
			mproto.Field{Name: "pk", Type: mproto.VT_LONG},
		},
		RowsAffected: 1,
		Rows: [][]sqltypes.Value{
			[]sqltypes.Value{
				sqltypes.MakeNumeric([]byte("1")),
				sqltypes.MakeNumeric([]byte("100")),
			},
		},
	})
	db.AddQuery("SELECT pk FROM test_table LIMIT 0", &mproto.QueryResult{
		Fields: []mproto.Field{
			mproto.Field{Name: "pk", Type: mproto.VT_LONG},
		},
		RowsAffected: 1,
		Rows: [][]sqltypes.Value{
			[]sqltypes.Value{
				sqltypes.MakeNumeric([]byte("1")),
			},
		},
	})
	testUtils := newTestUtils()
	config := testUtils.newQueryServiceConfig()
	sqlQuery := NewSqlQuery(config)
	dbconfigs := testUtils.newDBConfigs()
	err := sqlQuery.StartService(nil, &dbconfigs, []SchemaOverride{}, testUtils.newMysqld(&dbconfigs))
	if err != nil {
		t.Fatalf("StartService failed: %v", err)
	}
	defer sqlQuery.StopService()
	ctx := context.Background()
	query := proto.SplitQueryRequest{
		Query: proto.BoundQuery{
			// add limit clause to make SplitQuery fail
			Sql:           "select * from test_table where count > :count limit 1000",
			BindVariables: nil,
		},
		SplitCount: 10,
		SessionID:  sqlQuery.sessionID,
	}

	reply := proto.SplitQueryResult{
		Queries: []proto.QuerySplit{
			proto.QuerySplit{
				Query: proto.BoundQuery{
					Sql:           "",
					BindVariables: nil,
				},
				RowCount: 10,
			},
		},
	}
	if err := sqlQuery.SplitQuery(ctx, nil, &query, &reply); err == nil {
		t.Fatalf("SqlQuery.SplitQuery should fail")
	}
}

func TestSqlQuerySplitQueryInvalidMinMax(t *testing.T) {
	db := setUpSQLQueryTest()
	testUtils := newTestUtils()
	pkMinMaxQuery := "SELECT MIN(pk), MAX(pk) FROM test_table"
	pkMinMaxQueryResp := &mproto.QueryResult{
		Fields: []mproto.Field{
			mproto.Field{Name: "pk", Type: mproto.VT_LONG},
		},
		RowsAffected: 1,
		Rows: [][]sqltypes.Value{
			// this make SplitQueryFail
			[]sqltypes.Value{
				sqltypes.MakeString([]byte("invalid")),
				sqltypes.MakeString([]byte("invalid")),
			},
		},
	}
	db.AddQuery("SELECT pk FROM test_table LIMIT 0", &mproto.QueryResult{
		Fields: []mproto.Field{
			mproto.Field{Name: "pk", Type: mproto.VT_LONG},
		},
		RowsAffected: 1,
		Rows: [][]sqltypes.Value{
			[]sqltypes.Value{
				sqltypes.MakeNumeric([]byte("1")),
			},
		},
	})
	db.AddQuery(pkMinMaxQuery, pkMinMaxQueryResp)

	config := testUtils.newQueryServiceConfig()
	sqlQuery := NewSqlQuery(config)
	dbconfigs := testUtils.newDBConfigs()
	err := sqlQuery.StartService(nil, &dbconfigs, []SchemaOverride{}, testUtils.newMysqld(&dbconfigs))
	if err != nil {
		t.Fatalf("StartService failed: %v", err)
	}
	defer sqlQuery.StopService()
	ctx := context.Background()
	query := proto.SplitQueryRequest{
		Query: proto.BoundQuery{
			Sql:           "select * from test_table where count > :count",
			BindVariables: nil,
		},
		SplitCount: 10,
		SessionID:  sqlQuery.sessionID,
	}

	reply := proto.SplitQueryResult{
		Queries: []proto.QuerySplit{
			proto.QuerySplit{
				Query: proto.BoundQuery{
					Sql:           "",
					BindVariables: nil,
				},
				RowCount: 10,
			},
		},
	}
	if err := sqlQuery.SplitQuery(ctx, nil, &query, &reply); err == nil {
		t.Fatalf("SqlQuery.SplitQuery should fail")
	}
}

func TestHandleExecUnknownError(t *testing.T) {
	ctx := context.Background()
	logStats := newSqlQueryStats("TestHandleExecError", ctx)
	query := proto.Query{
		Sql:           "select * from test_table",
		BindVariables: nil,
	}
	var err error
	testUtils := newTestUtils()
	config := testUtils.newQueryServiceConfig()
	sqlQuery := NewSqlQuery(config)
	defer sqlQuery.handleExecError(&query, &err, logStats)
	panic("unknown exec error")
}

func TestHandleExecTabletError(t *testing.T) {
	ctx := context.Background()
	logStats := newSqlQueryStats("TestHandleExecError", ctx)
	query := proto.Query{
		Sql:           "select * from test_table",
		BindVariables: nil,
	}
	var err error
	defer func() {
		want := "fatal: tablet error"
		if err == nil || err.Error() != want {
			t.Errorf("Error: %v, want '%s'", err, want)
		}
	}()
	testUtils := newTestUtils()
	config := testUtils.newQueryServiceConfig()
	sqlQuery := NewSqlQuery(config)
	defer sqlQuery.handleExecError(&query, &err, logStats)
	panic(NewTabletError(ErrFatal, vtrpc.ErrorCode_UNKNOWN_ERROR, "tablet error"))
}

func TestTerseErrors1(t *testing.T) {
	ctx := context.Background()
	logStats := newSqlQueryStats("TestHandleExecError", ctx)
	query := proto.Query{
		Sql:           "select * from test_table",
		BindVariables: nil,
	}
	var err error
	defer func() {
		want := "fatal: tablet error"
		if err == nil || err.Error() != want {
			t.Errorf("Error: %v, want '%s'", err, want)
		}
	}()
	testUtils := newTestUtils()
	config := testUtils.newQueryServiceConfig()
	sqlQuery := NewSqlQuery(config)
	sqlQuery.config.TerseErrors = true
	defer sqlQuery.handleExecError(&query, &err, logStats)
	panic(NewTabletError(ErrFatal, vtrpc.ErrorCode_UNKNOWN_ERROR, "tablet error"))
}

func TestTerseErrors2(t *testing.T) {
	ctx := context.Background()
	logStats := newSqlQueryStats("TestHandleExecError", ctx)
	query := proto.Query{
		Sql:           "select * from test_table",
		BindVariables: map[string]interface{}{"a": 1},
	}
	var err error
	defer func() {
		want := "error: (errno 10) during query: select * from test_table"
		if err == nil || err.Error() != want {
			t.Errorf("Error: %v, want '%s'", err, want)
		}
	}()
	testUtils := newTestUtils()
	config := testUtils.newQueryServiceConfig()
	sqlQuery := NewSqlQuery(config)
	sqlQuery.config.TerseErrors = true
	defer sqlQuery.handleExecError(&query, &err, logStats)
	panic(&TabletError{
		ErrorType: ErrFail,
		Message:   "msg",
		SqlError:  10,
	})
}

func TestTerseErrors3(t *testing.T) {
	ctx := context.Background()
	logStats := newSqlQueryStats("TestHandleExecError", ctx)
	query := proto.Query{
		Sql:           "select * from test_table",
		BindVariables: nil,
	}
	var err error
	defer func() {
		want := "error: msg"
		if err == nil || err.Error() != want {
			t.Errorf("Error: %v, want '%s'", err, want)
		}
	}()
	testUtils := newTestUtils()
	config := testUtils.newQueryServiceConfig()
	sqlQuery := NewSqlQuery(config)
	sqlQuery.config.TerseErrors = true
	defer sqlQuery.handleExecError(&query, &err, logStats)
	panic(&TabletError{
		ErrorType: ErrFail,
		Message:   "msg",
		SqlError:  10,
	})
}

func TestNeedInvalidator(t *testing.T) {
	testUtils := newTestUtils()
	dbconfigs := testUtils.newDBConfigs()

	// EnableRowCache is false
	if needInvalidator(nil, &dbconfigs) {
		t.Errorf("got true, want false")
	}

	// EnableInvalidator is false
	dbconfigs.App.EnableRowcache = true
	if needInvalidator(nil, &dbconfigs) {
		t.Errorf("got true, want false")
	}

	dbconfigs.App.EnableInvalidator = true
	if !needInvalidator(nil, &dbconfigs) {
		t.Errorf("got false, want true")
	}

	target := &pb.Target{}
	// TabletType is not MASTER
	if !needInvalidator(target, &dbconfigs) {
		t.Errorf("got false, want true")
	}

	target.TabletType = topodata.TabletType_MASTER
	if needInvalidator(target, &dbconfigs) {
		t.Errorf("got true, want false")
	}
}

func setUpSQLQueryTest() *fakesqldb.DB {
	db := fakesqldb.Register()
	for query, result := range getSupportedQueries() {
		db.AddQuery(query, result)
	}
	return db
}

func checkSQLQueryState(t *testing.T, sqlQuery *SqlQuery, expectState int64) {
	sqlQuery.mu.Lock()
	state := sqlQuery.state
	sqlQuery.mu.Unlock()
	if state != expectState {
		t.Fatalf("sqlquery should in state: %d, but get state: %d", expectState, state)
	}
}

func getSupportedQueries() map[string]*mproto.QueryResult {
	return map[string]*mproto.QueryResult{
		// queries for schema info
		"select unix_timestamp()": &mproto.QueryResult{
			RowsAffected: 1,
			Rows: [][]sqltypes.Value{
				[]sqltypes.Value{sqltypes.MakeString([]byte("1427325875"))},
			},
		},
		"select @@global.sql_mode": &mproto.QueryResult{
			RowsAffected: 1,
			Rows: [][]sqltypes.Value{
				[]sqltypes.Value{sqltypes.MakeString([]byte("STRICT_TRANS_TABLES"))},
			},
		},
		"select * from test_table where 1 != 1": &mproto.QueryResult{
			Fields: getTestTableFields(),
		},
		baseShowTables: &mproto.QueryResult{
			RowsAffected: 1,
			Rows: [][]sqltypes.Value{
				[]sqltypes.Value{
					sqltypes.MakeString([]byte("test_table")),
					sqltypes.MakeString([]byte("USER TABLE")),
					sqltypes.MakeString([]byte("1427325875")),
					sqltypes.MakeString([]byte("")),
					sqltypes.MakeString([]byte("1")),
					sqltypes.MakeString([]byte("2")),
					sqltypes.MakeString([]byte("3")),
					sqltypes.MakeString([]byte("4")),
				},
			},
		},
		"describe `test_table`": &mproto.QueryResult{
			RowsAffected: 3,
			Rows: [][]sqltypes.Value{
				[]sqltypes.Value{
					sqltypes.MakeString([]byte("pk")),
					sqltypes.MakeString([]byte("int")),
					sqltypes.MakeString([]byte{}),
					sqltypes.MakeString([]byte{}),
					sqltypes.MakeString([]byte("1")),
					sqltypes.MakeString([]byte{}),
				},
				[]sqltypes.Value{
					sqltypes.MakeString([]byte("name")),
					sqltypes.MakeString([]byte("int")),
					sqltypes.MakeString([]byte{}),
					sqltypes.MakeString([]byte{}),
					sqltypes.MakeString([]byte("1")),
					sqltypes.MakeString([]byte{}),
				},
				[]sqltypes.Value{
					sqltypes.MakeString([]byte("addr")),
					sqltypes.MakeString([]byte("int")),
					sqltypes.MakeString([]byte{}),
					sqltypes.MakeString([]byte{}),
					sqltypes.MakeString([]byte("1")),
					sqltypes.MakeString([]byte{}),
				},
			},
		},
		// for SplitQuery because it needs a primary key column
		"show index from `test_table`": &mproto.QueryResult{
			RowsAffected: 2,
			Rows: [][]sqltypes.Value{
				[]sqltypes.Value{
					sqltypes.MakeString([]byte{}),
					sqltypes.MakeString([]byte{}),
					sqltypes.MakeString([]byte("PRIMARY")),
					sqltypes.MakeString([]byte{}),
					sqltypes.MakeString([]byte("pk")),
					sqltypes.MakeString([]byte{}),
					sqltypes.MakeString([]byte("300")),
				},
				[]sqltypes.Value{
					sqltypes.MakeString([]byte{}),
					sqltypes.MakeString([]byte{}),
					sqltypes.MakeString([]byte("INDEX")),
					sqltypes.MakeString([]byte{}),
					sqltypes.MakeString([]byte("name")),
					sqltypes.MakeString([]byte{}),
					sqltypes.MakeString([]byte("300")),
				},
			},
		},
		"begin":    &mproto.QueryResult{},
		"commit":   &mproto.QueryResult{},
		"rollback": &mproto.QueryResult{},
		baseShowTables + " and table_name = 'test_table'": &mproto.QueryResult{
			RowsAffected: 1,
			Rows: [][]sqltypes.Value{
				[]sqltypes.Value{
					sqltypes.MakeString([]byte("test_table")),
					sqltypes.MakeString([]byte("USER TABLE")),
					sqltypes.MakeString([]byte("1427325875")),
					sqltypes.MakeString([]byte("")),
					sqltypes.MakeString([]byte("1")),
					sqltypes.MakeString([]byte("2")),
					sqltypes.MakeString([]byte("3")),
					sqltypes.MakeString([]byte("4")),
				},
			},
		},
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
