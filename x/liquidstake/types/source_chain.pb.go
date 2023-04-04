// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: celinium/liquidstake/v1/source_chain.proto

package types

import (
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type Validator struct {
	// The address of source chain validator account.
	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	// Total amount of delegation.
	DelegationAmount Int `protobuf:"bytes,2,opt,name=delegationAmount,proto3,customtype=Int" json:"delegationAmount"`
	// The weight used for distribute delegation funds.
	Weight uint64 `protobuf:"varint,3,opt,name=weight,proto3" json:"weight,omitempty"`
}

func (m *Validator) Reset()         { *m = Validator{} }
func (m *Validator) String() string { return proto.CompactTextString(m) }
func (*Validator) ProtoMessage()    {}
func (*Validator) Descriptor() ([]byte, []int) {
	return fileDescriptor_9717b2e9147633e9, []int{0}
}
func (m *Validator) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Validator) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Validator.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Validator) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Validator.Merge(m, src)
}
func (m *Validator) XXX_Size() int {
	return m.Size()
}
func (m *Validator) XXX_DiscardUnknown() {
	xxx_messageInfo_Validator.DiscardUnknown(m)
}

var xxx_messageInfo_Validator proto.InternalMessageInfo

func (m *Validator) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

func (m *Validator) GetWeight() uint64 {
	if m != nil {
		return m.Weight
	}
	return 0
}

type SourceChain struct {
	// The chain id of source chain.
	ChainID string `protobuf:"bytes,1,opt,name=chainID,proto3" json:"chainID,omitempty"`
	// ibc connection id
	ConnectionID string `protobuf:"bytes,2,opt,name=connectionID,proto3" json:"connectionID,omitempty"`
	// ibc transfer channel id
	TrasnferChannelID string `protobuf:"bytes,3,opt,name=trasnferChannelID,proto3" json:"trasnferChannelID,omitempty"`
	// validator address prefix of source chain.
	Bech32ValidatorAddrPrefix string       `protobuf:"bytes,4,opt,name=bech32ValidatorAddrPrefix,proto3" json:"bech32ValidatorAddrPrefix,omitempty"`
	Validators                []*Validator `protobuf:"bytes,5,rep,name=validators,proto3" json:"validators,omitempty"`
	// The address of Interchain account for receiving POS reward
	WithdrawAddress string `protobuf:"bytes,6,opt,name=withdrawAddress,proto3" json:"withdrawAddress,omitempty"`
	// The address of Interchain account for delegation
	DelegateAddress string `protobuf:"bytes,7,opt,name=delegateAddress,proto3" json:"delegateAddress,omitempty"`
	// The address of Interchain account for unbound
	UnboudAddress string `protobuf:"bytes,8,opt,name=unboudAddress,proto3" json:"unboudAddress,omitempty"`
	// Redemption ratio in the current epoch
	Redemptionratio Dec `protobuf:"bytes,9,opt,name=redemptionratio,proto3,customtype=Dec" json:"redemptionratio"`
	// The denom of cross chain token.
	IbcDenom string `protobuf:"bytes,10,opt,name=ibcDenom,proto3" json:"ibcDenom,omitempty"`
	// The denom of source chain native token.
	NativeDenom string `protobuf:"bytes,11,opt,name=nativeDenom,proto3" json:"nativeDenom,omitempty"`
	// Derivative token denom generated after liquid stake
	DerivativeDenom string `protobuf:"bytes,12,opt,name=derivativeDenom,proto3" json:"derivativeDenom,omitempty"`
	// The amount of staked token.
	StakedAmount Int `protobuf:"bytes,13,opt,name=stakedAmount,proto3,customtype=Int" json:"stakedAmount"`
}

func (m *SourceChain) Reset()         { *m = SourceChain{} }
func (m *SourceChain) String() string { return proto.CompactTextString(m) }
func (*SourceChain) ProtoMessage()    {}
func (*SourceChain) Descriptor() ([]byte, []int) {
	return fileDescriptor_9717b2e9147633e9, []int{1}
}
func (m *SourceChain) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *SourceChain) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_SourceChain.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *SourceChain) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SourceChain.Merge(m, src)
}
func (m *SourceChain) XXX_Size() int {
	return m.Size()
}
func (m *SourceChain) XXX_DiscardUnknown() {
	xxx_messageInfo_SourceChain.DiscardUnknown(m)
}

var xxx_messageInfo_SourceChain proto.InternalMessageInfo

func (m *SourceChain) GetChainID() string {
	if m != nil {
		return m.ChainID
	}
	return ""
}

func (m *SourceChain) GetConnectionID() string {
	if m != nil {
		return m.ConnectionID
	}
	return ""
}

func (m *SourceChain) GetTrasnferChannelID() string {
	if m != nil {
		return m.TrasnferChannelID
	}
	return ""
}

func (m *SourceChain) GetBech32ValidatorAddrPrefix() string {
	if m != nil {
		return m.Bech32ValidatorAddrPrefix
	}
	return ""
}

func (m *SourceChain) GetValidators() []*Validator {
	if m != nil {
		return m.Validators
	}
	return nil
}

func (m *SourceChain) GetWithdrawAddress() string {
	if m != nil {
		return m.WithdrawAddress
	}
	return ""
}

func (m *SourceChain) GetDelegateAddress() string {
	if m != nil {
		return m.DelegateAddress
	}
	return ""
}

func (m *SourceChain) GetUnboudAddress() string {
	if m != nil {
		return m.UnboudAddress
	}
	return ""
}

func (m *SourceChain) GetIbcDenom() string {
	if m != nil {
		return m.IbcDenom
	}
	return ""
}

func (m *SourceChain) GetNativeDenom() string {
	if m != nil {
		return m.NativeDenom
	}
	return ""
}

func (m *SourceChain) GetDerivativeDenom() string {
	if m != nil {
		return m.DerivativeDenom
	}
	return ""
}

func init() {
	proto.RegisterType((*Validator)(nil), "celinium.liquidstake.v1.Validator")
	proto.RegisterType((*SourceChain)(nil), "celinium.liquidstake.v1.SourceChain")
}

func init() {
	proto.RegisterFile("celinium/liquidstake/v1/source_chain.proto", fileDescriptor_9717b2e9147633e9)
}

var fileDescriptor_9717b2e9147633e9 = []byte{
	// 526 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x53, 0x31, 0x6f, 0x1a, 0x4d,
	0x10, 0xe5, 0x3e, 0x30, 0x98, 0x01, 0xcb, 0x5f, 0x56, 0x56, 0x72, 0xa0, 0xe8, 0x40, 0x54, 0x28,
	0x8a, 0x41, 0xc6, 0x52, 0x8a, 0x28, 0x89, 0x64, 0xb8, 0x14, 0x74, 0xd1, 0x59, 0x4a, 0x91, 0xc6,
	0x5a, 0xf6, 0xc6, 0xb0, 0x0a, 0xec, 0x92, 0xbd, 0x3d, 0x70, 0xfe, 0x45, 0x7e, 0x42, 0x8a, 0xf4,
	0x69, 0xfc, 0x23, 0x5c, 0x5a, 0xae, 0xa2, 0x14, 0x56, 0x04, 0x4d, 0x7e, 0x46, 0x74, 0x7b, 0x77,
	0x08, 0xb0, 0x2c, 0xba, 0x9b, 0x99, 0xf7, 0xde, 0xad, 0xde, 0x9b, 0x81, 0x17, 0x0c, 0xc7, 0x5c,
	0xf0, 0x70, 0xd2, 0x1e, 0xf3, 0x2f, 0x21, 0xf7, 0x03, 0x4d, 0x3f, 0x63, 0x7b, 0x76, 0xd2, 0x0e,
	0x64, 0xa8, 0x18, 0x5e, 0xb0, 0x11, 0xe5, 0xa2, 0x35, 0x55, 0x52, 0x4b, 0xf2, 0x2c, 0xc5, 0xb6,
	0xd6, 0xb0, 0xad, 0xd9, 0x49, 0xf5, 0x68, 0x28, 0x87, 0xd2, 0x60, 0xda, 0xd1, 0x57, 0x0c, 0xaf,
	0x56, 0x98, 0x0c, 0x26, 0x32, 0xb8, 0x88, 0x07, 0x71, 0x11, 0x8f, 0x1a, 0x3f, 0x2c, 0x28, 0x7e,
	0xa4, 0x63, 0xee, 0x53, 0x2d, 0x15, 0xe9, 0x40, 0x81, 0xfa, 0xbe, 0xc2, 0x20, 0xb0, 0xad, 0xba,
	0xd5, 0x2c, 0x76, 0xed, 0xbb, 0xeb, 0xe3, 0xa3, 0x84, 0x70, 0x16, 0x4f, 0xce, 0xb5, 0xe2, 0x62,
	0xe8, 0xa5, 0x40, 0xf2, 0x1e, 0xfe, 0xf7, 0x71, 0x8c, 0x43, 0xaa, 0xb9, 0x14, 0x67, 0x13, 0x19,
	0x0a, 0x6d, 0xff, 0x67, 0xc8, 0x95, 0x9b, 0xfb, 0x5a, 0xe6, 0xf7, 0x7d, 0x2d, 0xdb, 0x17, 0xfa,
	0xee, 0xfa, 0x18, 0x12, 0x9d, 0xbe, 0xd0, 0xde, 0x03, 0x0a, 0x79, 0x0a, 0xf9, 0x39, 0xf2, 0xe1,
	0x48, 0xdb, 0xd9, 0xba, 0xd5, 0xcc, 0x79, 0x49, 0xf5, 0x3a, 0xf7, 0xf7, 0x7b, 0xcd, 0x6a, 0xfc,
	0xdc, 0x83, 0xd2, 0xb9, 0xf1, 0xa1, 0x17, 0xd9, 0x40, 0x6c, 0x28, 0x18, 0x3f, 0xfa, 0x6e, 0xfc,
	0x50, 0x2f, 0x2d, 0x49, 0x03, 0xca, 0x4c, 0x0a, 0x81, 0x2c, 0xd2, 0xee, 0xbb, 0xf1, 0x53, 0xbc,
	0x8d, 0x1e, 0x79, 0x09, 0x4f, 0xb4, 0xa2, 0x81, 0xb8, 0x44, 0xd5, 0x1b, 0x51, 0x21, 0x70, 0xdc,
	0x77, 0xcd, 0x6f, 0x8b, 0xde, 0xc3, 0x01, 0x79, 0x03, 0x95, 0x01, 0xb2, 0xd1, 0x69, 0x67, 0xe5,
	0x53, 0xe4, 0xc4, 0x07, 0x85, 0x97, 0xfc, 0xca, 0xce, 0x19, 0xd6, 0xe3, 0x00, 0xd2, 0x05, 0x98,
	0xa5, 0xed, 0xc0, 0xde, 0xab, 0x67, 0x9b, 0xa5, 0x4e, 0xa3, 0xf5, 0x48, 0x7e, 0xad, 0x95, 0x82,
	0xb7, 0xc6, 0x22, 0x5d, 0x38, 0x9c, 0x73, 0x3d, 0xf2, 0x15, 0x9d, 0x27, 0x21, 0xd8, 0xf9, 0x1d,
	0xf1, 0x6c, 0x13, 0x22, 0x8d, 0xc4, 0x73, 0x4c, 0x35, 0x0a, 0xbb, 0x34, 0xb6, 0x08, 0xe4, 0x1d,
	0x1c, 0x84, 0x62, 0x20, 0x43, 0x3f, 0x55, 0xd8, 0xdf, 0xa1, 0xb0, 0x09, 0x27, 0x3d, 0x38, 0x54,
	0xe8, 0xe3, 0x64, 0x1a, 0xe5, 0xa0, 0xa2, 0xf4, 0xed, 0xe2, 0xe6, 0xa6, 0xb8, 0xc8, 0xd6, 0x36,
	0xc5, 0x45, 0xe6, 0x6d, 0x33, 0x48, 0x15, 0xf6, 0xf9, 0x80, 0xb9, 0x28, 0xe4, 0xc4, 0x06, 0xe3,
	0xfe, 0xaa, 0x26, 0x75, 0x28, 0x09, 0xaa, 0xf9, 0x0c, 0xe3, 0x71, 0xc9, 0x8c, 0xd7, 0x5b, 0xa4,
	0x19, 0xd9, 0xa0, 0xf8, 0x6c, 0x0d, 0x55, 0x36, 0xa8, 0xed, 0x36, 0x79, 0x0b, 0x65, 0x13, 0x8b,
	0x9f, 0xec, 0xf4, 0xc1, 0xae, 0x9d, 0xde, 0x80, 0x77, 0x5f, 0xdd, 0x2c, 0x1c, 0xeb, 0x76, 0xe1,
	0x58, 0x7f, 0x16, 0x8e, 0xf5, 0x6d, 0xe9, 0x64, 0x6e, 0x97, 0x4e, 0xe6, 0xd7, 0xd2, 0xc9, 0x7c,
	0x7a, 0xbe, 0x3a, 0xf4, 0xab, 0x8d, 0x53, 0xd7, 0x5f, 0xa7, 0x18, 0x0c, 0xf2, 0xe6, 0x2e, 0x4f,
	0xff, 0x05, 0x00, 0x00, 0xff, 0xff, 0x40, 0x77, 0x12, 0x58, 0x0f, 0x04, 0x00, 0x00,
}

func (this *Validator) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*Validator)
	if !ok {
		that2, ok := that.(Validator)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if this.Address != that1.Address {
		return false
	}
	if !this.DelegationAmount.Equal(that1.DelegationAmount) {
		return false
	}
	if this.Weight != that1.Weight {
		return false
	}
	return true
}
func (m *Validator) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Validator) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Validator) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Weight != 0 {
		i = encodeVarintSourceChain(dAtA, i, uint64(m.Weight))
		i--
		dAtA[i] = 0x18
	}
	{
		size := m.DelegationAmount.Size()
		i -= size
		if _, err := m.DelegationAmount.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintSourceChain(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	if len(m.Address) > 0 {
		i -= len(m.Address)
		copy(dAtA[i:], m.Address)
		i = encodeVarintSourceChain(dAtA, i, uint64(len(m.Address)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *SourceChain) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SourceChain) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *SourceChain) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size := m.StakedAmount.Size()
		i -= size
		if _, err := m.StakedAmount.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintSourceChain(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x6a
	if len(m.DerivativeDenom) > 0 {
		i -= len(m.DerivativeDenom)
		copy(dAtA[i:], m.DerivativeDenom)
		i = encodeVarintSourceChain(dAtA, i, uint64(len(m.DerivativeDenom)))
		i--
		dAtA[i] = 0x62
	}
	if len(m.NativeDenom) > 0 {
		i -= len(m.NativeDenom)
		copy(dAtA[i:], m.NativeDenom)
		i = encodeVarintSourceChain(dAtA, i, uint64(len(m.NativeDenom)))
		i--
		dAtA[i] = 0x5a
	}
	if len(m.IbcDenom) > 0 {
		i -= len(m.IbcDenom)
		copy(dAtA[i:], m.IbcDenom)
		i = encodeVarintSourceChain(dAtA, i, uint64(len(m.IbcDenom)))
		i--
		dAtA[i] = 0x52
	}
	{
		size := m.Redemptionratio.Size()
		i -= size
		if _, err := m.Redemptionratio.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintSourceChain(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x4a
	if len(m.UnboudAddress) > 0 {
		i -= len(m.UnboudAddress)
		copy(dAtA[i:], m.UnboudAddress)
		i = encodeVarintSourceChain(dAtA, i, uint64(len(m.UnboudAddress)))
		i--
		dAtA[i] = 0x42
	}
	if len(m.DelegateAddress) > 0 {
		i -= len(m.DelegateAddress)
		copy(dAtA[i:], m.DelegateAddress)
		i = encodeVarintSourceChain(dAtA, i, uint64(len(m.DelegateAddress)))
		i--
		dAtA[i] = 0x3a
	}
	if len(m.WithdrawAddress) > 0 {
		i -= len(m.WithdrawAddress)
		copy(dAtA[i:], m.WithdrawAddress)
		i = encodeVarintSourceChain(dAtA, i, uint64(len(m.WithdrawAddress)))
		i--
		dAtA[i] = 0x32
	}
	if len(m.Validators) > 0 {
		for iNdEx := len(m.Validators) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Validators[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintSourceChain(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x2a
		}
	}
	if len(m.Bech32ValidatorAddrPrefix) > 0 {
		i -= len(m.Bech32ValidatorAddrPrefix)
		copy(dAtA[i:], m.Bech32ValidatorAddrPrefix)
		i = encodeVarintSourceChain(dAtA, i, uint64(len(m.Bech32ValidatorAddrPrefix)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.TrasnferChannelID) > 0 {
		i -= len(m.TrasnferChannelID)
		copy(dAtA[i:], m.TrasnferChannelID)
		i = encodeVarintSourceChain(dAtA, i, uint64(len(m.TrasnferChannelID)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.ConnectionID) > 0 {
		i -= len(m.ConnectionID)
		copy(dAtA[i:], m.ConnectionID)
		i = encodeVarintSourceChain(dAtA, i, uint64(len(m.ConnectionID)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.ChainID) > 0 {
		i -= len(m.ChainID)
		copy(dAtA[i:], m.ChainID)
		i = encodeVarintSourceChain(dAtA, i, uint64(len(m.ChainID)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintSourceChain(dAtA []byte, offset int, v uint64) int {
	offset -= sovSourceChain(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Validator) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Address)
	if l > 0 {
		n += 1 + l + sovSourceChain(uint64(l))
	}
	l = m.DelegationAmount.Size()
	n += 1 + l + sovSourceChain(uint64(l))
	if m.Weight != 0 {
		n += 1 + sovSourceChain(uint64(m.Weight))
	}
	return n
}

func (m *SourceChain) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ChainID)
	if l > 0 {
		n += 1 + l + sovSourceChain(uint64(l))
	}
	l = len(m.ConnectionID)
	if l > 0 {
		n += 1 + l + sovSourceChain(uint64(l))
	}
	l = len(m.TrasnferChannelID)
	if l > 0 {
		n += 1 + l + sovSourceChain(uint64(l))
	}
	l = len(m.Bech32ValidatorAddrPrefix)
	if l > 0 {
		n += 1 + l + sovSourceChain(uint64(l))
	}
	if len(m.Validators) > 0 {
		for _, e := range m.Validators {
			l = e.Size()
			n += 1 + l + sovSourceChain(uint64(l))
		}
	}
	l = len(m.WithdrawAddress)
	if l > 0 {
		n += 1 + l + sovSourceChain(uint64(l))
	}
	l = len(m.DelegateAddress)
	if l > 0 {
		n += 1 + l + sovSourceChain(uint64(l))
	}
	l = len(m.UnboudAddress)
	if l > 0 {
		n += 1 + l + sovSourceChain(uint64(l))
	}
	l = m.Redemptionratio.Size()
	n += 1 + l + sovSourceChain(uint64(l))
	l = len(m.IbcDenom)
	if l > 0 {
		n += 1 + l + sovSourceChain(uint64(l))
	}
	l = len(m.NativeDenom)
	if l > 0 {
		n += 1 + l + sovSourceChain(uint64(l))
	}
	l = len(m.DerivativeDenom)
	if l > 0 {
		n += 1 + l + sovSourceChain(uint64(l))
	}
	l = m.StakedAmount.Size()
	n += 1 + l + sovSourceChain(uint64(l))
	return n
}

func sovSourceChain(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozSourceChain(x uint64) (n int) {
	return sovSourceChain(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Validator) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowSourceChain
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
			return fmt.Errorf("proto: Validator: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Validator: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Address", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSourceChain
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
				return ErrInvalidLengthSourceChain
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthSourceChain
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Address = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DelegationAmount", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSourceChain
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
				return ErrInvalidLengthSourceChain
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthSourceChain
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.DelegationAmount.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Weight", wireType)
			}
			m.Weight = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSourceChain
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Weight |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipSourceChain(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthSourceChain
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
func (m *SourceChain) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowSourceChain
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
			return fmt.Errorf("proto: SourceChain: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SourceChain: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ChainID", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSourceChain
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
				return ErrInvalidLengthSourceChain
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthSourceChain
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ChainID = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ConnectionID", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSourceChain
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
				return ErrInvalidLengthSourceChain
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthSourceChain
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ConnectionID = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TrasnferChannelID", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSourceChain
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
				return ErrInvalidLengthSourceChain
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthSourceChain
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.TrasnferChannelID = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Bech32ValidatorAddrPrefix", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSourceChain
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
				return ErrInvalidLengthSourceChain
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthSourceChain
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Bech32ValidatorAddrPrefix = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Validators", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSourceChain
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
				return ErrInvalidLengthSourceChain
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthSourceChain
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Validators = append(m.Validators, &Validator{})
			if err := m.Validators[len(m.Validators)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field WithdrawAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSourceChain
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
				return ErrInvalidLengthSourceChain
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthSourceChain
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.WithdrawAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DelegateAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSourceChain
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
				return ErrInvalidLengthSourceChain
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthSourceChain
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.DelegateAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 8:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field UnboudAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSourceChain
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
				return ErrInvalidLengthSourceChain
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthSourceChain
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.UnboudAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 9:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Redemptionratio", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSourceChain
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
				return ErrInvalidLengthSourceChain
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthSourceChain
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Redemptionratio.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 10:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field IbcDenom", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSourceChain
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
				return ErrInvalidLengthSourceChain
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthSourceChain
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.IbcDenom = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 11:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field NativeDenom", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSourceChain
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
				return ErrInvalidLengthSourceChain
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthSourceChain
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.NativeDenom = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 12:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DerivativeDenom", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSourceChain
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
				return ErrInvalidLengthSourceChain
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthSourceChain
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.DerivativeDenom = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 13:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field StakedAmount", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSourceChain
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
				return ErrInvalidLengthSourceChain
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthSourceChain
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.StakedAmount.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipSourceChain(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthSourceChain
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
func skipSourceChain(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowSourceChain
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
					return 0, ErrIntOverflowSourceChain
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
					return 0, ErrIntOverflowSourceChain
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
				return 0, ErrInvalidLengthSourceChain
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupSourceChain
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthSourceChain
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthSourceChain        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowSourceChain          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupSourceChain = fmt.Errorf("proto: unexpected end of group")
)
