// Code generated by protoc-gen-go. DO NOT EDIT.
// source: vschema.proto

/*
Package vschema is a generated protocol buffer package.

It is generated from these files:
	vschema.proto

It has these top-level messages:
	Keyspace
	Vindex
	Table
	ColumnVindex
	AutoIncrement
	Column
	SrvVSchema
*/
package vschema

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import query "github.com/youtube/vitess/go/vt/proto/query"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// Keyspace is the vschema for a keyspace.
type Keyspace struct {
	// If sharded is false, vindexes and tables are ignored.
	Sharded  bool               `protobuf:"varint,1,opt,name=sharded" json:"sharded,omitempty"`
	Vindexes map[string]*Vindex `protobuf:"bytes,2,rep,name=vindexes" json:"vindexes,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	Tables   map[string]*Table  `protobuf:"bytes,3,rep,name=tables" json:"tables,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *Keyspace) Reset()                    { *m = Keyspace{} }
func (m *Keyspace) String() string            { return proto.CompactTextString(m) }
func (*Keyspace) ProtoMessage()               {}
func (*Keyspace) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Keyspace) GetSharded() bool {
	if m != nil {
		return m.Sharded
	}
	return false
}

func (m *Keyspace) GetVindexes() map[string]*Vindex {
	if m != nil {
		return m.Vindexes
	}
	return nil
}

func (m *Keyspace) GetTables() map[string]*Table {
	if m != nil {
		return m.Tables
	}
	return nil
}

// Vindex is the vindex info for a Keyspace.
type Vindex struct {
	// The type must match one of the predefined
	// (or plugged in) vindex names.
	Type string `protobuf:"bytes,1,opt,name=type" json:"type,omitempty"`
	// params is a map of attribute value pairs
	// that must be defined as required by the
	// vindex constructors. The values can only
	// be strings.
	Params map[string]string `protobuf:"bytes,2,rep,name=params" json:"params,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	// A lookup vindex can have an owner table defined.
	// If so, rows in the lookup table are created or
	// deleted in sync with corresponding rows in the
	// owner table.
	Owner string `protobuf:"bytes,3,opt,name=owner" json:"owner,omitempty"`
}

func (m *Vindex) Reset()                    { *m = Vindex{} }
func (m *Vindex) String() string            { return proto.CompactTextString(m) }
func (*Vindex) ProtoMessage()               {}
func (*Vindex) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *Vindex) GetType() string {
	if m != nil {
		return m.Type
	}
	return ""
}

func (m *Vindex) GetParams() map[string]string {
	if m != nil {
		return m.Params
	}
	return nil
}

func (m *Vindex) GetOwner() string {
	if m != nil {
		return m.Owner
	}
	return ""
}

// Table is the table info for a Keyspace.
type Table struct {
	// If the table is a sequence, type must be
	// "sequence". Otherwise, it should be empty.
	Type string `protobuf:"bytes,1,opt,name=type" json:"type,omitempty"`
	// column_vindexes associates columns to vindexes.
	ColumnVindexes []*ColumnVindex `protobuf:"bytes,2,rep,name=column_vindexes,json=columnVindexes" json:"column_vindexes,omitempty"`
	// auto_increment is specified if a column needs
	// to be associated with a sequence.
	AutoIncrement *AutoIncrement `protobuf:"bytes,3,opt,name=auto_increment,json=autoIncrement" json:"auto_increment,omitempty"`
	// columns lists the columns for the table.
	Columns []*Column `protobuf:"bytes,4,rep,name=columns" json:"columns,omitempty"`
}

func (m *Table) Reset()                    { *m = Table{} }
func (m *Table) String() string            { return proto.CompactTextString(m) }
func (*Table) ProtoMessage()               {}
func (*Table) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *Table) GetType() string {
	if m != nil {
		return m.Type
	}
	return ""
}

func (m *Table) GetColumnVindexes() []*ColumnVindex {
	if m != nil {
		return m.ColumnVindexes
	}
	return nil
}

func (m *Table) GetAutoIncrement() *AutoIncrement {
	if m != nil {
		return m.AutoIncrement
	}
	return nil
}

func (m *Table) GetColumns() []*Column {
	if m != nil {
		return m.Columns
	}
	return nil
}

// ColumnVindex is used to associate a column to a vindex.
type ColumnVindex struct {
	// Legacy implemenation, moving forward all vindexes should define a list of columns.
	Column string `protobuf:"bytes,1,opt,name=column" json:"column,omitempty"`
	// The name must match a vindex defined in Keyspace.
	Name string `protobuf:"bytes,2,opt,name=name" json:"name,omitempty"`
	// List of columns that define this Vindex
	Columns []string `protobuf:"bytes,3,rep,name=columns" json:"columns,omitempty"`
}

func (m *ColumnVindex) Reset()                    { *m = ColumnVindex{} }
func (m *ColumnVindex) String() string            { return proto.CompactTextString(m) }
func (*ColumnVindex) ProtoMessage()               {}
func (*ColumnVindex) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *ColumnVindex) GetColumn() string {
	if m != nil {
		return m.Column
	}
	return ""
}

func (m *ColumnVindex) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *ColumnVindex) GetColumns() []string {
	if m != nil {
		return m.Columns
	}
	return nil
}

// Autoincrement is used to designate a column as auto-inc.
type AutoIncrement struct {
	Column string `protobuf:"bytes,1,opt,name=column" json:"column,omitempty"`
	// The sequence must match a table of type SEQUENCE.
	Sequence string `protobuf:"bytes,2,opt,name=sequence" json:"sequence,omitempty"`
}

func (m *AutoIncrement) Reset()                    { *m = AutoIncrement{} }
func (m *AutoIncrement) String() string            { return proto.CompactTextString(m) }
func (*AutoIncrement) ProtoMessage()               {}
func (*AutoIncrement) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *AutoIncrement) GetColumn() string {
	if m != nil {
		return m.Column
	}
	return ""
}

func (m *AutoIncrement) GetSequence() string {
	if m != nil {
		return m.Sequence
	}
	return ""
}

// Column describes a column.
type Column struct {
	Name string     `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	Type query.Type `protobuf:"varint,2,opt,name=type,enum=query.Type" json:"type,omitempty"`
}

func (m *Column) Reset()                    { *m = Column{} }
func (m *Column) String() string            { return proto.CompactTextString(m) }
func (*Column) ProtoMessage()               {}
func (*Column) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *Column) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Column) GetType() query.Type {
	if m != nil {
		return m.Type
	}
	return query.Type_NULL_TYPE
}

// SrvVSchema is the roll-up of all the Keyspace schema for a cell.
type SrvVSchema struct {
	// keyspaces is a map of keyspace name -> Keyspace object.
	Keyspaces map[string]*Keyspace `protobuf:"bytes,1,rep,name=keyspaces" json:"keyspaces,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *SrvVSchema) Reset()                    { *m = SrvVSchema{} }
func (m *SrvVSchema) String() string            { return proto.CompactTextString(m) }
func (*SrvVSchema) ProtoMessage()               {}
func (*SrvVSchema) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *SrvVSchema) GetKeyspaces() map[string]*Keyspace {
	if m != nil {
		return m.Keyspaces
	}
	return nil
}

func init() {
	proto.RegisterType((*Keyspace)(nil), "vschema.Keyspace")
	proto.RegisterType((*Vindex)(nil), "vschema.Vindex")
	proto.RegisterType((*Table)(nil), "vschema.Table")
	proto.RegisterType((*ColumnVindex)(nil), "vschema.ColumnVindex")
	proto.RegisterType((*AutoIncrement)(nil), "vschema.AutoIncrement")
	proto.RegisterType((*Column)(nil), "vschema.Column")
	proto.RegisterType((*SrvVSchema)(nil), "vschema.SrvVSchema")
}

func init() { proto.RegisterFile("vschema.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 445 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x53, 0x5d, 0x6b, 0xd4, 0x40,
	0x14, 0x65, 0x76, 0xdd, 0x74, 0xf7, 0xc6, 0x4d, 0x75, 0xa8, 0x25, 0x44, 0xc4, 0x25, 0x28, 0xee,
	0x53, 0x1e, 0xb6, 0x08, 0x7e, 0xa0, 0x28, 0xc5, 0x87, 0xa2, 0xa0, 0xa4, 0xa5, 0xaf, 0x65, 0x9a,
	0xbd, 0xd0, 0xd2, 0xcd, 0x24, 0x66, 0x92, 0x68, 0xfe, 0x8a, 0x2f, 0x82, 0xff, 0xc0, 0x7f, 0x28,
	0x3b, 0x5f, 0x9d, 0x74, 0xe3, 0xdb, 0x1c, 0xce, 0x3d, 0x67, 0xce, 0x9d, 0x7b, 0x07, 0xe6, 0xad,
	0xc8, 0xae, 0x30, 0x67, 0x49, 0x59, 0x15, 0x75, 0x41, 0xf7, 0x34, 0x8c, 0xff, 0x8e, 0x60, 0xfa,
	0x19, 0x3b, 0x51, 0xb2, 0x0c, 0x69, 0x08, 0x7b, 0xe2, 0x8a, 0x55, 0x6b, 0x5c, 0x87, 0x64, 0x41,
	0x96, 0xd3, 0xd4, 0x40, 0xfa, 0x16, 0xa6, 0xed, 0x35, 0x5f, 0xe3, 0x4f, 0x14, 0xe1, 0x68, 0x31,
	0x5e, 0xfa, 0xab, 0xa7, 0x89, 0x71, 0x34, 0xf2, 0xe4, 0x5c, 0x57, 0x7c, 0xe2, 0x75, 0xd5, 0xa5,
	0x56, 0x40, 0x5f, 0x82, 0x57, 0xb3, 0xcb, 0x0d, 0x8a, 0x70, 0x2c, 0xa5, 0x4f, 0x76, 0xa5, 0x67,
	0x92, 0x57, 0x42, 0x5d, 0x1c, 0x7d, 0x81, 0x79, 0xcf, 0x91, 0x3e, 0x80, 0xf1, 0x0d, 0x76, 0x32,
	0xda, 0x2c, 0xdd, 0x1e, 0xe9, 0x73, 0x98, 0xb4, 0x6c, 0xd3, 0x60, 0x38, 0x5a, 0x90, 0xa5, 0xbf,
	0xda, 0xb7, 0xc6, 0x4a, 0x98, 0x2a, 0xf6, 0xcd, 0xe8, 0x15, 0x89, 0x4e, 0xc0, 0x77, 0x2e, 0x19,
	0xf0, 0x7a, 0xd6, 0xf7, 0x0a, 0xac, 0x97, 0x94, 0x39, 0x56, 0xf1, 0x1f, 0x02, 0x9e, 0xba, 0x80,
	0x52, 0xb8, 0x57, 0x77, 0x25, 0x6a, 0x1f, 0x79, 0xa6, 0x47, 0xe0, 0x95, 0xac, 0x62, 0xb9, 0x79,
	0xa9, 0xc7, 0x77, 0x52, 0x25, 0xdf, 0x24, 0xab, 0x9b, 0x55, 0xa5, 0xf4, 0x00, 0x26, 0xc5, 0x0f,
	0x8e, 0x55, 0x38, 0x96, 0x4e, 0x0a, 0x44, 0xaf, 0xc1, 0x77, 0x8a, 0x07, 0x42, 0x1f, 0xb8, 0xa1,
	0x67, 0x6e, 0xc8, 0x5f, 0x04, 0x26, 0x32, 0xf9, 0x60, 0xc6, 0xf7, 0xb0, 0x9f, 0x15, 0x9b, 0x26,
	0xe7, 0x17, 0x77, 0xc6, 0xfa, 0xc8, 0x86, 0x3d, 0x96, 0xbc, 0x7e, 0xc8, 0x20, 0x73, 0x10, 0x0a,
	0xfa, 0x0e, 0x02, 0xd6, 0xd4, 0xc5, 0xc5, 0x35, 0xcf, 0x2a, 0xcc, 0x91, 0xd7, 0x32, 0xb7, 0xbf,
	0x3a, 0xb4, 0xf2, 0x8f, 0x4d, 0x5d, 0x9c, 0x18, 0x36, 0x9d, 0x33, 0x17, 0xc6, 0x67, 0x70, 0xdf,
	0xb5, 0xa7, 0x87, 0xe0, 0xa9, 0x0b, 0x74, 0x48, 0x8d, 0xb6, 0xd1, 0x39, 0xcb, 0x4d, 0x77, 0xf2,
	0xbc, 0x5d, 0x52, 0xc5, 0xaa, 0x75, 0x9a, 0xa5, 0x06, 0xc6, 0xc7, 0x30, 0xef, 0xdd, 0xfa, 0x5f,
	0xdb, 0x08, 0xa6, 0x02, 0xbf, 0x37, 0xc8, 0x33, 0x63, 0x6d, 0x71, 0xfc, 0x9b, 0x00, 0x9c, 0x56,
	0xed, 0xf9, 0xa9, 0x6c, 0x83, 0x7e, 0x80, 0xd9, 0x8d, 0x5e, 0x52, 0x11, 0x12, 0xf9, 0x44, 0xb1,
	0xed, 0xf1, 0xb6, 0xce, 0x6e, 0xb2, 0x1e, 0xeb, 0xad, 0x28, 0xfa, 0x0a, 0x41, 0x9f, 0x1c, 0x18,
	0xe3, 0x8b, 0xfe, 0xee, 0x3d, 0xdc, 0xf9, 0x20, 0xce, 0x64, 0x2f, 0x3d, 0xf9, 0x85, 0x8f, 0xfe,
	0x05, 0x00, 0x00, 0xff, 0xff, 0x08, 0x79, 0x71, 0xa3, 0xd3, 0x03, 0x00, 0x00,
}
