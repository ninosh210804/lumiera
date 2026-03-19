package shipmentpb

import (
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

const (
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Status int32

const (
	Status_STATUS_UNSPECIFIED Status = 0
	Status_STATUS_PENDING     Status = 1
	Status_STATUS_PICKED_UP   Status = 2
	Status_STATUS_IN_TRANSIT  Status = 3
	Status_STATUS_DELIVERED   Status = 4
	Status_STATUS_CANCELLED   Status = 5
)

var (
	Status_name = map[int32]string{
		0: "STATUS_UNSPECIFIED",
		1: "STATUS_PENDING",
		2: "STATUS_PICKED_UP",
		3: "STATUS_IN_TRANSIT",
		4: "STATUS_DELIVERED",
		5: "STATUS_CANCELLED",
	}
	Status_value = map[string]int32{
		"STATUS_UNSPECIFIED": 0,
		"STATUS_PENDING":     1,
		"STATUS_PICKED_UP":   2,
		"STATUS_IN_TRANSIT":  3,
		"STATUS_DELIVERED":   4,
		"STATUS_CANCELLED":   5,
	}
)

func (x Status) Enum() *Status {
	p := new(Status)
	*p = x
	return p
}

func (x Status) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Status) Descriptor() protoreflect.EnumDescriptor {
	return file_api_proto_shipment_proto_enumTypes[0].Descriptor()
}

func (Status) Type() protoreflect.EnumType {
	return &file_api_proto_shipment_proto_enumTypes[0]
}

func (x Status) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

func (Status) EnumDescriptor() ([]byte, []int) {
	return file_api_proto_shipment_proto_rawDescGZIP(), []int{0}
}

type Shipment struct {
	state           protoimpl.MessageState `protogen:"open.v1"`
	Id              string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	ReferenceNumber string                 `protobuf:"bytes,2,opt,name=reference_number,json=referenceNumber,proto3" json:"reference_number,omitempty"`
	Origin          string                 `protobuf:"bytes,3,opt,name=origin,proto3" json:"origin,omitempty"`
	Destination     string                 `protobuf:"bytes,4,opt,name=destination,proto3" json:"destination,omitempty"`
	DriverName      string                 `protobuf:"bytes,5,opt,name=driver_name,json=driverName,proto3" json:"driver_name,omitempty"`
	VehicleId       string                 `protobuf:"bytes,6,opt,name=vehicle_id,json=vehicleId,proto3" json:"vehicle_id,omitempty"`
	Amount          float64                `protobuf:"fixed64,7,opt,name=amount,proto3" json:"amount,omitempty"`
	DriverRevenue   float64                `protobuf:"fixed64,8,opt,name=driver_revenue,json=driverRevenue,proto3" json:"driver_revenue,omitempty"`
	CurrentStatus   Status                 `protobuf:"varint,9,opt,name=current_status,json=currentStatus,proto3,enum=shipment.v1.Status" json:"current_status,omitempty"`
	CreatedAt       *timestamppb.Timestamp `protobuf:"bytes,10,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	UpdatedAt       *timestamppb.Timestamp `protobuf:"bytes,11,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
	unknownFields   protoimpl.UnknownFields
	sizeCache       protoimpl.SizeCache
}

func (x *Shipment) Reset() {
	*x = Shipment{}
	mi := &file_api_proto_shipment_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Shipment) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Shipment) ProtoMessage() {}

func (x *Shipment) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_shipment_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*Shipment) Descriptor() ([]byte, []int) {
	return file_api_proto_shipment_proto_rawDescGZIP(), []int{0}
}

func (x *Shipment) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Shipment) GetReferenceNumber() string {
	if x != nil {
		return x.ReferenceNumber
	}
	return ""
}

func (x *Shipment) GetOrigin() string {
	if x != nil {
		return x.Origin
	}
	return ""
}

func (x *Shipment) GetDestination() string {
	if x != nil {
		return x.Destination
	}
	return ""
}

func (x *Shipment) GetDriverName() string {
	if x != nil {
		return x.DriverName
	}
	return ""
}

func (x *Shipment) GetVehicleId() string {
	if x != nil {
		return x.VehicleId
	}
	return ""
}

func (x *Shipment) GetAmount() float64 {
	if x != nil {
		return x.Amount
	}
	return 0
}

func (x *Shipment) GetDriverRevenue() float64 {
	if x != nil {
		return x.DriverRevenue
	}
	return 0
}

func (x *Shipment) GetCurrentStatus() Status {
	if x != nil {
		return x.CurrentStatus
	}
	return Status_STATUS_UNSPECIFIED
}

func (x *Shipment) GetCreatedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.CreatedAt
	}
	return nil
}

func (x *Shipment) GetUpdatedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.UpdatedAt
	}
	return nil
}

type ShipmentEvent struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	ShipmentId    string                 `protobuf:"bytes,2,opt,name=shipment_id,json=shipmentId,proto3" json:"shipment_id,omitempty"`
	Status        Status                 `protobuf:"varint,3,opt,name=status,proto3,enum=shipment.v1.Status" json:"status,omitempty"`
	Note          string                 `protobuf:"bytes,4,opt,name=note,proto3" json:"note,omitempty"`
	OccurredAt    *timestamppb.Timestamp `protobuf:"bytes,5,opt,name=occurred_at,json=occurredAt,proto3" json:"occurred_at,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ShipmentEvent) Reset() {
	*x = ShipmentEvent{}
	mi := &file_api_proto_shipment_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ShipmentEvent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ShipmentEvent) ProtoMessage() {}

func (x *ShipmentEvent) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_shipment_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ShipmentEvent.ProtoReflect.Descriptor instead.
func (*ShipmentEvent) Descriptor() ([]byte, []int) {
	return file_api_proto_shipment_proto_rawDescGZIP(), []int{1}
}

func (x *ShipmentEvent) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *ShipmentEvent) GetShipmentId() string {
	if x != nil {
		return x.ShipmentId
	}
	return ""
}

func (x *ShipmentEvent) GetStatus() Status {
	if x != nil {
		return x.Status
	}
	return Status_STATUS_UNSPECIFIED
}

func (x *ShipmentEvent) GetNote() string {
	if x != nil {
		return x.Note
	}
	return ""
}

func (x *ShipmentEvent) GetOccurredAt() *timestamppb.Timestamp {
	if x != nil {
		return x.OccurredAt
	}
	return nil
}

type CreateShipmentRequest struct {
	state           protoimpl.MessageState `protogen:"open.v1"`
	ReferenceNumber string                 `protobuf:"bytes,1,opt,name=reference_number,json=referenceNumber,proto3" json:"reference_number,omitempty"`
	Origin          string                 `protobuf:"bytes,2,opt,name=origin,proto3" json:"origin,omitempty"`
	Destination     string                 `protobuf:"bytes,3,opt,name=destination,proto3" json:"destination,omitempty"`
	DriverName      string                 `protobuf:"bytes,4,opt,name=driver_name,json=driverName,proto3" json:"driver_name,omitempty"`
	VehicleId       string                 `protobuf:"bytes,5,opt,name=vehicle_id,json=vehicleId,proto3" json:"vehicle_id,omitempty"`
	Amount          float64                `protobuf:"fixed64,6,opt,name=amount,proto3" json:"amount,omitempty"`
	DriverRevenue   float64                `protobuf:"fixed64,7,opt,name=driver_revenue,json=driverRevenue,proto3" json:"driver_revenue,omitempty"`
	unknownFields   protoimpl.UnknownFields
	sizeCache       protoimpl.SizeCache
}

func (x *CreateShipmentRequest) Reset() {
	*x = CreateShipmentRequest{}
	mi := &file_api_proto_shipment_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CreateShipmentRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateShipmentRequest) ProtoMessage() {}

func (x *CreateShipmentRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_shipment_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateShipmentRequest.ProtoReflect.Descriptor instead.
func (*CreateShipmentRequest) Descriptor() ([]byte, []int) {
	return file_api_proto_shipment_proto_rawDescGZIP(), []int{2}
}

func (x *CreateShipmentRequest) GetReferenceNumber() string {
	if x != nil {
		return x.ReferenceNumber
	}
	return ""
}

func (x *CreateShipmentRequest) GetOrigin() string {
	if x != nil {
		return x.Origin
	}
	return ""
}

func (x *CreateShipmentRequest) GetDestination() string {
	if x != nil {
		return x.Destination
	}
	return ""
}

func (x *CreateShipmentRequest) GetDriverName() string {
	if x != nil {
		return x.DriverName
	}
	return ""
}

func (x *CreateShipmentRequest) GetVehicleId() string {
	if x != nil {
		return x.VehicleId
	}
	return ""
}

func (x *CreateShipmentRequest) GetAmount() float64 {
	if x != nil {
		return x.Amount
	}
	return 0
}

func (x *CreateShipmentRequest) GetDriverRevenue() float64 {
	if x != nil {
		return x.DriverRevenue
	}
	return 0
}

type CreateShipmentResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Shipment      *Shipment              `protobuf:"bytes,1,opt,name=shipment,proto3" json:"shipment,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CreateShipmentResponse) Reset() {
	*x = CreateShipmentResponse{}
	mi := &file_api_proto_shipment_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CreateShipmentResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateShipmentResponse) ProtoMessage() {}

func (x *CreateShipmentResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_shipment_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateShipmentResponse.ProtoReflect.Descriptor instead.
func (*CreateShipmentResponse) Descriptor() ([]byte, []int) {
	return file_api_proto_shipment_proto_rawDescGZIP(), []int{3}
}

func (x *CreateShipmentResponse) GetShipment() *Shipment {
	if x != nil {
		return x.Shipment
	}
	return nil
}

type GetShipmentRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetShipmentRequest) Reset() {
	*x = GetShipmentRequest{}
	mi := &file_api_proto_shipment_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetShipmentRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetShipmentRequest) ProtoMessage() {}

func (x *GetShipmentRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_shipment_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetShipmentRequest.ProtoReflect.Descriptor instead.
func (*GetShipmentRequest) Descriptor() ([]byte, []int) {
	return file_api_proto_shipment_proto_rawDescGZIP(), []int{4}
}

func (x *GetShipmentRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type GetShipmentResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Shipment      *Shipment              `protobuf:"bytes,1,opt,name=shipment,proto3" json:"shipment,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetShipmentResponse) Reset() {
	*x = GetShipmentResponse{}
	mi := &file_api_proto_shipment_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetShipmentResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetShipmentResponse) ProtoMessage() {}

func (x *GetShipmentResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_shipment_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetShipmentResponse.ProtoReflect.Descriptor instead.
func (*GetShipmentResponse) Descriptor() ([]byte, []int) {
	return file_api_proto_shipment_proto_rawDescGZIP(), []int{5}
}

func (x *GetShipmentResponse) GetShipment() *Shipment {
	if x != nil {
		return x.Shipment
	}
	return nil
}

type AddShipmentEventRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	ShipmentId    string                 `protobuf:"bytes,1,opt,name=shipment_id,json=shipmentId,proto3" json:"shipment_id,omitempty"`
	Status        Status                 `protobuf:"varint,2,opt,name=status,proto3,enum=shipment.v1.Status" json:"status,omitempty"`
	Note          string                 `protobuf:"bytes,3,opt,name=note,proto3" json:"note,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *AddShipmentEventRequest) Reset() {
	*x = AddShipmentEventRequest{}
	mi := &file_api_proto_shipment_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AddShipmentEventRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddShipmentEventRequest) ProtoMessage() {}

func (x *AddShipmentEventRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_shipment_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AddShipmentEventRequest.ProtoReflect.Descriptor instead.
func (*AddShipmentEventRequest) Descriptor() ([]byte, []int) {
	return file_api_proto_shipment_proto_rawDescGZIP(), []int{6}
}

func (x *AddShipmentEventRequest) GetShipmentId() string {
	if x != nil {
		return x.ShipmentId
	}
	return ""
}

func (x *AddShipmentEventRequest) GetStatus() Status {
	if x != nil {
		return x.Status
	}
	return Status_STATUS_UNSPECIFIED
}

func (x *AddShipmentEventRequest) GetNote() string {
	if x != nil {
		return x.Note
	}
	return ""
}

type AddShipmentEventResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Event         *ShipmentEvent         `protobuf:"bytes,1,opt,name=event,proto3" json:"event,omitempty"`
	Shipment      *Shipment              `protobuf:"bytes,2,opt,name=shipment,proto3" json:"shipment,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *AddShipmentEventResponse) Reset() {
	*x = AddShipmentEventResponse{}
	mi := &file_api_proto_shipment_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AddShipmentEventResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddShipmentEventResponse) ProtoMessage() {}

func (x *AddShipmentEventResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_shipment_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AddShipmentEventResponse.ProtoReflect.Descriptor instead.
func (*AddShipmentEventResponse) Descriptor() ([]byte, []int) {
	return file_api_proto_shipment_proto_rawDescGZIP(), []int{7}
}

func (x *AddShipmentEventResponse) GetEvent() *ShipmentEvent {
	if x != nil {
		return x.Event
	}
	return nil
}

func (x *AddShipmentEventResponse) GetShipment() *Shipment {
	if x != nil {
		return x.Shipment
	}
	return nil
}

type ListShipmentEventsRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	ShipmentId    string                 `protobuf:"bytes,1,opt,name=shipment_id,json=shipmentId,proto3" json:"shipment_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ListShipmentEventsRequest) Reset() {
	*x = ListShipmentEventsRequest{}
	mi := &file_api_proto_shipment_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ListShipmentEventsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListShipmentEventsRequest) ProtoMessage() {}

func (x *ListShipmentEventsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_shipment_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListShipmentEventsRequest.ProtoReflect.Descriptor instead.
func (*ListShipmentEventsRequest) Descriptor() ([]byte, []int) {
	return file_api_proto_shipment_proto_rawDescGZIP(), []int{8}
}

func (x *ListShipmentEventsRequest) GetShipmentId() string {
	if x != nil {
		return x.ShipmentId
	}
	return ""
}

type ListShipmentEventsResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Events        []*ShipmentEvent       `protobuf:"bytes,1,rep,name=events,proto3" json:"events,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ListShipmentEventsResponse) Reset() {
	*x = ListShipmentEventsResponse{}
	mi := &file_api_proto_shipment_proto_msgTypes[9]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ListShipmentEventsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListShipmentEventsResponse) ProtoMessage() {}

func (x *ListShipmentEventsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_shipment_proto_msgTypes[9]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListShipmentEventsResponse.ProtoReflect.Descriptor instead.
func (*ListShipmentEventsResponse) Descriptor() ([]byte, []int) {
	return file_api_proto_shipment_proto_rawDescGZIP(), []int{9}
}

func (x *ListShipmentEventsResponse) GetEvents() []*ShipmentEvent {
	if x != nil {
		return x.Events
	}
	return nil
}

var File_api_proto_shipment_proto protoreflect.FileDescriptor

const file_api_proto_shipment_proto_rawDesc = "" +
	"\n" +
	"\x18api/proto/shipment.proto\x12\vshipment.v1\x1a\x1fgoogle/protobuf/timestamp.proto\"\xb0\x03\n" +
	"\bShipment\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\tR\x02id\x12)\n" +
	"\x10reference_number\x18\x02 \x01(\tR\x0freferenceNumber\x12\x16\n" +
	"\x06origin\x18\x03 \x01(\tR\x06origin\x12 \n" +
	"\vdestination\x18\x04 \x01(\tR\vdestination\x12\x1f\n" +
	"\vdriver_name\x18\x05 \x01(\tR\n" +
	"driverName\x12\x1d\n" +
	"\n" +
	"vehicle_id\x18\x06 \x01(\tR\tvehicleId\x12\x16\n" +
	"\x06amount\x18\a \x01(\x01R\x06amount\x12%\n" +
	"\x0edriver_revenue\x18\b \x01(\x01R\rdriverRevenue\x12:\n" +
	"\x0ecurrent_status\x18\t \x01(\x0e2\x13.shipment.v1.StatusR\rcurrentStatus\x129\n" +
	"\n" +
	"created_at\x18\n" +
	" \x01(\v2\x1a.google.protobuf.TimestampR\tcreatedAt\x129\n" +
	"\n" +
	"updated_at\x18\v \x01(\v2\x1a.google.protobuf.TimestampR\tupdatedAt\"\xbe\x01\n" +
	"\rShipmentEvent\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\tR\x02id\x12\x1f\n" +
	"\vshipment_id\x18\x02 \x01(\tR\n" +
	"shipmentId\x12+\n" +
	"\x06status\x18\x03 \x01(\x0e2\x13.shipment.v1.StatusR\x06status\x12\x12\n" +
	"\x04note\x18\x04 \x01(\tR\x04note\x12;\n" +
	"\voccurred_at\x18\x05 \x01(\v2\x1a.google.protobuf.TimestampR\n" +
	"occurredAt\"\xfb\x01\n" +
	"\x15CreateShipmentRequest\x12)\n" +
	"\x10reference_number\x18\x01 \x01(\tR\x0freferenceNumber\x12\x16\n" +
	"\x06origin\x18\x02 \x01(\tR\x06origin\x12 \n" +
	"\vdestination\x18\x03 \x01(\tR\vdestination\x12\x1f\n" +
	"\vdriver_name\x18\x04 \x01(\tR\n" +
	"driverName\x12\x1d\n" +
	"\n" +
	"vehicle_id\x18\x05 \x01(\tR\tvehicleId\x12\x16\n" +
	"\x06amount\x18\x06 \x01(\x01R\x06amount\x12%\n" +
	"\x0edriver_revenue\x18\a \x01(\x01R\rdriverRevenue\"K\n" +
	"\x16CreateShipmentResponse\x121\n" +
	"\bshipment\x18\x01 \x01(\v2\x15.shipment.v1.ShipmentR\bshipment\"$\n" +
	"\x12GetShipmentRequest\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\tR\x02id\"H\n" +
	"\x13GetShipmentResponse\x121\n" +
	"\bshipment\x18\x01 \x01(\v2\x15.shipment.v1.ShipmentR\bshipment\"{\n" +
	"\x17AddShipmentEventRequest\x12\x1f\n" +
	"\vshipment_id\x18\x01 \x01(\tR\n" +
	"shipmentId\x12+\n" +
	"\x06status\x18\x02 \x01(\x0e2\x13.shipment.v1.StatusR\x06status\x12\x12\n" +
	"\x04note\x18\x03 \x01(\tR\x04note\"\x7f\n" +
	"\x18AddShipmentEventResponse\x120\n" +
	"\x05event\x18\x01 \x01(\v2\x1a.shipment.v1.ShipmentEventR\x05event\x121\n" +
	"\bshipment\x18\x02 \x01(\v2\x15.shipment.v1.ShipmentR\bshipment\"<\n" +
	"\x19ListShipmentEventsRequest\x12\x1f\n" +
	"\vshipment_id\x18\x01 \x01(\tR\n" +
	"shipmentId\"P\n" +
	"\x1aListShipmentEventsResponse\x122\n" +
	"\x06events\x18\x01 \x03(\v2\x1a.shipment.v1.ShipmentEventR\x06events*\x8d\x01\n" +
	"\x06Status\x12\x16\n" +
	"\x12STATUS_UNSPECIFIED\x10\x00\x12\x12\n" +
	"\x0eSTATUS_PENDING\x10\x01\x12\x14\n" +
	"\x10STATUS_PICKED_UP\x10\x02\x12\x15\n" +
	"\x11STATUS_IN_TRANSIT\x10\x03\x12\x14\n" +
	"\x10STATUS_DELIVERED\x10\x04\x12\x14\n" +
	"\x10STATUS_CANCELLED\x10\x052\x86\x03\n" +
	"\x0fShipmentService\x12Y\n" +
	"\x0eCreateShipment\x12\".shipment.v1.CreateShipmentRequest\x1a#.shipment.v1.CreateShipmentResponse\x12P\n" +
	"\vGetShipment\x12\x1f.shipment.v1.GetShipmentRequest\x1a .shipment.v1.GetShipmentResponse\x12_\n" +
	"\x10AddShipmentEvent\x12$.shipment.v1.AddShipmentEventRequest\x1a%.shipment.v1.AddShipmentEventResponse\x12e\n" +
	"\x12ListShipmentEvents\x12&.shipment.v1.ListShipmentEventsRequest\x1a'.shipment.v1.ListShipmentEventsResponseB'Z%shipment-service/api/proto/shipmentpbb\x06proto3"

var (
	file_api_proto_shipment_proto_rawDescOnce sync.Once
	file_api_proto_shipment_proto_rawDescData []byte
)

func file_api_proto_shipment_proto_rawDescGZIP() []byte {
	file_api_proto_shipment_proto_rawDescOnce.Do(func() {
		file_api_proto_shipment_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_api_proto_shipment_proto_rawDesc), len(file_api_proto_shipment_proto_rawDesc)))
	})
	return file_api_proto_shipment_proto_rawDescData
}

var file_api_proto_shipment_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_api_proto_shipment_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_api_proto_shipment_proto_goTypes = []any{
	(Status)(0),                        // 0: shipment.v1.Status
	(*Shipment)(nil),                   // 1: shipment.v1.Shipment
	(*ShipmentEvent)(nil),              // 2: shipment.v1.ShipmentEvent
	(*CreateShipmentRequest)(nil),      // 3: shipment.v1.CreateShipmentRequest
	(*CreateShipmentResponse)(nil),     // 4: shipment.v1.CreateShipmentResponse
	(*GetShipmentRequest)(nil),         // 5: shipment.v1.GetShipmentRequest
	(*GetShipmentResponse)(nil),        // 6: shipment.v1.GetShipmentResponse
	(*AddShipmentEventRequest)(nil),    // 7: shipment.v1.AddShipmentEventRequest
	(*AddShipmentEventResponse)(nil),   // 8: shipment.v1.AddShipmentEventResponse
	(*ListShipmentEventsRequest)(nil),  // 9: shipment.v1.ListShipmentEventsRequest
	(*ListShipmentEventsResponse)(nil), // 10: shipment.v1.ListShipmentEventsResponse
	(*timestamppb.Timestamp)(nil),      // 11: google.protobuf.Timestamp
}
var file_api_proto_shipment_proto_depIdxs = []int32{
	0,  // 0: shipment.v1.Shipment.current_status:type_name -> shipment.v1.Status
	11, // 1: shipment.v1.Shipment.created_at:type_name -> google.protobuf.Timestamp
	11, // 2: shipment.v1.Shipment.updated_at:type_name -> google.protobuf.Timestamp
	0,  // 3: shipment.v1.ShipmentEvent.status:type_name -> shipment.v1.Status
	11, // 4: shipment.v1.ShipmentEvent.occurred_at:type_name -> google.protobuf.Timestamp
	1,  // 5: shipment.v1.CreateShipmentResponse.shipment:type_name -> shipment.v1.Shipment
	1,  // 6: shipment.v1.GetShipmentResponse.shipment:type_name -> shipment.v1.Shipment
	0,  // 7: shipment.v1.AddShipmentEventRequest.status:type_name -> shipment.v1.Status
	2,  // 8: shipment.v1.AddShipmentEventResponse.event:type_name -> shipment.v1.ShipmentEvent
	1,  // 9: shipment.v1.AddShipmentEventResponse.shipment:type_name -> shipment.v1.Shipment
	2,  // 10: shipment.v1.ListShipmentEventsResponse.events:type_name -> shipment.v1.ShipmentEvent
	3,  // 11: shipment.v1.ShipmentService.CreateShipment:input_type -> shipment.v1.CreateShipmentRequest
	5,  // 12: shipment.v1.ShipmentService.GetShipment:input_type -> shipment.v1.GetShipmentRequest
	7,  // 13: shipment.v1.ShipmentService.AddShipmentEvent:input_type -> shipment.v1.AddShipmentEventRequest
	9,  // 14: shipment.v1.ShipmentService.ListShipmentEvents:input_type -> shipment.v1.ListShipmentEventsRequest
	4,  // 15: shipment.v1.ShipmentService.CreateShipment:output_type -> shipment.v1.CreateShipmentResponse
	6,  // 16: shipment.v1.ShipmentService.GetShipment:output_type -> shipment.v1.GetShipmentResponse
	8,  // 17: shipment.v1.ShipmentService.AddShipmentEvent:output_type -> shipment.v1.AddShipmentEventResponse
	10, // 18: shipment.v1.ShipmentService.ListShipmentEvents:output_type -> shipment.v1.ListShipmentEventsResponse
	15, // [15:19] is the sub-list for method output_type
	11, // [11:15] is the sub-list for method input_type
	11, // [11:11] is the sub-list for extension type_name
	11, // [11:11] is the sub-list for extension extendee
	0,  // [0:11] is the sub-list for field type_name
}

func init() { file_api_proto_shipment_proto_init() }
func file_api_proto_shipment_proto_init() {
	if File_api_proto_shipment_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_api_proto_shipment_proto_rawDesc), len(file_api_proto_shipment_proto_rawDesc)),
			NumEnums:      1,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_api_proto_shipment_proto_goTypes,
		DependencyIndexes: file_api_proto_shipment_proto_depIdxs,
		EnumInfos:         file_api_proto_shipment_proto_enumTypes,
		MessageInfos:      file_api_proto_shipment_proto_msgTypes,
	}.Build()
	File_api_proto_shipment_proto = out.File
	file_api_proto_shipment_proto_goTypes = nil
	file_api_proto_shipment_proto_depIdxs = nil
}
