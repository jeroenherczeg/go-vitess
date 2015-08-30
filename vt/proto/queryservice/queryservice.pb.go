// Code generated by protoc-gen-go.
// source: queryservice.proto
// DO NOT EDIT!

/*
Package queryservice is a generated protocol buffer package.

It is generated from these files:
	queryservice.proto

It has these top-level messages:
*/
package queryservice

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import query "github.com/youtube/vitess/go/vt/proto/query"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// Client API for Query service

type QueryClient interface {
	// GetSessionId gets a session id from the server. This call is being
	// deprecated in favor of using the Target field of the subsequent
	// queries, but is still here for backward compatibility.
	GetSessionId(ctx context.Context, in *query.GetSessionIdRequest, opts ...grpc.CallOption) (*query.GetSessionIdResponse, error)
	// Execute executes the specified SQL query (might be in a
	// transaction context, if Query.transaction_id is set).
	Execute(ctx context.Context, in *query.ExecuteRequest, opts ...grpc.CallOption) (*query.ExecuteResponse, error)
	// ExecuteBatch executes a list of queries, and returns the result
	// for each query.
	ExecuteBatch(ctx context.Context, in *query.ExecuteBatchRequest, opts ...grpc.CallOption) (*query.ExecuteBatchResponse, error)
	// StreamExecute executes a streaming query. Use this method if the
	// query returns a large number of rows. The first QueryResult will
	// contain the Fields, subsequent QueryResult messages will contain
	// the rows.
	StreamExecute(ctx context.Context, in *query.StreamExecuteRequest, opts ...grpc.CallOption) (Query_StreamExecuteClient, error)
	// Begin a transaction.
	Begin(ctx context.Context, in *query.BeginRequest, opts ...grpc.CallOption) (*query.BeginResponse, error)
	// Commit a transaction.
	Commit(ctx context.Context, in *query.CommitRequest, opts ...grpc.CallOption) (*query.CommitResponse, error)
	// Rollback a transaction.
	Rollback(ctx context.Context, in *query.RollbackRequest, opts ...grpc.CallOption) (*query.RollbackResponse, error)
	// SplitQuery is the API to facilitate MapReduce-type iterations
	// over large data sets (like full table dumps).
	SplitQuery(ctx context.Context, in *query.SplitQueryRequest, opts ...grpc.CallOption) (*query.SplitQueryResponse, error)
	// StreamHealth runs a streaming RPC to the tablet, that returns the
	// current health of the tablet on a regular basis.
	StreamHealth(ctx context.Context, in *query.StreamHealthRequest, opts ...grpc.CallOption) (Query_StreamHealthClient, error)
}

type queryClient struct {
	cc *grpc.ClientConn
}

func NewQueryClient(cc *grpc.ClientConn) QueryClient {
	return &queryClient{cc}
}

func (c *queryClient) GetSessionId(ctx context.Context, in *query.GetSessionIdRequest, opts ...grpc.CallOption) (*query.GetSessionIdResponse, error) {
	out := new(query.GetSessionIdResponse)
	err := grpc.Invoke(ctx, "/queryservice.Query/GetSessionId", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) Execute(ctx context.Context, in *query.ExecuteRequest, opts ...grpc.CallOption) (*query.ExecuteResponse, error) {
	out := new(query.ExecuteResponse)
	err := grpc.Invoke(ctx, "/queryservice.Query/Execute", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) ExecuteBatch(ctx context.Context, in *query.ExecuteBatchRequest, opts ...grpc.CallOption) (*query.ExecuteBatchResponse, error) {
	out := new(query.ExecuteBatchResponse)
	err := grpc.Invoke(ctx, "/queryservice.Query/ExecuteBatch", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) StreamExecute(ctx context.Context, in *query.StreamExecuteRequest, opts ...grpc.CallOption) (Query_StreamExecuteClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_Query_serviceDesc.Streams[0], c.cc, "/queryservice.Query/StreamExecute", opts...)
	if err != nil {
		return nil, err
	}
	x := &queryStreamExecuteClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Query_StreamExecuteClient interface {
	Recv() (*query.StreamExecuteResponse, error)
	grpc.ClientStream
}

type queryStreamExecuteClient struct {
	grpc.ClientStream
}

func (x *queryStreamExecuteClient) Recv() (*query.StreamExecuteResponse, error) {
	m := new(query.StreamExecuteResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *queryClient) Begin(ctx context.Context, in *query.BeginRequest, opts ...grpc.CallOption) (*query.BeginResponse, error) {
	out := new(query.BeginResponse)
	err := grpc.Invoke(ctx, "/queryservice.Query/Begin", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) Commit(ctx context.Context, in *query.CommitRequest, opts ...grpc.CallOption) (*query.CommitResponse, error) {
	out := new(query.CommitResponse)
	err := grpc.Invoke(ctx, "/queryservice.Query/Commit", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) Rollback(ctx context.Context, in *query.RollbackRequest, opts ...grpc.CallOption) (*query.RollbackResponse, error) {
	out := new(query.RollbackResponse)
	err := grpc.Invoke(ctx, "/queryservice.Query/Rollback", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) SplitQuery(ctx context.Context, in *query.SplitQueryRequest, opts ...grpc.CallOption) (*query.SplitQueryResponse, error) {
	out := new(query.SplitQueryResponse)
	err := grpc.Invoke(ctx, "/queryservice.Query/SplitQuery", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) StreamHealth(ctx context.Context, in *query.StreamHealthRequest, opts ...grpc.CallOption) (Query_StreamHealthClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_Query_serviceDesc.Streams[1], c.cc, "/queryservice.Query/StreamHealth", opts...)
	if err != nil {
		return nil, err
	}
	x := &queryStreamHealthClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Query_StreamHealthClient interface {
	Recv() (*query.StreamHealthResponse, error)
	grpc.ClientStream
}

type queryStreamHealthClient struct {
	grpc.ClientStream
}

func (x *queryStreamHealthClient) Recv() (*query.StreamHealthResponse, error) {
	m := new(query.StreamHealthResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for Query service

type QueryServer interface {
	// GetSessionId gets a session id from the server. This call is being
	// deprecated in favor of using the Target field of the subsequent
	// queries, but is still here for backward compatibility.
	GetSessionId(context.Context, *query.GetSessionIdRequest) (*query.GetSessionIdResponse, error)
	// Execute executes the specified SQL query (might be in a
	// transaction context, if Query.transaction_id is set).
	Execute(context.Context, *query.ExecuteRequest) (*query.ExecuteResponse, error)
	// ExecuteBatch executes a list of queries, and returns the result
	// for each query.
	ExecuteBatch(context.Context, *query.ExecuteBatchRequest) (*query.ExecuteBatchResponse, error)
	// StreamExecute executes a streaming query. Use this method if the
	// query returns a large number of rows. The first QueryResult will
	// contain the Fields, subsequent QueryResult messages will contain
	// the rows.
	StreamExecute(*query.StreamExecuteRequest, Query_StreamExecuteServer) error
	// Begin a transaction.
	Begin(context.Context, *query.BeginRequest) (*query.BeginResponse, error)
	// Commit a transaction.
	Commit(context.Context, *query.CommitRequest) (*query.CommitResponse, error)
	// Rollback a transaction.
	Rollback(context.Context, *query.RollbackRequest) (*query.RollbackResponse, error)
	// SplitQuery is the API to facilitate MapReduce-type iterations
	// over large data sets (like full table dumps).
	SplitQuery(context.Context, *query.SplitQueryRequest) (*query.SplitQueryResponse, error)
	// StreamHealth runs a streaming RPC to the tablet, that returns the
	// current health of the tablet on a regular basis.
	StreamHealth(*query.StreamHealthRequest, Query_StreamHealthServer) error
}

func RegisterQueryServer(s *grpc.Server, srv QueryServer) {
	s.RegisterService(&_Query_serviceDesc, srv)
}

func _Query_GetSessionId_Handler(srv interface{}, ctx context.Context, codec grpc.Codec, buf []byte) (interface{}, error) {
	in := new(query.GetSessionIdRequest)
	if err := codec.Unmarshal(buf, in); err != nil {
		return nil, err
	}
	out, err := srv.(QueryServer).GetSessionId(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func _Query_Execute_Handler(srv interface{}, ctx context.Context, codec grpc.Codec, buf []byte) (interface{}, error) {
	in := new(query.ExecuteRequest)
	if err := codec.Unmarshal(buf, in); err != nil {
		return nil, err
	}
	out, err := srv.(QueryServer).Execute(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func _Query_ExecuteBatch_Handler(srv interface{}, ctx context.Context, codec grpc.Codec, buf []byte) (interface{}, error) {
	in := new(query.ExecuteBatchRequest)
	if err := codec.Unmarshal(buf, in); err != nil {
		return nil, err
	}
	out, err := srv.(QueryServer).ExecuteBatch(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func _Query_StreamExecute_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(query.StreamExecuteRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(QueryServer).StreamExecute(m, &queryStreamExecuteServer{stream})
}

type Query_StreamExecuteServer interface {
	Send(*query.StreamExecuteResponse) error
	grpc.ServerStream
}

type queryStreamExecuteServer struct {
	grpc.ServerStream
}

func (x *queryStreamExecuteServer) Send(m *query.StreamExecuteResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _Query_Begin_Handler(srv interface{}, ctx context.Context, codec grpc.Codec, buf []byte) (interface{}, error) {
	in := new(query.BeginRequest)
	if err := codec.Unmarshal(buf, in); err != nil {
		return nil, err
	}
	out, err := srv.(QueryServer).Begin(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func _Query_Commit_Handler(srv interface{}, ctx context.Context, codec grpc.Codec, buf []byte) (interface{}, error) {
	in := new(query.CommitRequest)
	if err := codec.Unmarshal(buf, in); err != nil {
		return nil, err
	}
	out, err := srv.(QueryServer).Commit(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func _Query_Rollback_Handler(srv interface{}, ctx context.Context, codec grpc.Codec, buf []byte) (interface{}, error) {
	in := new(query.RollbackRequest)
	if err := codec.Unmarshal(buf, in); err != nil {
		return nil, err
	}
	out, err := srv.(QueryServer).Rollback(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func _Query_SplitQuery_Handler(srv interface{}, ctx context.Context, codec grpc.Codec, buf []byte) (interface{}, error) {
	in := new(query.SplitQueryRequest)
	if err := codec.Unmarshal(buf, in); err != nil {
		return nil, err
	}
	out, err := srv.(QueryServer).SplitQuery(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func _Query_StreamHealth_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(query.StreamHealthRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(QueryServer).StreamHealth(m, &queryStreamHealthServer{stream})
}

type Query_StreamHealthServer interface {
	Send(*query.StreamHealthResponse) error
	grpc.ServerStream
}

type queryStreamHealthServer struct {
	grpc.ServerStream
}

func (x *queryStreamHealthServer) Send(m *query.StreamHealthResponse) error {
	return x.ServerStream.SendMsg(m)
}

var _Query_serviceDesc = grpc.ServiceDesc{
	ServiceName: "queryservice.Query",
	HandlerType: (*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetSessionId",
			Handler:    _Query_GetSessionId_Handler,
		},
		{
			MethodName: "Execute",
			Handler:    _Query_Execute_Handler,
		},
		{
			MethodName: "ExecuteBatch",
			Handler:    _Query_ExecuteBatch_Handler,
		},
		{
			MethodName: "Begin",
			Handler:    _Query_Begin_Handler,
		},
		{
			MethodName: "Commit",
			Handler:    _Query_Commit_Handler,
		},
		{
			MethodName: "Rollback",
			Handler:    _Query_Rollback_Handler,
		},
		{
			MethodName: "SplitQuery",
			Handler:    _Query_SplitQuery_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StreamExecute",
			Handler:       _Query_StreamExecute_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "StreamHealth",
			Handler:       _Query_StreamHealth_Handler,
			ServerStreams: true,
		},
	},
}
