// Copyright 2015, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tableacl

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/youtube/vitess/go/testfiles"
	tableaclpb "github.com/youtube/vitess/go/vt/proto/tableacl"
	"github.com/youtube/vitess/go/vt/tableacl/acl"
	"github.com/youtube/vitess/go/vt/tableacl/simpleacl"
)

type fakeAclFactory struct{}

func (factory *fakeAclFactory) New(entries []string) (acl.ACL, error) {
	return nil, fmt.Errorf("unable to create a new ACL")
}

type fakeACL struct{}

func (acl *fakeACL) IsMember(principal string) bool {
	return false
}

func TestInitWithInvalidFilePath(t *testing.T) {
	setUpTableACL(&simpleacl.Factory{})
	defer func() {
		err := recover()
		if err == nil {
			t.Fatalf("init should fail for an invalid config file path")
		}
	}()
	Init("/invalid_file_path")
}

func TestInitWithValidConfig(t *testing.T) {
	setUpTableACL(&simpleacl.Factory{})
	Init(testfiles.Locate("tableacl/test_table_tableacl_config.json"))
}

func TestInitFromProto(t *testing.T) {
	setUpTableACL(&simpleacl.Factory{})
	readerACL := Authorized("my_test_table", READER)
	if !reflect.DeepEqual(readerACL, acl.AcceptAllACL{}) {
		t.Fatalf("tableacl has not been initialized, got: %v, want: %v", readerACL, acl.AcceptAllACL{})
	}
	config := &tableaclpb.Config{
		TableGroups: []*tableaclpb.TableGroupSpec{{
			Name:                 "group01",
			TableNamesOrPrefixes: []string{"test_table"},
			Readers:              []string{"vt"},
		}},
	}
	if err := InitFromProto(config); err != nil {
		t.Fatalf("tableacl init should succeed, but got error: %v", err)
	}

	readerACL = Authorized("unknown_table", READER)
	if !reflect.DeepEqual(acl.AcceptAllACL{}, readerACL) {
		t.Fatalf("there is no config for unknown_table, should grand all permissions")
	}

	readerACL = Authorized("test_table", READER)
	if !readerACL.IsMember("vt") {
		t.Fatalf("user: vt should have reader permission to table: test_table")
	}
}

func TestTableACLAuthorize(t *testing.T) {
	setUpTableACL(&simpleacl.Factory{})
	config := &tableaclpb.Config{
		TableGroups: []*tableaclpb.TableGroupSpec{
			{
				Name:                 "group01",
				TableNamesOrPrefixes: []string{"test_music"},
				Readers:              []string{"u1", "u2"},
				Writers:              []string{"u1", "u3"},
				Admins:               []string{"u1"},
			},
			{
				Name:                 "group02",
				TableNamesOrPrefixes: []string{"test_music", "test_video"},
				Readers:              []string{"u1", "u2"},
				Writers:              []string{"u3"},
				Admins:               []string{"u4"},
			},
			{
				Name:                 "group03",
				TableNamesOrPrefixes: []string{"test_other%"},
				Readers:              []string{"u2"},
				Writers:              []string{"u2", "u3"},
				Admins:               []string{"u3"},
			},
			{
				Name:                 "group04",
				TableNamesOrPrefixes: []string{"test_data%"},
				Readers:              []string{"u1", "u2"},
				Writers:              []string{"u1", "u3"},
				Admins:               []string{"u1"},
			},
		},
	}
	if err := InitFromProto(config); err != nil {
		t.Fatalf("InitFromProto(<data>) = %v, want: nil", err)
	}

	readerACL := Authorized("test_data_any", READER)
	if !readerACL.IsMember("u1") {
		t.Fatalf("user u1 should have reader permission to table test_data_any")
	}
	if !readerACL.IsMember("u2") {
		t.Fatalf("user u2 should have reader permission to table test_data_any")
	}
}

func TestFailedToCreateACL(t *testing.T) {
	setUpTableACL(&fakeAclFactory{})
	config := &tableaclpb.Config{
		TableGroups: []*tableaclpb.TableGroupSpec{{
			Name:                 "group01",
			TableNamesOrPrefixes: []string{"test_table"},
			Readers:              []string{"vt"},
			Writers:              []string{"vt"},
		}},
	}
	if err := InitFromProto(config); err == nil {
		t.Fatalf("tableacl init should fail because fake ACL returns an error")
	}
}

func TestDoubleRegisterTheSameKey(t *testing.T) {
	acls = make(map[string]acl.Factory)
	name := fmt.Sprintf("tableacl-name-%d", rand.Int63())
	Register(name, &simpleacl.Factory{})
	defer func() {
		err := recover()
		if err == nil {
			t.Fatalf("the second tableacl register should fail")
		}
	}()
	Register(name, &simpleacl.Factory{})
}

func TestGetAclFactory(t *testing.T) {
	acls = make(map[string]acl.Factory)
	defaultACL = ""
	name := fmt.Sprintf("tableacl-name-%d", rand.Int63())
	aclFactory := &simpleacl.Factory{}
	Register(name, aclFactory)
	if !reflect.DeepEqual(aclFactory, GetCurrentAclFactory()) {
		t.Fatalf("should return registered acl factory even if default acl is not set.")
	}
	Register(name+"2", aclFactory)
	defer func() {
		err := recover()
		if err == nil {
			t.Fatalf("there are more than one acl factories, but the default is not set")
		}
	}()
	GetCurrentAclFactory()
}

func TestGetAclFactoryWithWrongDefault(t *testing.T) {
	acls = make(map[string]acl.Factory)
	defaultACL = ""
	name := fmt.Sprintf("tableacl-name-%d", rand.Int63())
	aclFactory := &simpleacl.Factory{}
	Register(name, aclFactory)
	Register(name+"2", aclFactory)
	SetDefaultACL("wrong_name")
	defer func() {
		err := recover()
		if err == nil {
			t.Fatalf("there are more than one acl factories, but the default given does not match any of these.")
		}
	}()
	GetCurrentAclFactory()
}

func setUpTableACL(factory acl.Factory) {
	name := fmt.Sprintf("tableacl-name-%d", rand.Int63())
	Register(name, factory)
	SetDefaultACL(name)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
