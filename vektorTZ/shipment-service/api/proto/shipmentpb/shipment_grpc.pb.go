package shipmentpb

import (
	context "context"

	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

const _ = grpc.SupportPackageIsVersion9

const (
	ShipmentService_CreateShipment_FullMethodName     = "/shipment.v1.ShipmentService/CreateShipment"
	ShipmentService_GetShipment_FullMethodName        = "/shipment.v1.ShipmentService/GetShipment"
	ShipmentService_AddShipmentEvent_FullMethodName   = "/shipment.v1.ShipmentService/AddShipmentEvent"
	ShipmentService_ListShipmentEvents_FullMethodName = "/shipment.v1.ShipmentService/ListShipmentEvents"
)

type ShipmentServiceClient interface {
	CreateShipment(ctx context.Context, in *CreateShipmentRequest, opts ...grpc.CallOption) (*CreateShipmentResponse, error)
	GetShipment(ctx context.Context, in *GetShipmentRequest, opts ...grpc.CallOption) (*GetShipmentResponse, error)
	AddShipmentEvent(ctx context.Context, in *AddShipmentEventRequest, opts ...grpc.CallOption) (*AddShipmentEventResponse, error)
	ListShipmentEvents(ctx context.Context, in *ListShipmentEventsRequest, opts ...grpc.CallOption) (*ListShipmentEventsResponse, error)
}

type shipmentServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewShipmentServiceClient(cc grpc.ClientConnInterface) ShipmentServiceClient {
	return &shipmentServiceClient{cc}
}

func (c *shipmentServiceClient) CreateShipment(ctx context.Context, in *CreateShipmentRequest, opts ...grpc.CallOption) (*CreateShipmentResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CreateShipmentResponse)
	err := c.cc.Invoke(ctx, ShipmentService_CreateShipment_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shipmentServiceClient) GetShipment(ctx context.Context, in *GetShipmentRequest, opts ...grpc.CallOption) (*GetShipmentResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetShipmentResponse)
	err := c.cc.Invoke(ctx, ShipmentService_GetShipment_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shipmentServiceClient) AddShipmentEvent(ctx context.Context, in *AddShipmentEventRequest, opts ...grpc.CallOption) (*AddShipmentEventResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(AddShipmentEventResponse)
	err := c.cc.Invoke(ctx, ShipmentService_AddShipmentEvent_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shipmentServiceClient) ListShipmentEvents(ctx context.Context, in *ListShipmentEventsRequest, opts ...grpc.CallOption) (*ListShipmentEventsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ListShipmentEventsResponse)
	err := c.cc.Invoke(ctx, ShipmentService_ListShipmentEvents_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type ShipmentServiceServer interface {
	CreateShipment(context.Context, *CreateShipmentRequest) (*CreateShipmentResponse, error)
	GetShipment(context.Context, *GetShipmentRequest) (*GetShipmentResponse, error)
	AddShipmentEvent(context.Context, *AddShipmentEventRequest) (*AddShipmentEventResponse, error)
	ListShipmentEvents(context.Context, *ListShipmentEventsRequest) (*ListShipmentEventsResponse, error)
	mustEmbedUnimplementedShipmentServiceServer()
}

type UnimplementedShipmentServiceServer struct{}

func (UnimplementedShipmentServiceServer) CreateShipment(context.Context, *CreateShipmentRequest) (*CreateShipmentResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method CreateShipment not implemented")
}
func (UnimplementedShipmentServiceServer) GetShipment(context.Context, *GetShipmentRequest) (*GetShipmentResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method GetShipment not implemented")
}
func (UnimplementedShipmentServiceServer) AddShipmentEvent(context.Context, *AddShipmentEventRequest) (*AddShipmentEventResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method AddShipmentEvent not implemented")
}
func (UnimplementedShipmentServiceServer) ListShipmentEvents(context.Context, *ListShipmentEventsRequest) (*ListShipmentEventsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method ListShipmentEvents not implemented")
}
func (UnimplementedShipmentServiceServer) mustEmbedUnimplementedShipmentServiceServer() {}
func (UnimplementedShipmentServiceServer) testEmbeddedByValue()                         {}

type UnsafeShipmentServiceServer interface {
	mustEmbedUnimplementedShipmentServiceServer()
}

func RegisterShipmentServiceServer(s grpc.ServiceRegistrar, srv ShipmentServiceServer) {

	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&ShipmentService_ServiceDesc, srv)
}

func _ShipmentService_CreateShipment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateShipmentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShipmentServiceServer).CreateShipment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ShipmentService_CreateShipment_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShipmentServiceServer).CreateShipment(ctx, req.(*CreateShipmentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ShipmentService_GetShipment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetShipmentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShipmentServiceServer).GetShipment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ShipmentService_GetShipment_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShipmentServiceServer).GetShipment(ctx, req.(*GetShipmentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ShipmentService_AddShipmentEvent_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddShipmentEventRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShipmentServiceServer).AddShipmentEvent(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ShipmentService_AddShipmentEvent_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShipmentServiceServer).AddShipmentEvent(ctx, req.(*AddShipmentEventRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ShipmentService_ListShipmentEvents_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListShipmentEventsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShipmentServiceServer).ListShipmentEvents(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ShipmentService_ListShipmentEvents_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShipmentServiceServer).ListShipmentEvents(ctx, req.(*ListShipmentEventsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var ShipmentService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "shipment.v1.ShipmentService",
	HandlerType: (*ShipmentServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateShipment",
			Handler:    _ShipmentService_CreateShipment_Handler,
		},
		{
			MethodName: "GetShipment",
			Handler:    _ShipmentService_GetShipment_Handler,
		},
		{
			MethodName: "AddShipmentEvent",
			Handler:    _ShipmentService_AddShipmentEvent_Handler,
		},
		{
			MethodName: "ListShipmentEvents",
			Handler:    _ShipmentService_ListShipmentEvents_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/proto/shipment.proto",
}
