// Copyright 2014, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vindexes

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/youtube/vitess/go/cistring"
	vschemapb "github.com/youtube/vitess/go/vt/proto/vschema"
)

// stFU satisfies Functional, Unique.
type stFU struct {
	name   string
	Params map[string]string
}

func (v *stFU) String() string                                  { return v.name }
func (*stFU) Cost() int                                         { return 1 }
func (*stFU) Verify(VCursor, interface{}, []byte) (bool, error) { return false, nil }
func (*stFU) Map(VCursor, []interface{}) ([][]byte, error)      { return nil, nil }

func NewSTFU(name string, params map[string]string) (Vindex, error) {
	return &stFU{name: name, Params: params}, nil
}

// stF satisfies Functional, but no Map. Invalid vindex.
type stF struct {
	name   string
	Params map[string]string
}

func (v *stF) String() string                                  { return v.name }
func (*stF) Cost() int                                         { return 0 }
func (*stF) Verify(VCursor, interface{}, []byte) (bool, error) { return false, nil }

func NewSTF(name string, params map[string]string) (Vindex, error) {
	return &stF{name: name, Params: params}, nil
}

// stLN satisfies Lookup, NonUnique.
type stLN struct {
	name   string
	Params map[string]string
}

func (v *stLN) String() string                                  { return v.name }
func (*stLN) Cost() int                                         { return 0 }
func (*stLN) Verify(VCursor, interface{}, []byte) (bool, error) { return false, nil }
func (*stLN) Map(VCursor, []interface{}) ([][][]byte, error)    { return nil, nil }
func (*stLN) Create(VCursor, interface{}, []byte) error         { return nil }
func (*stLN) Delete(VCursor, []interface{}, []byte) error       { return nil }

func NewSTLN(name string, params map[string]string) (Vindex, error) {
	return &stLN{name: name, Params: params}, nil
}

// stLU satisfies Lookup, Unique.
type stLU struct {
	name   string
	Params map[string]string
}

func (v *stLU) String() string                                  { return v.name }
func (*stLU) Cost() int                                         { return 2 }
func (*stLU) Verify(VCursor, interface{}, []byte) (bool, error) { return false, nil }
func (*stLU) Map(VCursor, []interface{}) ([][]byte, error)      { return nil, nil }
func (*stLU) Create(VCursor, interface{}, []byte) error         { return nil }
func (*stLU) Delete(VCursor, []interface{}, []byte) error       { return nil }

func NewSTLU(name string, params map[string]string) (Vindex, error) {
	return &stLU{name: name, Params: params}, nil
}

func init() {
	Register("stfu", NewSTFU)
	Register("stf", NewSTF)
	Register("stln", NewSTLN)
	Register("stlu", NewSTLU)
}

func TestUnshardedVSchema(t *testing.T) {
	good := VSchemaFormal{
		Keyspaces: map[string]vschemapb.Keyspace{
			"unsharded": {
				Tables: map[string]*vschemapb.Table{
					"t1": {},
				},
			},
		},
	}
	got, err := BuildVSchema(&good)
	if err != nil {
		t.Error(err)
	}
	ks := &Keyspace{
		Name: "unsharded",
	}
	t1 := &Table{
		Name:     "t1",
		Keyspace: ks,
	}
	want := &VSchema{
		tables: map[string]*Table{
			"t1": t1,
		},
		Keyspaces: map[string]*KeyspaceSchema{
			"unsharded": {
				Keyspace: ks,
				Tables: map[string]*Table{
					"t1": t1,
				},
			},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("BuildVSchema:s\n%v, want\n%v", got, want)
	}
}

func TestShardedVSchemaOwned(t *testing.T) {
	good := VSchemaFormal{
		Keyspaces: map[string]vschemapb.Keyspace{
			"sharded": {
				Sharded: true,
				Vindexes: map[string]*vschemapb.Vindex{
					"stfu1": {
						Type: "stfu",
						Params: map[string]string{
							"stfu1": "1",
						},
						Owner: "t1",
					},
					"stln1": {
						Type:  "stln",
						Owner: "t1",
					},
				},
				Tables: map[string]*vschemapb.Table{
					"t1": {
						ColumnVindexes: []*vschemapb.ColumnVindex{
							{
								Column: "c1",
								Name:   "stfu1",
							}, {
								Column: "c2",
								Name:   "stln1",
							},
						},
					},
				},
			},
		},
	}
	got, err := BuildVSchema(&good)
	if err != nil {
		t.Error(err)
	}
	ks := &Keyspace{
		Name:    "sharded",
		Sharded: true,
	}
	t1 := &Table{
		Name:     "t1",
		Keyspace: ks,
		ColumnVindexes: []*ColumnVindex{
			{
				Column: cistring.New("c1"),
				Type:   "stfu",
				Name:   "stfu1",
				Vindex: &stFU{
					name: "stfu1",
					Params: map[string]string{
						"stfu1": "1",
					},
				},
			},
			{
				Column: cistring.New("c2"),
				Type:   "stln",
				Name:   "stln1",
				Owned:  true,
				Vindex: &stLN{name: "stln1"},
			},
		},
	}
	t1.Ordered = []*ColumnVindex{
		t1.ColumnVindexes[1],
		t1.ColumnVindexes[0],
	}
	t1.Owned = t1.ColumnVindexes[1:]
	want := &VSchema{
		tables: map[string]*Table{
			"t1": t1,
		},
		Keyspaces: map[string]*KeyspaceSchema{
			"sharded": {
				Keyspace: ks,
				Tables: map[string]*Table{
					"t1": t1,
				},
			},
		},
	}
	if !reflect.DeepEqual(got, want) {
		gotjson, _ := json.Marshal(got)
		wantjson, _ := json.Marshal(want)
		t.Errorf("BuildVSchema:s\n%s, want\n%s", gotjson, wantjson)
	}
}

func TestShardedVSchemaNotOwned(t *testing.T) {
	good := VSchemaFormal{
		Keyspaces: map[string]vschemapb.Keyspace{
			"sharded": {
				Sharded: true,
				Vindexes: map[string]*vschemapb.Vindex{
					"stlu1": {
						Type:  "stlu",
						Owner: "",
					},
					"stfu1": {
						Type:  "stfu",
						Owner: "",
					},
				},
				Tables: map[string]*vschemapb.Table{
					"t1": {
						ColumnVindexes: []*vschemapb.ColumnVindex{
							{
								Column: "c1",
								Name:   "stlu1",
							}, {
								Column: "c2",
								Name:   "stfu1",
							},
						},
					},
				},
			},
		},
	}
	got, err := BuildVSchema(&good)
	if err != nil {
		t.Error(err)
	}
	ks := &Keyspace{
		Name:    "sharded",
		Sharded: true,
	}
	t1 := &Table{
		Name:     "t1",
		Keyspace: ks,
		ColumnVindexes: []*ColumnVindex{
			{
				Column: cistring.New("c1"),
				Type:   "stlu",
				Name:   "stlu1",
				Owned:  false,
				Vindex: &stLU{name: "stlu1"},
			},
			{
				Column: cistring.New("c2"),
				Type:   "stfu",
				Name:   "stfu1",
				Owned:  false,
				Vindex: &stFU{name: "stfu1"},
			},
		},
	}
	t1.Ordered = []*ColumnVindex{
		t1.ColumnVindexes[1],
		t1.ColumnVindexes[0],
	}
	want := &VSchema{
		tables: map[string]*Table{
			"t1": t1,
		},
		Keyspaces: map[string]*KeyspaceSchema{
			"sharded": {
				Keyspace: ks,
				Tables: map[string]*Table{
					"t1": t1,
				},
			},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("BuildVSchema:s\n%v, want\n%v", got, want)
	}
}

func TestBuildVSchemaVindexNotFoundFail(t *testing.T) {
	bad := VSchemaFormal{
		Keyspaces: map[string]vschemapb.Keyspace{
			"sharded": {
				Sharded: true,
				Vindexes: map[string]*vschemapb.Vindex{
					"noexist": {
						Type: "noexist",
					},
				},
				Tables: map[string]*vschemapb.Table{
					"t1": {
						ColumnVindexes: []*vschemapb.ColumnVindex{
							{
								Column: "c1",
								Name:   "noexist",
							},
						},
					},
				},
			},
		},
	}
	_, err := BuildVSchema(&bad)
	want := "vindexType noexist not found"
	if err == nil || err.Error() != want {
		t.Errorf("BuildVSchema: %v, want %v", err, want)
	}
}

func TestBuildVSchemaNoColumnVindexFail(t *testing.T) {
	bad := VSchemaFormal{
		Keyspaces: map[string]vschemapb.Keyspace{
			"sharded": {
				Sharded: true,
				Vindexes: map[string]*vschemapb.Vindex{
					"stfu": {
						Type: "stfu",
					},
				},
				Tables: map[string]*vschemapb.Table{
					"t1": {},
				},
			},
		},
	}
	_, err := BuildVSchema(&bad)
	want := "missing primary col vindex for table: t1"
	if err == nil || err.Error() != want {
		t.Errorf("BuildVSchema: %v, want %v", err, want)
	}
}

func TestBuildVSchemaInvalidVindexFail(t *testing.T) {
	bad := VSchemaFormal{
		Keyspaces: map[string]vschemapb.Keyspace{
			"sharded": {
				Sharded: true,
				Vindexes: map[string]*vschemapb.Vindex{
					"stf": {
						Type: "stf",
					},
				},
				Tables: map[string]*vschemapb.Table{
					"t1": {
						ColumnVindexes: []*vschemapb.ColumnVindex{
							{
								Column: "c1",
								Name:   "stf",
							},
						},
					},
				},
			},
		},
	}
	_, err := BuildVSchema(&bad)
	want := "vindex stf needs to be Unique or NonUnique"
	if err == nil || err.Error() != want {
		t.Errorf("BuildVSchema: %v, want %v", err, want)
	}
}

func TestBuildVSchemaDupSeq(t *testing.T) {
	good := VSchemaFormal{
		Keyspaces: map[string]vschemapb.Keyspace{
			"ksa": {
				Tables: map[string]*vschemapb.Table{
					"t1": {
						Type: "sequence",
					},
				},
			},
			"ksb": {
				Tables: map[string]*vschemapb.Table{
					"t1": {
						Type: "sequence",
					},
				},
			},
		},
	}
	ksa := &Keyspace{
		Name: "ksa",
	}
	ksb := &Keyspace{
		Name: "ksb",
	}
	got, _ := BuildVSchema(&good)
	t1a := &Table{
		Name:       "t1",
		Keyspace:   ksa,
		IsSequence: true,
	}
	t1b := &Table{
		Name:       "t1",
		Keyspace:   ksb,
		IsSequence: true,
	}
	want := &VSchema{
		tables: map[string]*Table{
			"t1": nil,
		},
		Keyspaces: map[string]*KeyspaceSchema{
			"ksa": {
				Keyspace: ksa,
				Tables: map[string]*Table{
					"t1": t1a,
				},
			},
			"ksb": {
				Keyspace: ksb,
				Tables: map[string]*Table{
					"t1": t1b,
				},
			},
		},
	}
	if !reflect.DeepEqual(got, want) {
		gotjson, _ := json.Marshal(got)
		wantjson, _ := json.Marshal(want)
		t.Errorf("BuildVSchema:s\n%s, want\n%s", gotjson, wantjson)
	}
}

func TestBuildVSchemaDupTable(t *testing.T) {
	good := VSchemaFormal{
		Keyspaces: map[string]vschemapb.Keyspace{
			"ksa": {
				Tables: map[string]*vschemapb.Table{
					"t1": {},
				},
			},
			"ksb": {
				Tables: map[string]*vschemapb.Table{
					"t1": {},
				},
			},
		},
	}
	got, _ := BuildVSchema(&good)
	ksa := &Keyspace{
		Name: "ksa",
	}
	t1a := &Table{
		Name:     "t1",
		Keyspace: ksa,
	}
	ksb := &Keyspace{
		Name: "ksb",
	}
	t1b := &Table{
		Name:     "t1",
		Keyspace: ksb,
	}
	want := &VSchema{
		tables: map[string]*Table{
			"t1": nil,
		},
		Keyspaces: map[string]*KeyspaceSchema{
			"ksa": {
				Keyspace: ksa,
				Tables: map[string]*Table{
					"t1": t1a,
				},
			},
			"ksb": {
				Keyspace: ksb,
				Tables: map[string]*Table{
					"t1": t1b,
				},
			},
		},
	}
	if !reflect.DeepEqual(got, want) {
		gotjson, _ := json.Marshal(got)
		wantjson, _ := json.Marshal(want)
		t.Errorf("BuildVSchema:s\n%s, want\n%s", gotjson, wantjson)
	}
}

func TestBuildVSchemaNoindexFail(t *testing.T) {
	bad := VSchemaFormal{
		Keyspaces: map[string]vschemapb.Keyspace{
			"sharded": {
				Sharded: true,
				Vindexes: map[string]*vschemapb.Vindex{
					"stfu": {
						Type: "stfu",
					},
				},
				Tables: map[string]*vschemapb.Table{
					"t1": {
						ColumnVindexes: []*vschemapb.ColumnVindex{
							{
								Column: "c1",
								Name:   "notexist",
							},
						},
					},
				},
			},
		},
	}
	_, err := BuildVSchema(&bad)
	want := "vindex notexist not found for table t1"
	if err == nil || err.Error() != want {
		t.Errorf("BuildVSchema: %v, want %v", err, want)
	}
}

func TestBuildVSchemaNotUniqueFail(t *testing.T) {
	bad := VSchemaFormal{
		Keyspaces: map[string]vschemapb.Keyspace{
			"sharded": {
				Sharded: true,
				Vindexes: map[string]*vschemapb.Vindex{
					"stln": {
						Type: "stln",
					},
				},
				Tables: map[string]*vschemapb.Table{
					"t1": {
						ColumnVindexes: []*vschemapb.ColumnVindex{
							{
								Column: "c1",
								Name:   "stln",
							},
						},
					},
				},
			},
		},
	}
	_, err := BuildVSchema(&bad)
	want := "primary vindex stln is not Unique for table t1"
	if err == nil || err.Error() != want {
		t.Errorf("BuildVSchema: %v, want %v", err, want)
	}
}

func TestBuildVSchemaPrimaryNonFunctionalFail(t *testing.T) {
	bad := VSchemaFormal{
		Keyspaces: map[string]vschemapb.Keyspace{
			"sharded": {
				Sharded: true,
				Vindexes: map[string]*vschemapb.Vindex{
					"stlu": {
						Type:  "stlu",
						Owner: "t1",
					},
				},
				Tables: map[string]*vschemapb.Table{
					"t1": {
						ColumnVindexes: []*vschemapb.ColumnVindex{
							{
								Column: "c1",
								Name:   "stlu",
							},
						},
					},
				},
			},
		},
	}
	_, err := BuildVSchema(&bad)
	want := "primary vindex stlu cannot be owned for table t1"
	if err == nil || err.Error() != want {
		t.Errorf("BuildVSchema: %v, want %v", err, want)
	}
}

func TestSequence(t *testing.T) {
	good := VSchemaFormal{
		Keyspaces: map[string]vschemapb.Keyspace{
			"unsharded": {
				Tables: map[string]*vschemapb.Table{
					"seq": {
						Type: "sequence",
					},
				},
			},
			"sharded": {
				Sharded: true,
				Vindexes: map[string]*vschemapb.Vindex{
					"stfu1": {
						Type: "stfu",
						Params: map[string]string{
							"stfu1": "1",
						},
					},
				},
				Tables: map[string]*vschemapb.Table{
					"t1": {
						ColumnVindexes: []*vschemapb.ColumnVindex{
							{
								Column: "c1",
								Name:   "stfu1",
							},
						},
						AutoIncrement: &vschemapb.AutoIncrement{
							Column:   "c1",
							Sequence: "seq",
						},
					},
				},
			},
		},
	}
	got, err := BuildVSchema(&good)
	if err != nil {
		t.Error(err)
	}
	ksu := &Keyspace{
		Name: "unsharded",
	}
	kss := &Keyspace{
		Name:    "sharded",
		Sharded: true,
	}
	seq := &Table{
		Name:       "seq",
		Keyspace:   ksu,
		IsSequence: true,
	}
	t1 := &Table{
		Name:     "t1",
		Keyspace: kss,
		ColumnVindexes: []*ColumnVindex{
			{
				Column: cistring.New("c1"),
				Type:   "stfu",
				Name:   "stfu1",
				Vindex: &stFU{
					name: "stfu1",
					Params: map[string]string{
						"stfu1": "1",
					},
				},
			},
		},
		AutoIncrement: &AutoIncrement{
			Column:   cistring.New("c1"),
			Sequence: seq,
		},
	}
	t1.Ordered = []*ColumnVindex{
		t1.ColumnVindexes[0],
	}
	want := &VSchema{
		tables: map[string]*Table{
			"seq": seq,
			"t1":  t1,
		},
		Keyspaces: map[string]*KeyspaceSchema{
			"unsharded": {
				Keyspace: ksu,
				Tables: map[string]*Table{
					"seq": seq,
				},
			},
			"sharded": {
				Keyspace: kss,
				Tables: map[string]*Table{
					"t1": t1,
				},
			},
		},
	}
	if !reflect.DeepEqual(got, want) {
		gotjson, _ := json.Marshal(got)
		wantjson, _ := json.Marshal(want)
		t.Errorf("BuildVSchema:s\n%s, want\n%s", gotjson, wantjson)
	}
}

func TestBadSequence(t *testing.T) {
	bad := VSchemaFormal{
		Keyspaces: map[string]vschemapb.Keyspace{
			"sharded": {
				Sharded: true,
				Vindexes: map[string]*vschemapb.Vindex{
					"stfu1": {
						Type: "stfu",
					},
				},
				Tables: map[string]*vschemapb.Table{
					"t1": {
						ColumnVindexes: []*vschemapb.ColumnVindex{
							{
								Column: "c1",
								Name:   "stfu1",
							},
						},
						AutoIncrement: &vschemapb.AutoIncrement{
							Column:   "c1",
							Sequence: "seq",
						},
					},
				},
			},
		},
	}
	_, err := BuildVSchema(&bad)
	want := "sequence seq not found for table t1"
	if err == nil || err.Error() != want {
		t.Errorf("BuildVSchema: %v, want %v", err, want)
	}
}

func TestFind(t *testing.T) {
	input := VSchemaFormal{
		Keyspaces: map[string]vschemapb.Keyspace{
			"ksa": {
				Tables: map[string]*vschemapb.Table{
					"ta": {},
					"t1": {},
				},
			},
			"ksb": {
				Sharded: true,
				Vindexes: map[string]*vschemapb.Vindex{
					"stfu1": {
						Type: "stfu",
						Params: map[string]string{
							"stfu1": "1",
						},
					},
				},
				Tables: map[string]*vschemapb.Table{
					"tb": {
						ColumnVindexes: []*vschemapb.ColumnVindex{
							{
								Column: "c1",
								Name:   "stfu1",
							},
						},
					},
					"t1": {
						ColumnVindexes: []*vschemapb.ColumnVindex{
							{
								Column: "c1",
								Name:   "stfu1",
							},
						},
					},
				},
			},
		},
	}
	vschema, _ := BuildVSchema(&input)
	_, err := vschema.Find("", "t1")
	wantErr := "ambiguous table reference: t1"
	if err == nil || err.Error() != wantErr {
		t.Errorf("Find(\"\"): %v, want %s", err, wantErr)
	}
	_, err = vschema.Find("", "none")
	wantErr = "table none not found"
	if err == nil || err.Error() != wantErr {
		t.Errorf("Find(\"\"): %v, want %s", err, wantErr)
	}
	got, err := vschema.Find("", "ta")
	if err != nil {
		t.Error(err)
		return
	}
	ta := &Table{
		Name: "ta",
		Keyspace: &Keyspace{
			Name: "ksa",
		},
	}
	if !reflect.DeepEqual(got, ta) {
		t.Errorf("Find(\"t1a\"): %+v, want %+v", got, ta)
	}
	got, err = vschema.Find("ksa", "ta")
	if !reflect.DeepEqual(got, ta) {
		t.Errorf("Find(\"t1a\"): %+v, want %+v", got, ta)
	}
	none := &Table{
		Name: "none",
		Keyspace: &Keyspace{
			Name: "ksa",
		},
	}
	got, err = vschema.Find("ksa", "none")
	if !reflect.DeepEqual(got, none) {
		t.Errorf("Find(\"t1a\"): %+v, want %+v", got, none)
	}
	_, err = vschema.Find("ksb", "none")
	wantErr = "table none not found"
	if err == nil || err.Error() != wantErr {
		t.Errorf("Find(\"\"): %v, want %s", err, wantErr)
	}
	_, err = vschema.Find("none", "aa")
	wantErr = "keyspace none not found in vschema"
	if err == nil || err.Error() != wantErr {
		t.Errorf("Find(\"\"): %v, want %s", err, wantErr)
	}
}

func TestBuildKeyspaceSchema(t *testing.T) {
	good := &vschemapb.Keyspace{
		Tables: map[string]*vschemapb.Table{
			"t1": {
				AutoIncrement: &vschemapb.AutoIncrement{
					Column:   "col",
					Sequence: "outside",
				},
			},
			"t2": {},
		},
	}
	got, err := BuildKeyspaceSchema(good, "ks")
	if err != nil {
		t.Error(err)
	}
	ks := &Keyspace{
		Name: "ks",
	}
	t1 := &Table{
		Name:     "t1",
		Keyspace: ks,
	}
	t2 := &Table{
		Name:     "t2",
		Keyspace: ks,
	}
	want := &KeyspaceSchema{
		Keyspace: ks,
		Tables: map[string]*Table{
			"t1": t1,
			"t2": t2,
		},
	}
	if !reflect.DeepEqual(got, want) {
		gs, _ := json.Marshal(got)
		ws, _ := json.Marshal(want)
		t.Errorf("BuildKeyspaceSchema:\n%s, want\n%s", gs, ws)
	}
}

func TestValidate(t *testing.T) {
	good := &vschemapb.Keyspace{
		Tables: map[string]*vschemapb.Table{
			"t1": {
				AutoIncrement: &vschemapb.AutoIncrement{
					Column:   "col",
					Sequence: "outside",
				},
			},
			"t2": {},
		},
	}
	err := ValidateKeyspace(good)
	if err != nil {
		t.Error(err)
	}
	bad := &vschemapb.Keyspace{
		Sharded: true,
		Vindexes: map[string]*vschemapb.Vindex{
			"hash": {
				Type: "absent",
			},
		},
		Tables: map[string]*vschemapb.Table{
			"t2": {},
		},
	}
	err = ValidateKeyspace(bad)
	want := "vindexType absent not found"
	if err == nil || !strings.HasPrefix(err.Error(), want) {
		t.Errorf("Validate: %v, must start with %s", err, want)
	}
}

func TestVSchemaJSON(t *testing.T) {
	in := `
	{
		"sharded": true,
		"tables": {
			"t1": {
				"column_vindexes":[{
						"column":"c1",
						"name":"stfu1"
					},{
						"column":"c2",
						"name":"stln1"
					}],
				"auto_increment": {
					"column": "col",
					"sequence": "outside"
				}
			},
			"t2": {}
		}
	}
`
	var got vschemapb.Keyspace
	if err := json.Unmarshal([]byte(in), &got); err != nil {
		t.Error(err)
	}
	want := vschemapb.Keyspace{
		Sharded: true,
		Tables: map[string]*vschemapb.Table{
			"t1": {
				ColumnVindexes: []*vschemapb.ColumnVindex{
					{
						Column: "c1",
						Name:   "stfu1",
					}, {
						Column: "c2",
						Name:   "stln1",
					},
				},
				AutoIncrement: &vschemapb.AutoIncrement{
					Column:   "col",
					Sequence: "outside",
				},
			},
			"t2": {},
		},
	}
	if !reflect.DeepEqual(got, want) {
		gs, _ := json.Marshal(got)
		ws, _ := json.Marshal(want)
		t.Errorf("VSchemaFormalForKeyspace():\n%s, want\n%s", gs, ws)
	}
}

func TestFindSingleKeyspace(t *testing.T) {
	input := VSchemaFormal{
		Keyspaces: map[string]vschemapb.Keyspace{
			"ksa": {
				Tables: map[string]*vschemapb.Table{
					"ta": {},
					"t1": {},
				},
			},
		},
	}
	vschema, _ := BuildVSchema(&input)
	none := &Table{
		Name: "none",
		Keyspace: &Keyspace{
			Name: "ksa",
		},
	}
	got, _ := vschema.Find("", "none")
	if !reflect.DeepEqual(got, none) {
		t.Errorf("Find(\"t1a\"): %+v, want %+v", got, none)
	}
	input = VSchemaFormal{
		Keyspaces: map[string]vschemapb.Keyspace{
			"ksb": {
				Sharded: true,
				Vindexes: map[string]*vschemapb.Vindex{
					"stfu1": {
						Type: "stfu",
					},
				},
				Tables: map[string]*vschemapb.Table{
					"tb": {
						ColumnVindexes: []*vschemapb.ColumnVindex{
							{
								Column: "c1",
								Name:   "stfu1",
							},
						},
					},
					"t1": {
						ColumnVindexes: []*vschemapb.ColumnVindex{
							{
								Column: "c1",
								Name:   "stfu1",
							},
						},
					},
				},
			},
		},
	}
	vschema, _ = BuildVSchema(&input)
	_, err := vschema.Find("", "none")
	wantErr := "table none not found"
	if err == nil || err.Error() != wantErr {
		t.Errorf("Find(\"\"): %v, want %s", err, wantErr)
	}
}
