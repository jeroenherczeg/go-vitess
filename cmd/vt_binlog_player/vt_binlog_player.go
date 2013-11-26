// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
The vt_binlog_player reads data from the a remote host via vt_binlog_server.
This is mostly intended for online data migrations.
This program reads the current status from blp_recovery (by uid),
and updates it.
*/
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	log "github.com/golang/glog"
	"github.com/youtube/vitess/go/mysql"
	"github.com/youtube/vitess/go/vt/key"
	"github.com/youtube/vitess/go/vt/mysqlctl"
	"github.com/youtube/vitess/go/vt/servenv"
)

var (
	uid          = flag.Uint("uid", 0, "id of the blp_checkpoint row")
	start        = flag.String("start", "", "keyrange start to use in hex")
	end          = flag.String("end", "", "keyrange end to use in hex")
	port         = flag.Int("port", 0, "port for the server")
	dbConfigFile = flag.String("db-config-file", "", "json file for db credentials")
	debug        = flag.Bool("debug", true, "run a debug version - prints the sql statements rather than executing them")
)

func readDbConfig(dbConfigFile string) (*mysql.ConnectionParams, error) {
	dbConfigData, err := ioutil.ReadFile(dbConfigFile)
	if err != nil {
		return nil, fmt.Errorf("Error %s in reading db-config-file %s", err, dbConfigFile)
	}
	log.Infof("dbConfigData %v", string(dbConfigData))

	dbConfig := new(mysql.ConnectionParams)
	err = json.Unmarshal(dbConfigData, dbConfig)
	if err != nil {
		return nil, fmt.Errorf("error in unmarshaling dbconfig data, err '%v'", err)
	}
	return dbConfig, nil
}

// TODO: Either write a test for this tool or delete it.
func main() {
	flag.Parse()
	servenv.Init()
	defer servenv.Close()

	keyRange, err := key.ParseKeyRangeParts(*start, *end)
	if err != nil {
		log.Fatalf("Invalid key range: %v", err)
	}

	if *dbConfigFile == "" {
		log.Fatalf("Cannot start without db-config-file")
	}
	dbConfig, err := readDbConfig(*dbConfigFile)
	if err != nil {
		log.Fatalf("Cannot read db config file: %v", err)
	}

	interrupted := make(chan struct{})
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		for _ = range c {
			close(interrupted)
		}
	}()

	var vtClient mysqlctl.VtClient
	vtClient = mysqlctl.NewDbClient(dbConfig)
	err = vtClient.Connect()
	if err != nil {
		log.Fatalf("error in initializing dbClient: %v", err)
	}
	brs, err := mysqlctl.ReadStartPosition(vtClient, uint32(*uid))
	if err != nil {
		log.Fatalf("Cannot read start position from db: %v", err)
	}
	if *debug {
		vtClient = mysqlctl.NewDummyVtClient()
	}
	blp := mysqlctl.NewBinlogPlayer(vtClient, fmt.Sprintf("localhost:%d", *port), keyRange, brs)
	err = blp.ApplyBinlogEvents(interrupted)
	if err != nil {
		log.Errorf("Error in applying binlog events, err %v", err)
	}
	log.Infof("vt_binlog_player done")
}
