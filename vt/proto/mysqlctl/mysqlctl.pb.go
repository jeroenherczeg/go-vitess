// Code generated by protoc-gen-go. DO NOT EDIT.
// source: mysqlctl.proto

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
	ReinitConfigRequest
	ReinitConfigResponse
	RefreshConfigRequest
	RefreshConfigResponse
*/
package mysqlctl

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type StartRequest struct {
	MysqldArgs []string `protobuf:"bytes,1,rep,name=mysqld_args,json=mysqldArgs" json:"mysqld_args,omitempty"`
}

func (m *StartRequest) Reset()                    { *m = StartRequest{} }
func (m *StartRequest) String() string            { return proto.CompactTextString(m) }
func (*StartRequest) ProtoMessage()               {}
func (*StartRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *StartRequest) GetMysqldArgs() []string {
	if m != nil {
		return m.MysqldArgs
	}
	return nil
}

type StartResponse struct {
}

func (m *StartResponse) Reset()                    { *m = StartResponse{} }
func (m *StartResponse) String() string            { return proto.CompactTextString(m) }
func (*StartResponse) ProtoMessage()               {}
func (*StartResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

type ShutdownRequest struct {
	WaitForMysqld bool `protobuf:"varint,1,opt,name=wait_for_mysqld,json=waitForMysqld" json:"wait_for_mysqld,omitempty"`
}

func (m *ShutdownRequest) Reset()                    { *m = ShutdownRequest{} }
func (m *ShutdownRequest) String() string            { return proto.CompactTextString(m) }
func (*ShutdownRequest) ProtoMessage()               {}
func (*ShutdownRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *ShutdownRequest) GetWaitForMysqld() bool {
	if m != nil {
		return m.WaitForMysqld
	}
	return false
}

type ShutdownResponse struct {
}

func (m *ShutdownResponse) Reset()                    { *m = ShutdownResponse{} }
func (m *ShutdownResponse) String() string            { return proto.CompactTextString(m) }
func (*ShutdownResponse) ProtoMessage()               {}
func (*ShutdownResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

type RunMysqlUpgradeRequest struct {
}

func (m *RunMysqlUpgradeRequest) Reset()                    { *m = RunMysqlUpgradeRequest{} }
func (m *RunMysqlUpgradeRequest) String() string            { return proto.CompactTextString(m) }
func (*RunMysqlUpgradeRequest) ProtoMessage()               {}
func (*RunMysqlUpgradeRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

type RunMysqlUpgradeResponse struct {
}

func (m *RunMysqlUpgradeResponse) Reset()                    { *m = RunMysqlUpgradeResponse{} }
func (m *RunMysqlUpgradeResponse) String() string            { return proto.CompactTextString(m) }
func (*RunMysqlUpgradeResponse) ProtoMessage()               {}
func (*RunMysqlUpgradeResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

type ReinitConfigRequest struct {
}

func (m *ReinitConfigRequest) Reset()                    { *m = ReinitConfigRequest{} }
func (m *ReinitConfigRequest) String() string            { return proto.CompactTextString(m) }
func (*ReinitConfigRequest) ProtoMessage()               {}
func (*ReinitConfigRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

type ReinitConfigResponse struct {
}

func (m *ReinitConfigResponse) Reset()                    { *m = ReinitConfigResponse{} }
func (m *ReinitConfigResponse) String() string            { return proto.CompactTextString(m) }
func (*ReinitConfigResponse) ProtoMessage()               {}
func (*ReinitConfigResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

type RefreshConfigRequest struct {
}

func (m *RefreshConfigRequest) Reset()                    { *m = RefreshConfigRequest{} }
func (m *RefreshConfigRequest) String() string            { return proto.CompactTextString(m) }
func (*RefreshConfigRequest) ProtoMessage()               {}
func (*RefreshConfigRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

type RefreshConfigResponse struct {
}

func (m *RefreshConfigResponse) Reset()                    { *m = RefreshConfigResponse{} }
func (m *RefreshConfigResponse) String() string            { return proto.CompactTextString(m) }
func (*RefreshConfigResponse) ProtoMessage()               {}
func (*RefreshConfigResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{9} }

func init() {
	proto.RegisterType((*StartRequest)(nil), "mysqlctl.StartRequest")
	proto.RegisterType((*StartResponse)(nil), "mysqlctl.StartResponse")
	proto.RegisterType((*ShutdownRequest)(nil), "mysqlctl.ShutdownRequest")
	proto.RegisterType((*ShutdownResponse)(nil), "mysqlctl.ShutdownResponse")
	proto.RegisterType((*RunMysqlUpgradeRequest)(nil), "mysqlctl.RunMysqlUpgradeRequest")
	proto.RegisterType((*RunMysqlUpgradeResponse)(nil), "mysqlctl.RunMysqlUpgradeResponse")
	proto.RegisterType((*ReinitConfigRequest)(nil), "mysqlctl.ReinitConfigRequest")
	proto.RegisterType((*ReinitConfigResponse)(nil), "mysqlctl.ReinitConfigResponse")
	proto.RegisterType((*RefreshConfigRequest)(nil), "mysqlctl.RefreshConfigRequest")
	proto.RegisterType((*RefreshConfigResponse)(nil), "mysqlctl.RefreshConfigResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for MysqlCtl service

type MysqlCtlClient interface {
	Start(ctx context.Context, in *StartRequest, opts ...grpc.CallOption) (*StartResponse, error)
	Shutdown(ctx context.Context, in *ShutdownRequest, opts ...grpc.CallOption) (*ShutdownResponse, error)
	RunMysqlUpgrade(ctx context.Context, in *RunMysqlUpgradeRequest, opts ...grpc.CallOption) (*RunMysqlUpgradeResponse, error)
	ReinitConfig(ctx context.Context, in *ReinitConfigRequest, opts ...grpc.CallOption) (*ReinitConfigResponse, error)
	RefreshConfig(ctx context.Context, in *RefreshConfigRequest, opts ...grpc.CallOption) (*RefreshConfigResponse, error)
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

func (c *mysqlCtlClient) ReinitConfig(ctx context.Context, in *ReinitConfigRequest, opts ...grpc.CallOption) (*ReinitConfigResponse, error) {
	out := new(ReinitConfigResponse)
	err := grpc.Invoke(ctx, "/mysqlctl.MysqlCtl/ReinitConfig", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mysqlCtlClient) RefreshConfig(ctx context.Context, in *RefreshConfigRequest, opts ...grpc.CallOption) (*RefreshConfigResponse, error) {
	out := new(RefreshConfigResponse)
	err := grpc.Invoke(ctx, "/mysqlctl.MysqlCtl/RefreshConfig", in, out, c.cc, opts...)
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
	ReinitConfig(context.Context, *ReinitConfigRequest) (*ReinitConfigResponse, error)
	RefreshConfig(context.Context, *RefreshConfigRequest) (*RefreshConfigResponse, error)
}

func RegisterMysqlCtlServer(s *grpc.Server, srv MysqlCtlServer) {
	s.RegisterService(&_MysqlCtl_serviceDesc, srv)
}

func _MysqlCtl_Start_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StartRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MysqlCtlServer).Start(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/mysqlctl.MysqlCtl/Start",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MysqlCtlServer).Start(ctx, req.(*StartRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MysqlCtl_Shutdown_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ShutdownRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MysqlCtlServer).Shutdown(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/mysqlctl.MysqlCtl/Shutdown",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MysqlCtlServer).Shutdown(ctx, req.(*ShutdownRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MysqlCtl_RunMysqlUpgrade_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RunMysqlUpgradeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MysqlCtlServer).RunMysqlUpgrade(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/mysqlctl.MysqlCtl/RunMysqlUpgrade",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MysqlCtlServer).RunMysqlUpgrade(ctx, req.(*RunMysqlUpgradeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MysqlCtl_ReinitConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReinitConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MysqlCtlServer).ReinitConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/mysqlctl.MysqlCtl/ReinitConfig",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MysqlCtlServer).ReinitConfig(ctx, req.(*ReinitConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MysqlCtl_RefreshConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RefreshConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MysqlCtlServer).RefreshConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/mysqlctl.MysqlCtl/RefreshConfig",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MysqlCtlServer).RefreshConfig(ctx, req.(*RefreshConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
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
		{
			MethodName: "ReinitConfig",
			Handler:    _MysqlCtl_ReinitConfig_Handler,
		},
		{
			MethodName: "RefreshConfig",
			Handler:    _MysqlCtl_RefreshConfig_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "mysqlctl.proto",
}

func init() { proto.RegisterFile("mysqlctl.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 313 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x92, 0x4d, 0x4f, 0x83, 0x30,
	0x18, 0xc7, 0x5d, 0x16, 0x0d, 0x3e, 0x6e, 0x62, 0xaa, 0x1b, 0xac, 0x89, 0x0e, 0x39, 0x98, 0x9d,
	0x66, 0xa2, 0x27, 0xbd, 0x19, 0x12, 0x6f, 0xc6, 0xa4, 0x8b, 0x89, 0x37, 0x82, 0x52, 0x18, 0x09,
	0x52, 0xd6, 0x96, 0x2c, 0x7e, 0x05, 0x3f, 0xb5, 0xb1, 0x14, 0x06, 0x63, 0xf3, 0xc8, 0xff, 0xed,
	0x09, 0x3f, 0x80, 0xd3, 0xaf, 0x6f, 0xb1, 0x4a, 0x3f, 0x65, 0x3a, 0xcf, 0x39, 0x93, 0x0c, 0x19,
	0xd5, 0xb3, 0x7b, 0x0b, 0x83, 0x85, 0x0c, 0xb8, 0x24, 0x74, 0x55, 0x50, 0x21, 0xd1, 0x14, 0x4e,
	0x94, 0x17, 0xfa, 0x01, 0x8f, 0x85, 0xdd, 0x73, 0xfa, 0xb3, 0x63, 0x02, 0xa5, 0xf4, 0xc4, 0x63,
	0xe1, 0x9a, 0x30, 0xd4, 0x05, 0x91, 0xb3, 0x4c, 0x50, 0xf7, 0x01, 0xcc, 0xc5, 0xb2, 0x90, 0x21,
	0x5b, 0x67, 0xd5, 0xc8, 0x0d, 0x98, 0xeb, 0x20, 0x91, 0x7e, 0xc4, 0xb8, 0x5f, 0x56, 0xed, 0x9e,
	0xd3, 0x9b, 0x19, 0x64, 0xf8, 0x27, 0x3f, 0x33, 0xfe, 0xa2, 0x44, 0x17, 0xc1, 0xd9, 0xa6, 0xaa,
	0xe7, 0x6c, 0x18, 0x93, 0x22, 0x53, 0x81, 0xb7, 0x3c, 0xe6, 0x41, 0x48, 0xf5, 0xaa, 0x3b, 0x01,
	0xab, 0xe3, 0xe8, 0xd2, 0x08, 0xce, 0x09, 0x4d, 0xb2, 0x44, 0x7a, 0x2c, 0x8b, 0x92, 0xb8, 0x6a,
	0x8c, 0xe1, 0xa2, 0x2d, 0xeb, 0xb8, 0xd2, 0x23, 0x4e, 0xc5, 0xb2, 0x9d, 0xb7, 0x60, 0xb4, 0xa5,
	0x97, 0x85, 0xbb, 0x9f, 0x3e, 0x18, 0xea, 0xb0, 0x27, 0x53, 0xf4, 0x08, 0x87, 0x8a, 0x00, 0x1a,
	0xcf, 0x6b, 0xac, 0x4d, 0x86, 0xd8, 0xea, 0xe8, 0xfa, 0xee, 0x01, 0xf2, 0xc0, 0xa8, 0xde, 0x18,
	0x4d, 0x1a, 0xb1, 0x36, 0x40, 0x8c, 0x77, 0x59, 0xf5, 0xc8, 0x3b, 0x98, 0x5b, 0x20, 0x90, 0xb3,
	0x29, 0xec, 0xa6, 0x87, 0xaf, 0xff, 0x49, 0xd4, 0xcb, 0xaf, 0x30, 0x68, 0x02, 0x43, 0x97, 0x8d,
	0x52, 0x97, 0x2f, 0xbe, 0xda, 0x67, 0xd7, 0x83, 0x04, 0x86, 0x2d, 0xa2, 0xa8, 0x55, 0xe9, 0x7e,
	0x02, 0x3c, 0xdd, 0xeb, 0x57, 0x9b, 0x1f, 0x47, 0xea, 0x1f, 0xbe, 0xff, 0x0d, 0x00, 0x00, 0xff,
	0xff, 0x81, 0x96, 0x68, 0x13, 0xd5, 0x02, 0x00, 0x00,
}
