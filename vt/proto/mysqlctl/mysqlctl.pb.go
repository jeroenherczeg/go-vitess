// Code generated by protoc-gen-go.
// source: mysqlctl.proto
// DO NOT EDIT!

/*
Package mysqlctl is a generated protocol buffer package.

It is generated from these files:
	mysqlctl.proto

It has these top-level messages:
	StartRequest
	StartResponse
	ShutdownRequest
	ShutdownResponse
	RunMysqlUpgradeRequest
	RunMysqlUpgradeResponse
*/
package mysqlctl

import proto "github.com/golang/protobuf/proto"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal

type StartRequest struct {
}

func (m *StartRequest) Reset()         { *m = StartRequest{} }
func (m *StartRequest) String() string { return proto.CompactTextString(m) }
func (*StartRequest) ProtoMessage()    {}

type StartResponse struct {
}

func (m *StartResponse) Reset()         { *m = StartResponse{} }
func (m *StartResponse) String() string { return proto.CompactTextString(m) }
func (*StartResponse) ProtoMessage()    {}

type ShutdownRequest struct {
	WaitForMysqld bool `protobuf:"varint,1,opt,name=wait_for_mysqld" json:"wait_for_mysqld,omitempty"`
}

func (m *ShutdownRequest) Reset()         { *m = ShutdownRequest{} }
func (m *ShutdownRequest) String() string { return proto.CompactTextString(m) }
func (*ShutdownRequest) ProtoMessage()    {}

type ShutdownResponse struct {
}

func (m *ShutdownResponse) Reset()         { *m = ShutdownResponse{} }
func (m *ShutdownResponse) String() string { return proto.CompactTextString(m) }
func (*ShutdownResponse) ProtoMessage()    {}

type RunMysqlUpgradeRequest struct {
}

func (m *RunMysqlUpgradeRequest) Reset()         { *m = RunMysqlUpgradeRequest{} }
func (m *RunMysqlUpgradeRequest) String() string { return proto.CompactTextString(m) }
func (*RunMysqlUpgradeRequest) ProtoMessage()    {}

type RunMysqlUpgradeResponse struct {
}

func (m *RunMysqlUpgradeResponse) Reset()         { *m = RunMysqlUpgradeResponse{} }
func (m *RunMysqlUpgradeResponse) String() string { return proto.CompactTextString(m) }
func (*RunMysqlUpgradeResponse) ProtoMessage()    {}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// Client API for MysqlCtl service

type MysqlCtlClient interface {
	Start(ctx context.Context, in *StartRequest, opts ...grpc.CallOption) (*StartResponse, error)
	Shutdown(ctx context.Context, in *ShutdownRequest, opts ...grpc.CallOption) (*ShutdownResponse, error)
	RunMysqlUpgrade(ctx context.Context, in *RunMysqlUpgradeRequest, opts ...grpc.CallOption) (*RunMysqlUpgradeResponse, error)
}

type mysqlCtlClient struct {
	cc *grpc.ClientConn
}

func NewMysqlCtlClient(cc *grpc.ClientConn) MysqlCtlClient {
	return &mysqlCtlClient{cc}
}

func (c *mysqlCtlClient) Start(ctx context.Context, in *StartRequest, opts ...grpc.CallOption) (*StartResponse, error) {
	out := new(StartResponse)
	err := grpc.Invoke(ctx, "/mysqlctl.MysqlCtl/Start", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mysqlCtlClient) Shutdown(ctx context.Context, in *ShutdownRequest, opts ...grpc.CallOption) (*ShutdownResponse, error) {
	out := new(ShutdownResponse)
	err := grpc.Invoke(ctx, "/mysqlctl.MysqlCtl/Shutdown", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mysqlCtlClient) RunMysqlUpgrade(ctx context.Context, in *RunMysqlUpgradeRequest, opts ...grpc.CallOption) (*RunMysqlUpgradeResponse, error) {
	out := new(RunMysqlUpgradeResponse)
	err := grpc.Invoke(ctx, "/mysqlctl.MysqlCtl/RunMysqlUpgrade", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for MysqlCtl service

type MysqlCtlServer interface {
	Start(context.Context, *StartRequest) (*StartResponse, error)
	Shutdown(context.Context, *ShutdownRequest) (*ShutdownResponse, error)
	RunMysqlUpgrade(context.Context, *RunMysqlUpgradeRequest) (*RunMysqlUpgradeResponse, error)
}

func RegisterMysqlCtlServer(s *grpc.Server, srv MysqlCtlServer) {
	s.RegisterService(&_MysqlCtl_serviceDesc, srv)
}

func _MysqlCtl_Start_Handler(srv interface{}, ctx context.Context, codec grpc.Codec, buf []byte) (interface{}, error) {
	in := new(StartRequest)
	if err := codec.Unmarshal(buf, in); err != nil {
		return nil, err
	}
	out, err := srv.(MysqlCtlServer).Start(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func _MysqlCtl_Shutdown_Handler(srv interface{}, ctx context.Context, codec grpc.Codec, buf []byte) (interface{}, error) {
	in := new(ShutdownRequest)
	if err := codec.Unmarshal(buf, in); err != nil {
		return nil, err
	}
	out, err := srv.(MysqlCtlServer).Shutdown(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func _MysqlCtl_RunMysqlUpgrade_Handler(srv interface{}, ctx context.Context, codec grpc.Codec, buf []byte) (interface{}, error) {
	in := new(RunMysqlUpgradeRequest)
	if err := codec.Unmarshal(buf, in); err != nil {
		return nil, err
	}
	out, err := srv.(MysqlCtlServer).RunMysqlUpgrade(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

var _MysqlCtl_serviceDesc = grpc.ServiceDesc{
	ServiceName: "mysqlctl.MysqlCtl",
	HandlerType: (*MysqlCtlServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Start",
			Handler:    _MysqlCtl_Start_Handler,
		},
		{
			MethodName: "Shutdown",
			Handler:    _MysqlCtl_Shutdown_Handler,
		},
		{
			MethodName: "RunMysqlUpgrade",
			Handler:    _MysqlCtl_RunMysqlUpgrade_Handler,
		},
	},
	Streams: []grpc.StreamDesc{},
}
