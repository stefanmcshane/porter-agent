// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.7
// source: logproto.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// PusherClient is the client API for Pusher service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type PusherClient interface {
	Push(ctx context.Context, in *PushRequest, opts ...grpc.CallOption) (*PushResponse, error)
}

type pusherClient struct {
	cc grpc.ClientConnInterface
}

func NewPusherClient(cc grpc.ClientConnInterface) PusherClient {
	return &pusherClient{cc}
}

func (c *pusherClient) Push(ctx context.Context, in *PushRequest, opts ...grpc.CallOption) (*PushResponse, error) {
	out := new(PushResponse)
	err := c.cc.Invoke(ctx, "/logproto.Pusher/Push", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PusherServer is the server API for Pusher service.
// All implementations must embed UnimplementedPusherServer
// for forward compatibility
type PusherServer interface {
	Push(context.Context, *PushRequest) (*PushResponse, error)
	mustEmbedUnimplementedPusherServer()
}

// UnimplementedPusherServer must be embedded to have forward compatible implementations.
type UnimplementedPusherServer struct {
}

func (UnimplementedPusherServer) Push(context.Context, *PushRequest) (*PushResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Push not implemented")
}
func (UnimplementedPusherServer) mustEmbedUnimplementedPusherServer() {}

// UnsafePusherServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to PusherServer will
// result in compilation errors.
type UnsafePusherServer interface {
	mustEmbedUnimplementedPusherServer()
}

func RegisterPusherServer(s grpc.ServiceRegistrar, srv PusherServer) {
	s.RegisterService(&Pusher_ServiceDesc, srv)
}

func _Pusher_Push_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PushRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PusherServer).Push(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/logproto.Pusher/Push",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PusherServer).Push(ctx, req.(*PushRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Pusher_ServiceDesc is the grpc.ServiceDesc for Pusher service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Pusher_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "logproto.Pusher",
	HandlerType: (*PusherServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Push",
			Handler:    _Pusher_Push_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "logproto.proto",
}

// QuerierClient is the client API for Querier service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type QuerierClient interface {
	Query(ctx context.Context, in *QueryRequest, opts ...grpc.CallOption) (Querier_QueryClient, error)
	Label(ctx context.Context, in *LabelRequest, opts ...grpc.CallOption) (*LabelResponse, error)
	Tail(ctx context.Context, in *TailRequest, opts ...grpc.CallOption) (Querier_TailClient, error)
	Series(ctx context.Context, in *SeriesRequest, opts ...grpc.CallOption) (*SeriesResponse, error)
	TailersCount(ctx context.Context, in *TailersCountRequest, opts ...grpc.CallOption) (*TailersCountResponse, error)
}

type querierClient struct {
	cc grpc.ClientConnInterface
}

func NewQuerierClient(cc grpc.ClientConnInterface) QuerierClient {
	return &querierClient{cc}
}

func (c *querierClient) Query(ctx context.Context, in *QueryRequest, opts ...grpc.CallOption) (Querier_QueryClient, error) {
	stream, err := c.cc.NewStream(ctx, &Querier_ServiceDesc.Streams[0], "/logproto.Querier/Query", opts...)
	if err != nil {
		return nil, err
	}
	x := &querierQueryClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Querier_QueryClient interface {
	Recv() (*QueryResponse, error)
	grpc.ClientStream
}

type querierQueryClient struct {
	grpc.ClientStream
}

func (x *querierQueryClient) Recv() (*QueryResponse, error) {
	m := new(QueryResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *querierClient) Label(ctx context.Context, in *LabelRequest, opts ...grpc.CallOption) (*LabelResponse, error) {
	out := new(LabelResponse)
	err := c.cc.Invoke(ctx, "/logproto.Querier/Label", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *querierClient) Tail(ctx context.Context, in *TailRequest, opts ...grpc.CallOption) (Querier_TailClient, error) {
	stream, err := c.cc.NewStream(ctx, &Querier_ServiceDesc.Streams[1], "/logproto.Querier/Tail", opts...)
	if err != nil {
		return nil, err
	}
	x := &querierTailClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Querier_TailClient interface {
	Recv() (*TailResponse, error)
	grpc.ClientStream
}

type querierTailClient struct {
	grpc.ClientStream
}

func (x *querierTailClient) Recv() (*TailResponse, error) {
	m := new(TailResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *querierClient) Series(ctx context.Context, in *SeriesRequest, opts ...grpc.CallOption) (*SeriesResponse, error) {
	out := new(SeriesResponse)
	err := c.cc.Invoke(ctx, "/logproto.Querier/Series", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *querierClient) TailersCount(ctx context.Context, in *TailersCountRequest, opts ...grpc.CallOption) (*TailersCountResponse, error) {
	out := new(TailersCountResponse)
	err := c.cc.Invoke(ctx, "/logproto.Querier/TailersCount", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QuerierServer is the server API for Querier service.
// All implementations must embed UnimplementedQuerierServer
// for forward compatibility
type QuerierServer interface {
	Query(*QueryRequest, Querier_QueryServer) error
	Label(context.Context, *LabelRequest) (*LabelResponse, error)
	Tail(*TailRequest, Querier_TailServer) error
	Series(context.Context, *SeriesRequest) (*SeriesResponse, error)
	TailersCount(context.Context, *TailersCountRequest) (*TailersCountResponse, error)
	mustEmbedUnimplementedQuerierServer()
}

// UnimplementedQuerierServer must be embedded to have forward compatible implementations.
type UnimplementedQuerierServer struct {
}

func (UnimplementedQuerierServer) Query(*QueryRequest, Querier_QueryServer) error {
	return status.Errorf(codes.Unimplemented, "method Query not implemented")
}
func (UnimplementedQuerierServer) Label(context.Context, *LabelRequest) (*LabelResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Label not implemented")
}
func (UnimplementedQuerierServer) Tail(*TailRequest, Querier_TailServer) error {
	return status.Errorf(codes.Unimplemented, "method Tail not implemented")
}
func (UnimplementedQuerierServer) Series(context.Context, *SeriesRequest) (*SeriesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Series not implemented")
}
func (UnimplementedQuerierServer) TailersCount(context.Context, *TailersCountRequest) (*TailersCountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TailersCount not implemented")
}
func (UnimplementedQuerierServer) mustEmbedUnimplementedQuerierServer() {}

// UnsafeQuerierServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to QuerierServer will
// result in compilation errors.
type UnsafeQuerierServer interface {
	mustEmbedUnimplementedQuerierServer()
}

func RegisterQuerierServer(s grpc.ServiceRegistrar, srv QuerierServer) {
	s.RegisterService(&Querier_ServiceDesc, srv)
}

func _Querier_Query_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(QueryRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(QuerierServer).Query(m, &querierQueryServer{stream})
}

type Querier_QueryServer interface {
	Send(*QueryResponse) error
	grpc.ServerStream
}

type querierQueryServer struct {
	grpc.ServerStream
}

func (x *querierQueryServer) Send(m *QueryResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _Querier_Label_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LabelRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QuerierServer).Label(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/logproto.Querier/Label",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QuerierServer).Label(ctx, req.(*LabelRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Querier_Tail_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(TailRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(QuerierServer).Tail(m, &querierTailServer{stream})
}

type Querier_TailServer interface {
	Send(*TailResponse) error
	grpc.ServerStream
}

type querierTailServer struct {
	grpc.ServerStream
}

func (x *querierTailServer) Send(m *TailResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _Querier_Series_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SeriesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QuerierServer).Series(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/logproto.Querier/Series",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QuerierServer).Series(ctx, req.(*SeriesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Querier_TailersCount_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TailersCountRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QuerierServer).TailersCount(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/logproto.Querier/TailersCount",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QuerierServer).TailersCount(ctx, req.(*TailersCountRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Querier_ServiceDesc is the grpc.ServiceDesc for Querier service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Querier_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "logproto.Querier",
	HandlerType: (*QuerierServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Label",
			Handler:    _Querier_Label_Handler,
		},
		{
			MethodName: "Series",
			Handler:    _Querier_Series_Handler,
		},
		{
			MethodName: "TailersCount",
			Handler:    _Querier_TailersCount_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Query",
			Handler:       _Querier_Query_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "Tail",
			Handler:       _Querier_Tail_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "logproto.proto",
}

// IngesterClient is the client API for Ingester service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type IngesterClient interface {
	TransferChunks(ctx context.Context, opts ...grpc.CallOption) (Ingester_TransferChunksClient, error)
}

type ingesterClient struct {
	cc grpc.ClientConnInterface
}

func NewIngesterClient(cc grpc.ClientConnInterface) IngesterClient {
	return &ingesterClient{cc}
}

func (c *ingesterClient) TransferChunks(ctx context.Context, opts ...grpc.CallOption) (Ingester_TransferChunksClient, error) {
	stream, err := c.cc.NewStream(ctx, &Ingester_ServiceDesc.Streams[0], "/logproto.Ingester/TransferChunks", opts...)
	if err != nil {
		return nil, err
	}
	x := &ingesterTransferChunksClient{stream}
	return x, nil
}

type Ingester_TransferChunksClient interface {
	Send(*TimeSeriesChunk) error
	CloseAndRecv() (*TransferChunksResponse, error)
	grpc.ClientStream
}

type ingesterTransferChunksClient struct {
	grpc.ClientStream
}

func (x *ingesterTransferChunksClient) Send(m *TimeSeriesChunk) error {
	return x.ClientStream.SendMsg(m)
}

func (x *ingesterTransferChunksClient) CloseAndRecv() (*TransferChunksResponse, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(TransferChunksResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// IngesterServer is the server API for Ingester service.
// All implementations must embed UnimplementedIngesterServer
// for forward compatibility
type IngesterServer interface {
	TransferChunks(Ingester_TransferChunksServer) error
	mustEmbedUnimplementedIngesterServer()
}

// UnimplementedIngesterServer must be embedded to have forward compatible implementations.
type UnimplementedIngesterServer struct {
}

func (UnimplementedIngesterServer) TransferChunks(Ingester_TransferChunksServer) error {
	return status.Errorf(codes.Unimplemented, "method TransferChunks not implemented")
}
func (UnimplementedIngesterServer) mustEmbedUnimplementedIngesterServer() {}

// UnsafeIngesterServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to IngesterServer will
// result in compilation errors.
type UnsafeIngesterServer interface {
	mustEmbedUnimplementedIngesterServer()
}

func RegisterIngesterServer(s grpc.ServiceRegistrar, srv IngesterServer) {
	s.RegisterService(&Ingester_ServiceDesc, srv)
}

func _Ingester_TransferChunks_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(IngesterServer).TransferChunks(&ingesterTransferChunksServer{stream})
}

type Ingester_TransferChunksServer interface {
	SendAndClose(*TransferChunksResponse) error
	Recv() (*TimeSeriesChunk, error)
	grpc.ServerStream
}

type ingesterTransferChunksServer struct {
	grpc.ServerStream
}

func (x *ingesterTransferChunksServer) SendAndClose(m *TransferChunksResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *ingesterTransferChunksServer) Recv() (*TimeSeriesChunk, error) {
	m := new(TimeSeriesChunk)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Ingester_ServiceDesc is the grpc.ServiceDesc for Ingester service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Ingester_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "logproto.Ingester",
	HandlerType: (*IngesterServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "TransferChunks",
			Handler:       _Ingester_TransferChunks_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "logproto.proto",
}