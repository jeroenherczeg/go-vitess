// Code generated by protoc-gen-go.
// source: replicationdata.proto
// DO NOT EDIT!

/*
Package replicationdata is a generated protocol buffer package.

It is generated from these files:
	replicationdata.proto

It has these top-level messages:
	Status
*/
package replicationdata

import proto "github.com/golang/protobuf/proto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal

// Status is the replication status for MySQL (returned by 'show slave status'
// and parsed into a Position and fields).
type Status struct {
	Position            string `protobuf:"bytes,1,opt,name=position" json:"position,omitempty"`
	SlaveIoRunning      bool   `protobuf:"varint,2,opt,name=slave_io_running" json:"slave_io_running,omitempty"`
	SlaveSqlRunning     bool   `protobuf:"varint,3,opt,name=slave_sql_running" json:"slave_sql_running,omitempty"`
	SecondsBehindMaster uint32 `protobuf:"varint,4,opt,name=seconds_behind_master" json:"seconds_behind_master,omitempty"`
	MasterHost          string `protobuf:"bytes,5,opt,name=master_host" json:"master_host,omitempty"`
	MasterPort          int32  `protobuf:"varint,6,opt,name=master_port" json:"master_port,omitempty"`
	MasterConnectRetry  int32  `protobuf:"varint,7,opt,name=master_connect_retry" json:"master_connect_retry,omitempty"`
}

func (m *Status) Reset()         { *m = Status{} }
func (m *Status) String() string { return proto.CompactTextString(m) }
func (*Status) ProtoMessage()    {}
