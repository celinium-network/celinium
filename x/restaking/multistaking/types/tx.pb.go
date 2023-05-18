// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: celinium/restaking/multistake/v1/tx.proto

package types

import (
	context "context"
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
	types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/gogoproto/gogoproto"
	grpc1 "github.com/gogo/protobuf/grpc"
	proto "github.com/gogo/protobuf/proto"
	github_com_gogo_protobuf_types "github.com/gogo/protobuf/types"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	_ "google.golang.org/protobuf/types/known/timestamppb"
	io "io"
	math "math"
	math_bits "math/bits"
	time "time"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf
var _ = time.Kitchen

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type MsgRegisterMultiStakingDenom struct {
	Sender string `protobuf:"bytes,1,opt,name=sender,proto3" json:"sender,omitempty" yaml:"sender"`
	Deonm  string `protobuf:"bytes,2,opt,name=deonm,proto3" json:"deonm,omitempty"`
}

func (m *MsgRegisterMultiStakingDenom) Reset()         { *m = MsgRegisterMultiStakingDenom{} }
func (m *MsgRegisterMultiStakingDenom) String() string { return proto.CompactTextString(m) }
func (*MsgRegisterMultiStakingDenom) ProtoMessage()    {}
func (*MsgRegisterMultiStakingDenom) Descriptor() ([]byte, []int) {
	return fileDescriptor_46a477979d5ff9d4, []int{0}
}
func (m *MsgRegisterMultiStakingDenom) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgRegisterMultiStakingDenom) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgRegisterMultiStakingDenom.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgRegisterMultiStakingDenom) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgRegisterMultiStakingDenom.Merge(m, src)
}
func (m *MsgRegisterMultiStakingDenom) XXX_Size() int {
	return m.Size()
}
func (m *MsgRegisterMultiStakingDenom) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgRegisterMultiStakingDenom.DiscardUnknown(m)
}

var xxx_messageInfo_MsgRegisterMultiStakingDenom proto.InternalMessageInfo

func (m *MsgRegisterMultiStakingDenom) GetSender() string {
	if m != nil {
		return m.Sender
	}
	return ""
}

func (m *MsgRegisterMultiStakingDenom) GetDeonm() string {
	if m != nil {
		return m.Deonm
	}
	return ""
}

type MsgRegisterMultiStakingDenomResponse struct {
}

func (m *MsgRegisterMultiStakingDenomResponse) Reset()         { *m = MsgRegisterMultiStakingDenomResponse{} }
func (m *MsgRegisterMultiStakingDenomResponse) String() string { return proto.CompactTextString(m) }
func (*MsgRegisterMultiStakingDenomResponse) ProtoMessage()    {}
func (*MsgRegisterMultiStakingDenomResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_46a477979d5ff9d4, []int{1}
}
func (m *MsgRegisterMultiStakingDenomResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgRegisterMultiStakingDenomResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgRegisterMultiStakingDenomResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgRegisterMultiStakingDenomResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgRegisterMultiStakingDenomResponse.Merge(m, src)
}
func (m *MsgRegisterMultiStakingDenomResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgRegisterMultiStakingDenomResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgRegisterMultiStakingDenomResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgRegisterMultiStakingDenomResponse proto.InternalMessageInfo

type MsgDelegate struct {
	DelegatorAddress string     `protobuf:"bytes,1,opt,name=delegator_address,json=delegatorAddress,proto3" json:"delegator_address,omitempty"`
	ValidatorAddress string     `protobuf:"bytes,2,opt,name=validator_address,json=validatorAddress,proto3" json:"validator_address,omitempty"`
	Amount           types.Coin `protobuf:"bytes,3,opt,name=amount,proto3" json:"amount"`
}

func (m *MsgDelegate) Reset()         { *m = MsgDelegate{} }
func (m *MsgDelegate) String() string { return proto.CompactTextString(m) }
func (*MsgDelegate) ProtoMessage()    {}
func (*MsgDelegate) Descriptor() ([]byte, []int) {
	return fileDescriptor_46a477979d5ff9d4, []int{2}
}
func (m *MsgDelegate) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgDelegate) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgDelegate.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgDelegate) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgDelegate.Merge(m, src)
}
func (m *MsgDelegate) XXX_Size() int {
	return m.Size()
}
func (m *MsgDelegate) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgDelegate.DiscardUnknown(m)
}

var xxx_messageInfo_MsgDelegate proto.InternalMessageInfo

// MsgDelegateResponse defines the Msg/Delegate response type.
type MsgDelegateResponse struct {
}

func (m *MsgDelegateResponse) Reset()         { *m = MsgDelegateResponse{} }
func (m *MsgDelegateResponse) String() string { return proto.CompactTextString(m) }
func (*MsgDelegateResponse) ProtoMessage()    {}
func (*MsgDelegateResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_46a477979d5ff9d4, []int{3}
}
func (m *MsgDelegateResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgDelegateResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgDelegateResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgDelegateResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgDelegateResponse.Merge(m, src)
}
func (m *MsgDelegateResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgDelegateResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgDelegateResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgDelegateResponse proto.InternalMessageInfo

type MsgUndelegate struct {
	DelegatorAddress string     `protobuf:"bytes,1,opt,name=delegator_address,json=delegatorAddress,proto3" json:"delegator_address,omitempty"`
	ValidatorAddress string     `protobuf:"bytes,2,opt,name=validator_address,json=validatorAddress,proto3" json:"validator_address,omitempty"`
	Amount           types.Coin `protobuf:"bytes,3,opt,name=amount,proto3" json:"amount"`
}

func (m *MsgUndelegate) Reset()         { *m = MsgUndelegate{} }
func (m *MsgUndelegate) String() string { return proto.CompactTextString(m) }
func (*MsgUndelegate) ProtoMessage()    {}
func (*MsgUndelegate) Descriptor() ([]byte, []int) {
	return fileDescriptor_46a477979d5ff9d4, []int{4}
}
func (m *MsgUndelegate) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgUndelegate) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgUndelegate.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgUndelegate) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgUndelegate.Merge(m, src)
}
func (m *MsgUndelegate) XXX_Size() int {
	return m.Size()
}
func (m *MsgUndelegate) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgUndelegate.DiscardUnknown(m)
}

var xxx_messageInfo_MsgUndelegate proto.InternalMessageInfo

type MsgUndelegateResponse struct {
	CompletionTime time.Time  `protobuf:"bytes,1,opt,name=completion_time,json=completionTime,proto3,stdtime" json:"completion_time"`
	Amount         types.Coin `protobuf:"bytes,2,opt,name=amount,proto3" json:"amount"`
}

func (m *MsgUndelegateResponse) Reset()         { *m = MsgUndelegateResponse{} }
func (m *MsgUndelegateResponse) String() string { return proto.CompactTextString(m) }
func (*MsgUndelegateResponse) ProtoMessage()    {}
func (*MsgUndelegateResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_46a477979d5ff9d4, []int{5}
}
func (m *MsgUndelegateResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgUndelegateResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgUndelegateResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgUndelegateResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgUndelegateResponse.Merge(m, src)
}
func (m *MsgUndelegateResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgUndelegateResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgUndelegateResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgUndelegateResponse proto.InternalMessageInfo

func (m *MsgUndelegateResponse) GetCompletionTime() time.Time {
	if m != nil {
		return m.CompletionTime
	}
	return time.Time{}
}

func (m *MsgUndelegateResponse) GetAmount() types.Coin {
	if m != nil {
		return m.Amount
	}
	return types.Coin{}
}

func init() {
	proto.RegisterType((*MsgRegisterMultiStakingDenom)(nil), "celinium.restaking.multistake.v1.MsgRegisterMultiStakingDenom")
	proto.RegisterType((*MsgRegisterMultiStakingDenomResponse)(nil), "celinium.restaking.multistake.v1.MsgRegisterMultiStakingDenomResponse")
	proto.RegisterType((*MsgDelegate)(nil), "celinium.restaking.multistake.v1.MsgDelegate")
	proto.RegisterType((*MsgDelegateResponse)(nil), "celinium.restaking.multistake.v1.MsgDelegateResponse")
	proto.RegisterType((*MsgUndelegate)(nil), "celinium.restaking.multistake.v1.MsgUndelegate")
	proto.RegisterType((*MsgUndelegateResponse)(nil), "celinium.restaking.multistake.v1.MsgUndelegateResponse")
}

func init() {
	proto.RegisterFile("celinium/restaking/multistake/v1/tx.proto", fileDescriptor_46a477979d5ff9d4)
}

var fileDescriptor_46a477979d5ff9d4 = []byte{
	// 528 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xdc, 0x54, 0x3f, 0x6f, 0x13, 0x31,
	0x14, 0x3f, 0xb7, 0x10, 0x15, 0x47, 0x05, 0x72, 0xa4, 0x52, 0x12, 0xc1, 0x5d, 0x38, 0x21, 0x68,
	0x87, 0xfa, 0x94, 0x30, 0x20, 0x75, 0x40, 0x6a, 0x28, 0x6c, 0xc7, 0x70, 0x05, 0x06, 0x96, 0xe8,
	0x92, 0x33, 0x96, 0xc5, 0xd9, 0x8e, 0xce, 0x4e, 0xd4, 0x7e, 0x03, 0xc6, 0x7e, 0x03, 0x2a, 0x31,
	0xb2, 0xf6, 0x43, 0x74, 0xac, 0x3a, 0x31, 0x15, 0x94, 0x0c, 0xc0, 0xca, 0x27, 0x40, 0x3e, 0xfb,
	0xd2, 0x3f, 0x82, 0x22, 0xc4, 0xc6, 0xe6, 0xf7, 0xde, 0xef, 0x3d, 0xff, 0xde, 0x4f, 0xfe, 0x19,
	0xae, 0x0d, 0x71, 0x46, 0x39, 0x1d, 0xb3, 0x30, 0xc7, 0x52, 0x25, 0x6f, 0x29, 0x27, 0x21, 0x1b,
	0x67, 0x8a, 0xea, 0x00, 0x87, 0x93, 0x4e, 0xa8, 0x76, 0xd0, 0x28, 0x17, 0x4a, 0xb8, 0xed, 0x12,
	0x8a, 0xe6, 0x50, 0x74, 0x0a, 0x45, 0x93, 0x4e, 0xcb, 0x27, 0x42, 0x90, 0x0c, 0x87, 0x05, 0x7e,
	0x30, 0x7e, 0x13, 0x2a, 0xca, 0x34, 0x94, 0x8d, 0xcc, 0x88, 0x56, 0x9d, 0x08, 0x22, 0x8a, 0x63,
	0xa8, 0x4f, 0x36, 0xdb, 0x1c, 0x0a, 0xc9, 0x84, 0xec, 0x9b, 0x82, 0x09, 0x6c, 0xc9, 0x33, 0x51,
	0x38, 0x48, 0xa4, 0x26, 0x33, 0xc0, 0x2a, 0xe9, 0x84, 0x43, 0x41, 0xb9, 0xa9, 0x07, 0x7d, 0x78,
	0x3b, 0x92, 0x24, 0xc6, 0x84, 0x4a, 0x85, 0xf3, 0x48, 0xb3, 0xd9, 0x36, 0xd4, 0xb6, 0x30, 0x17,
	0xcc, 0x5d, 0x83, 0x15, 0x89, 0x79, 0x8a, 0xf3, 0x06, 0x68, 0x83, 0xd5, 0x6b, 0xbd, 0xda, 0x8f,
	0x13, 0x7f, 0x79, 0x37, 0x61, 0xd9, 0x46, 0x60, 0xf2, 0x41, 0x6c, 0x01, 0x6e, 0x1d, 0x5e, 0x4d,
	0xb1, 0xe0, 0xac, 0xb1, 0xa0, 0x91, 0xb1, 0x09, 0x82, 0xfb, 0xf0, 0xde, 0x65, 0x17, 0xc4, 0x58,
	0x8e, 0x04, 0x97, 0x38, 0xf8, 0x0a, 0x60, 0x35, 0x92, 0x64, 0x0b, 0x67, 0x98, 0x24, 0x0a, 0xbb,
	0x4f, 0x61, 0x2d, 0x35, 0x67, 0x91, 0xf7, 0x93, 0x34, 0xcd, 0xb1, 0x94, 0x96, 0x43, 0xe3, 0xf8,
	0x60, 0xbd, 0x6e, 0xb7, 0xdc, 0x34, 0x95, 0x6d, 0x95, 0x53, 0x4e, 0xe2, 0x9b, 0xf3, 0x16, 0x9b,
	0x77, 0x9f, 0xc3, 0xda, 0x24, 0xc9, 0x68, 0x7a, 0x6e, 0x4c, 0x41, 0xb0, 0x77, 0xf7, 0xf8, 0x60,
	0xfd, 0x8e, 0x1d, 0xf3, 0xaa, 0xc4, 0x5c, 0x98, 0x37, 0xb9, 0x90, 0x77, 0x1f, 0xc1, 0x4a, 0xc2,
	0xc4, 0x98, 0xab, 0xc6, 0x62, 0x1b, 0xac, 0x56, 0xbb, 0x4d, 0x64, 0x27, 0x68, 0x81, 0x91, 0x15,
	0x18, 0x3d, 0x11, 0x94, 0xf7, 0xae, 0x1c, 0x9e, 0xf8, 0x4e, 0x6c, 0xe1, 0x1b, 0x4b, 0xef, 0xf6,
	0x7d, 0xe7, 0xdb, 0xbe, 0xef, 0x04, 0x2b, 0xf0, 0xd6, 0x99, 0x45, 0xe7, 0x02, 0x7c, 0x07, 0x70,
	0x39, 0x92, 0xe4, 0x25, 0x4f, 0xff, 0x7f, 0x09, 0xde, 0x03, 0xb8, 0x72, 0x6e, 0xd7, 0x52, 0x05,
	0x37, 0x82, 0x37, 0x86, 0x82, 0x8d, 0x32, 0xac, 0xa8, 0xe0, 0x7d, 0xfd, 0xfc, 0x8b, 0x8d, 0xab,
	0xdd, 0x16, 0x32, 0xde, 0x40, 0xa5, 0x37, 0xd0, 0x8b, 0xd2, 0x1b, 0xbd, 0x25, 0x7d, 0xcd, 0xde,
	0x67, 0x1f, 0xc4, 0xd7, 0x4f, 0x9b, 0x75, 0xf9, 0x0c, 0xd7, 0x85, 0xbf, 0xe2, 0xda, 0xfd, 0x08,
	0xe0, 0x62, 0x24, 0x89, 0xfb, 0x01, 0xc0, 0xe6, 0xef, 0xdd, 0xf1, 0x18, 0xfd, 0xc9, 0xd2, 0xe8,
	0xb2, 0xc7, 0xdf, 0x7a, 0xf6, 0x6f, 0xfd, 0xa5, 0x6a, 0xbd, 0xcd, 0xc3, 0xa9, 0x07, 0x8e, 0xa6,
	0x1e, 0xf8, 0x32, 0xf5, 0xc0, 0xde, 0xcc, 0x73, 0x8e, 0x66, 0x9e, 0xf3, 0x69, 0xe6, 0x39, 0xaf,
	0x1f, 0xcc, 0xbf, 0xa7, 0x9d, 0x5f, 0x7d, 0x50, 0x3a, 0x50, 0xbb, 0x23, 0x2c, 0x07, 0x95, 0x42,
	0xd7, 0x87, 0x3f, 0x03, 0x00, 0x00, 0xff, 0xff, 0x82, 0x41, 0x2a, 0xb0, 0xd0, 0x04, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// MsgClient is the client API for Msg service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type MsgClient interface {
	RegisterMultiStakingDenom(ctx context.Context, in *MsgRegisterMultiStakingDenom, opts ...grpc.CallOption) (*MsgRegisterMultiStakingDenomResponse, error)
}

type msgClient struct {
	cc grpc1.ClientConn
}

func NewMsgClient(cc grpc1.ClientConn) MsgClient {
	return &msgClient{cc}
}

func (c *msgClient) RegisterMultiStakingDenom(ctx context.Context, in *MsgRegisterMultiStakingDenom, opts ...grpc.CallOption) (*MsgRegisterMultiStakingDenomResponse, error) {
	out := new(MsgRegisterMultiStakingDenomResponse)
	err := c.cc.Invoke(ctx, "/celinium.restaking.multistake.v1.Msg/RegisterMultiStakingDenom", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MsgServer is the server API for Msg service.
type MsgServer interface {
	RegisterMultiStakingDenom(context.Context, *MsgRegisterMultiStakingDenom) (*MsgRegisterMultiStakingDenomResponse, error)
}

// UnimplementedMsgServer can be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func (*UnimplementedMsgServer) RegisterMultiStakingDenom(ctx context.Context, req *MsgRegisterMultiStakingDenom) (*MsgRegisterMultiStakingDenomResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegisterMultiStakingDenom not implemented")
}

func RegisterMsgServer(s grpc1.Server, srv MsgServer) {
	s.RegisterService(&_Msg_serviceDesc, srv)
}

func _Msg_RegisterMultiStakingDenom_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgRegisterMultiStakingDenom)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).RegisterMultiStakingDenom(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/celinium.restaking.multistake.v1.Msg/RegisterMultiStakingDenom",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).RegisterMultiStakingDenom(ctx, req.(*MsgRegisterMultiStakingDenom))
	}
	return interceptor(ctx, in, info, handler)
}

var _Msg_serviceDesc = grpc.ServiceDesc{
	ServiceName: "celinium.restaking.multistake.v1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RegisterMultiStakingDenom",
			Handler:    _Msg_RegisterMultiStakingDenom_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "celinium/restaking/multistake/v1/tx.proto",
}

func (m *MsgRegisterMultiStakingDenom) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgRegisterMultiStakingDenom) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgRegisterMultiStakingDenom) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Deonm) > 0 {
		i -= len(m.Deonm)
		copy(dAtA[i:], m.Deonm)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Deonm)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Sender) > 0 {
		i -= len(m.Sender)
		copy(dAtA[i:], m.Sender)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Sender)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MsgRegisterMultiStakingDenomResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgRegisterMultiStakingDenomResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgRegisterMultiStakingDenomResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *MsgDelegate) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgDelegate) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgDelegate) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.Amount.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintTx(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x1a
	if len(m.ValidatorAddress) > 0 {
		i -= len(m.ValidatorAddress)
		copy(dAtA[i:], m.ValidatorAddress)
		i = encodeVarintTx(dAtA, i, uint64(len(m.ValidatorAddress)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.DelegatorAddress) > 0 {
		i -= len(m.DelegatorAddress)
		copy(dAtA[i:], m.DelegatorAddress)
		i = encodeVarintTx(dAtA, i, uint64(len(m.DelegatorAddress)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MsgDelegateResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgDelegateResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgDelegateResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *MsgUndelegate) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgUndelegate) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgUndelegate) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.Amount.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintTx(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x1a
	if len(m.ValidatorAddress) > 0 {
		i -= len(m.ValidatorAddress)
		copy(dAtA[i:], m.ValidatorAddress)
		i = encodeVarintTx(dAtA, i, uint64(len(m.ValidatorAddress)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.DelegatorAddress) > 0 {
		i -= len(m.DelegatorAddress)
		copy(dAtA[i:], m.DelegatorAddress)
		i = encodeVarintTx(dAtA, i, uint64(len(m.DelegatorAddress)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MsgUndelegateResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgUndelegateResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgUndelegateResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.Amount.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintTx(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	n4, err4 := github_com_gogo_protobuf_types.StdTimeMarshalTo(m.CompletionTime, dAtA[i-github_com_gogo_protobuf_types.SizeOfStdTime(m.CompletionTime):])
	if err4 != nil {
		return 0, err4
	}
	i -= n4
	i = encodeVarintTx(dAtA, i, uint64(n4))
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func encodeVarintTx(dAtA []byte, offset int, v uint64) int {
	offset -= sovTx(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *MsgRegisterMultiStakingDenom) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Sender)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.Deonm)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	return n
}

func (m *MsgRegisterMultiStakingDenomResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *MsgDelegate) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.DelegatorAddress)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.ValidatorAddress)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = m.Amount.Size()
	n += 1 + l + sovTx(uint64(l))
	return n
}

func (m *MsgDelegateResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *MsgUndelegate) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.DelegatorAddress)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.ValidatorAddress)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = m.Amount.Size()
	n += 1 + l + sovTx(uint64(l))
	return n
}

func (m *MsgUndelegateResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = github_com_gogo_protobuf_types.SizeOfStdTime(m.CompletionTime)
	n += 1 + l + sovTx(uint64(l))
	l = m.Amount.Size()
	n += 1 + l + sovTx(uint64(l))
	return n
}

func sovTx(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozTx(x uint64) (n int) {
	return sovTx(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *MsgRegisterMultiStakingDenom) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MsgRegisterMultiStakingDenom: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgRegisterMultiStakingDenom: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Sender", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Sender = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Deonm", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Deonm = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *MsgRegisterMultiStakingDenomResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MsgRegisterMultiStakingDenomResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgRegisterMultiStakingDenomResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *MsgDelegate) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MsgDelegate: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgDelegate: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DelegatorAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.DelegatorAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ValidatorAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ValidatorAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Amount", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Amount.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *MsgDelegateResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MsgDelegateResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgDelegateResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *MsgUndelegate) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MsgUndelegate: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgUndelegate: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DelegatorAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.DelegatorAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ValidatorAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ValidatorAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Amount", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Amount.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *MsgUndelegateResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MsgUndelegateResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgUndelegateResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field CompletionTime", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := github_com_gogo_protobuf_types.StdTimeUnmarshal(&m.CompletionTime, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Amount", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Amount.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipTx(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowTx
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowTx
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowTx
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthTx
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupTx
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthTx
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthTx        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowTx          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupTx = fmt.Errorf("proto: unexpected end of group")
)
