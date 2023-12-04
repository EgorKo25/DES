// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.24.3
// source: proto/proto/des.proto

package extension_service_gen

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

const (
	UserExtensionService_GetUserExtension_FullMethodName = "/proto.UserExtensionService/GetUserExtension"
)

// UserExtensionServiceClient is the client API for UserExtensionService proto.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type UserExtensionServiceClient interface {
	GetUserExtension(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*GetResponse, error)
}

type userExtensionServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewUserExtensionServiceClient(cc grpc.ClientConnInterface) UserExtensionServiceClient {
	return &userExtensionServiceClient{cc}
}

func (c *userExtensionServiceClient) GetUserExtension(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*GetResponse, error) {
	out := new(GetResponse)
	err := c.cc.Invoke(ctx, UserExtensionService_GetUserExtension_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// UserExtensionServiceServer is the server API for UserExtensionService proto.
// All implementations must embed UnimplementedUserExtensionServiceServer
// for forward compatibility
type UserExtensionServiceServer interface {
	GetUserExtension(context.Context, *GetRequest) (*GetResponse, error)
	mustEmbedUnimplementedUserExtensionServiceServer()
}

// UnimplementedUserExtensionServiceServer must be embedded to have forward compatible implementations.
type UnimplementedUserExtensionServiceServer struct {
}

func (UnimplementedUserExtensionServiceServer) GetUserExtension(context.Context, *GetRequest) (*GetResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUserExtension not implemented")
}
func (UnimplementedUserExtensionServiceServer) mustEmbedUnimplementedUserExtensionServiceServer() {}

// UnsafeUserExtensionServiceServer may be embedded to opt out of forward compatibility for this proto.
// Use of this interface is not recommended, as added methods to UserExtensionServiceServer will
// result in compilation errors.
type UnsafeUserExtensionServiceServer interface {
	mustEmbedUnimplementedUserExtensionServiceServer()
}

func RegisterUserExtensionServiceServer(s grpc.ServiceRegistrar, srv UserExtensionServiceServer) {
	s.RegisterService(&UserExtensionService_ServiceDesc, srv)
}

func _UserExtensionService_GetUserExtension_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserExtensionServiceServer).GetUserExtension(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserExtensionService_GetUserExtension_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserExtensionServiceServer).GetUserExtension(ctx, req.(*GetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// UserExtensionService_ServiceDesc is the grpc.ServiceDesc for UserExtensionService proto.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var UserExtensionService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.UserExtensionService",
	HandlerType: (*UserExtensionServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetUserExtension",
			Handler:    _UserExtensionService_GetUserExtension_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/proto/des.proto",
}
