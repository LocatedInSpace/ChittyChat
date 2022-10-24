// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

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

// ChittyChatClient is the client API for ChittyChat service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ChittyChatClient interface {
	Join(ctx context.Context, in *Information, opts ...grpc.CallOption) (ChittyChat_JoinClient, error)
	Publish(ctx context.Context, opts ...grpc.CallOption) (ChittyChat_PublishClient, error)
}

type chittyChatClient struct {
	cc grpc.ClientConnInterface
}

func NewChittyChatClient(cc grpc.ClientConnInterface) ChittyChatClient {
	return &chittyChatClient{cc}
}

func (c *chittyChatClient) Join(ctx context.Context, in *Information, opts ...grpc.CallOption) (ChittyChat_JoinClient, error) {
	stream, err := c.cc.NewStream(ctx, &ChittyChat_ServiceDesc.Streams[0], "/proto.ChittyChat/Join", opts...)
	if err != nil {
		return nil, err
	}
	x := &chittyChatJoinClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type ChittyChat_JoinClient interface {
	Recv() (*StatusChange, error)
	grpc.ClientStream
}

type chittyChatJoinClient struct {
	grpc.ClientStream
}

func (x *chittyChatJoinClient) Recv() (*StatusChange, error) {
	m := new(StatusChange)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *chittyChatClient) Publish(ctx context.Context, opts ...grpc.CallOption) (ChittyChat_PublishClient, error) {
	stream, err := c.cc.NewStream(ctx, &ChittyChat_ServiceDesc.Streams[1], "/proto.ChittyChat/Publish", opts...)
	if err != nil {
		return nil, err
	}
	x := &chittyChatPublishClient{stream}
	return x, nil
}

type ChittyChat_PublishClient interface {
	Send(*MessageSent) error
	Recv() (*MessageRecv, error)
	grpc.ClientStream
}

type chittyChatPublishClient struct {
	grpc.ClientStream
}

func (x *chittyChatPublishClient) Send(m *MessageSent) error {
	return x.ClientStream.SendMsg(m)
}

func (x *chittyChatPublishClient) Recv() (*MessageRecv, error) {
	m := new(MessageRecv)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ChittyChatServer is the server API for ChittyChat service.
// All implementations must embed UnimplementedChittyChatServer
// for forward compatibility
type ChittyChatServer interface {
	Join(*Information, ChittyChat_JoinServer) error
	Publish(ChittyChat_PublishServer) error
	mustEmbedUnimplementedChittyChatServer()
}

// UnimplementedChittyChatServer must be embedded to have forward compatible implementations.
type UnimplementedChittyChatServer struct {
}

func (UnimplementedChittyChatServer) Join(*Information, ChittyChat_JoinServer) error {
	return status.Errorf(codes.Unimplemented, "method Join not implemented")
}
func (UnimplementedChittyChatServer) Publish(ChittyChat_PublishServer) error {
	return status.Errorf(codes.Unimplemented, "method Publish not implemented")
}
func (UnimplementedChittyChatServer) mustEmbedUnimplementedChittyChatServer() {}

// UnsafeChittyChatServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ChittyChatServer will
// result in compilation errors.
type UnsafeChittyChatServer interface {
	mustEmbedUnimplementedChittyChatServer()
}

func RegisterChittyChatServer(s grpc.ServiceRegistrar, srv ChittyChatServer) {
	s.RegisterService(&ChittyChat_ServiceDesc, srv)
}

func _ChittyChat_Join_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(Information)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ChittyChatServer).Join(m, &chittyChatJoinServer{stream})
}

type ChittyChat_JoinServer interface {
	Send(*StatusChange) error
	grpc.ServerStream
}

type chittyChatJoinServer struct {
	grpc.ServerStream
}

func (x *chittyChatJoinServer) Send(m *StatusChange) error {
	return x.ServerStream.SendMsg(m)
}

func _ChittyChat_Publish_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ChittyChatServer).Publish(&chittyChatPublishServer{stream})
}

type ChittyChat_PublishServer interface {
	Send(*MessageRecv) error
	Recv() (*MessageSent, error)
	grpc.ServerStream
}

type chittyChatPublishServer struct {
	grpc.ServerStream
}

func (x *chittyChatPublishServer) Send(m *MessageRecv) error {
	return x.ServerStream.SendMsg(m)
}

func (x *chittyChatPublishServer) Recv() (*MessageSent, error) {
	m := new(MessageSent)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ChittyChat_ServiceDesc is the grpc.ServiceDesc for ChittyChat service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ChittyChat_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.ChittyChat",
	HandlerType: (*ChittyChatServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Join",
			Handler:       _ChittyChat_Join_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "Publish",
			Handler:       _ChittyChat_Publish_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "proto/chittychat.proto",
}