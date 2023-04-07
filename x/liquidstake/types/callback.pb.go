// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: celinium/liquidstake/v1/callback.proto

package types

import (
	fmt "fmt"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/gogo/protobuf/proto"
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

// As IBC calls are asynchronous and their acknowledgement arrival order cannot be controlled, we need a callback mechanism.
// Following the IBC communication mechanism, we can save information such as {ibcChannelID+sequence}: IBCCallback.
// When an IBC ACK is received, deserialize the args based on the CallType and execute the corresponding operation.
type IBCCallback struct {
	// The type of the callback operation.
	CallType uint32 `protobuf:"varint,1,opt,name=callType,proto3" json:"callType,omitempty"`
	// The arguments of the callback, serialized as a string.
	Args string `protobuf:"bytes,2,opt,name=args,proto3" json:"args,omitempty"`
}

func (m *IBCCallback) Reset()         { *m = IBCCallback{} }
func (m *IBCCallback) String() string { return proto.CompactTextString(m) }
func (*IBCCallback) ProtoMessage()    {}
func (*IBCCallback) Descriptor() ([]byte, []int) {
	return fileDescriptor_ac472a3659c6833c, []int{0}
}
func (m *IBCCallback) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *IBCCallback) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_IBCCallback.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *IBCCallback) XXX_Merge(src proto.Message) {
	xxx_messageInfo_IBCCallback.Merge(m, src)
}
func (m *IBCCallback) XXX_Size() int {
	return m.Size()
}
func (m *IBCCallback) XXX_DiscardUnknown() {
	xxx_messageInfo_IBCCallback.DiscardUnknown(m)
}

var xxx_messageInfo_IBCCallback proto.InternalMessageInfo

func (m *IBCCallback) GetCallType() uint32 {
	if m != nil {
		return m.CallType
	}
	return 0
}

func (m *IBCCallback) GetArgs() string {
	if m != nil {
		return m.Args
	}
	return ""
}

type DelegateCallbackArgs struct {
	// Validators with delegate funds
	Validators []Validator `protobuf:"bytes,1,rep,name=validators,proto3" json:"validators"`
}

func (m *DelegateCallbackArgs) Reset()         { *m = DelegateCallbackArgs{} }
func (m *DelegateCallbackArgs) String() string { return proto.CompactTextString(m) }
func (*DelegateCallbackArgs) ProtoMessage()    {}
func (*DelegateCallbackArgs) Descriptor() ([]byte, []int) {
	return fileDescriptor_ac472a3659c6833c, []int{1}
}
func (m *DelegateCallbackArgs) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *DelegateCallbackArgs) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_DelegateCallbackArgs.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *DelegateCallbackArgs) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DelegateCallbackArgs.Merge(m, src)
}
func (m *DelegateCallbackArgs) XXX_Size() int {
	return m.Size()
}
func (m *DelegateCallbackArgs) XXX_DiscardUnknown() {
	xxx_messageInfo_DelegateCallbackArgs.DiscardUnknown(m)
}

var xxx_messageInfo_DelegateCallbackArgs proto.InternalMessageInfo

func (m *DelegateCallbackArgs) GetValidators() []Validator {
	if m != nil {
		return m.Validators
	}
	return nil
}

type UnbondCallbackArgs struct {
	// Validators with unbond funds
	Validators []Validator `protobuf:"bytes,1,rep,name=validators,proto3" json:"validators"`
	// unbond epoch
	Epoch uint64 `protobuf:"varint,2,opt,name=epoch,proto3" json:"epoch,omitempty"`
	// unbond chain ID
	ChainID string `protobuf:"bytes,3,opt,name=chainID,proto3" json:"chainID,omitempty"`
}

func (m *UnbondCallbackArgs) Reset()         { *m = UnbondCallbackArgs{} }
func (m *UnbondCallbackArgs) String() string { return proto.CompactTextString(m) }
func (*UnbondCallbackArgs) ProtoMessage()    {}
func (*UnbondCallbackArgs) Descriptor() ([]byte, []int) {
	return fileDescriptor_ac472a3659c6833c, []int{2}
}
func (m *UnbondCallbackArgs) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *UnbondCallbackArgs) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_UnbondCallbackArgs.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *UnbondCallbackArgs) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UnbondCallbackArgs.Merge(m, src)
}
func (m *UnbondCallbackArgs) XXX_Size() int {
	return m.Size()
}
func (m *UnbondCallbackArgs) XXX_DiscardUnknown() {
	xxx_messageInfo_UnbondCallbackArgs.DiscardUnknown(m)
}

var xxx_messageInfo_UnbondCallbackArgs proto.InternalMessageInfo

func (m *UnbondCallbackArgs) GetValidators() []Validator {
	if m != nil {
		return m.Validators
	}
	return nil
}

func (m *UnbondCallbackArgs) GetEpoch() uint64 {
	if m != nil {
		return m.Epoch
	}
	return 0
}

func (m *UnbondCallbackArgs) GetChainID() string {
	if m != nil {
		return m.ChainID
	}
	return ""
}

func init() {
	proto.RegisterType((*IBCCallback)(nil), "celinium.liquidstake.v1.IBCCallback")
	proto.RegisterType((*DelegateCallbackArgs)(nil), "celinium.liquidstake.v1.DelegateCallbackArgs")
	proto.RegisterType((*UnbondCallbackArgs)(nil), "celinium.liquidstake.v1.UnbondCallbackArgs")
}

func init() {
	proto.RegisterFile("celinium/liquidstake/v1/callback.proto", fileDescriptor_ac472a3659c6833c)
}

var fileDescriptor_ac472a3659c6833c = []byte{
	// 297 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x52, 0x4b, 0x4e, 0xcd, 0xc9,
	0xcc, 0xcb, 0x2c, 0xcd, 0xd5, 0xcf, 0xc9, 0x2c, 0x2c, 0xcd, 0x4c, 0x29, 0x2e, 0x49, 0xcc, 0x4e,
	0xd5, 0x2f, 0x33, 0xd4, 0x4f, 0x4e, 0xcc, 0xc9, 0x49, 0x4a, 0x4c, 0xce, 0xd6, 0x2b, 0x28, 0xca,
	0x2f, 0xc9, 0x17, 0x12, 0x87, 0xa9, 0xd3, 0x43, 0x52, 0xa7, 0x57, 0x66, 0x28, 0x25, 0x92, 0x9e,
	0x9f, 0x9e, 0x0f, 0x56, 0xa3, 0x0f, 0x62, 0x41, 0x94, 0x4b, 0x69, 0xe1, 0x32, 0xb6, 0x38, 0xbf,
	0xb4, 0x28, 0x39, 0x35, 0x3e, 0x39, 0x23, 0x31, 0x33, 0x0f, 0xa2, 0x56, 0xc9, 0x96, 0x8b, 0xdb,
	0xd3, 0xc9, 0xd9, 0x19, 0x6a, 0x9f, 0x90, 0x14, 0x17, 0x07, 0xc8, 0xee, 0x90, 0xca, 0x82, 0x54,
	0x09, 0x46, 0x05, 0x46, 0x0d, 0xde, 0x20, 0x38, 0x5f, 0x48, 0x88, 0x8b, 0x25, 0xb1, 0x28, 0xbd,
	0x58, 0x82, 0x49, 0x81, 0x51, 0x83, 0x33, 0x08, 0xcc, 0x56, 0x4a, 0xe0, 0x12, 0x71, 0x49, 0xcd,
	0x49, 0x4d, 0x4f, 0x2c, 0x49, 0x85, 0x99, 0xe1, 0x58, 0x94, 0x5e, 0x2c, 0xe4, 0xc1, 0xc5, 0x55,
	0x96, 0x98, 0x93, 0x99, 0x92, 0x58, 0x92, 0x5f, 0x54, 0x2c, 0xc1, 0xa8, 0xc0, 0xac, 0xc1, 0x6d,
	0xa4, 0xa4, 0x87, 0xc3, 0x1b, 0x7a, 0x61, 0x30, 0xa5, 0x4e, 0x2c, 0x27, 0xee, 0xc9, 0x33, 0x04,
	0x21, 0xe9, 0x55, 0xea, 0x63, 0xe4, 0x12, 0x0a, 0xcd, 0x4b, 0xca, 0xcf, 0x4b, 0xa1, 0x8d, 0x05,
	0x42, 0x22, 0x5c, 0xac, 0xa9, 0x05, 0xf9, 0xc9, 0x19, 0x60, 0x7f, 0xb1, 0x04, 0x41, 0x38, 0x42,
	0x12, 0x5c, 0xec, 0xe0, 0x60, 0xf2, 0x74, 0x91, 0x60, 0x06, 0xfb, 0x17, 0xc6, 0x75, 0x32, 0x3b,
	0xf1, 0x48, 0x8e, 0xf1, 0xc2, 0x23, 0x39, 0xc6, 0x07, 0x8f, 0xe4, 0x18, 0x27, 0x3c, 0x96, 0x63,
	0xb8, 0xf0, 0x58, 0x8e, 0xe1, 0xc6, 0x63, 0x39, 0x86, 0x28, 0x19, 0x78, 0xb8, 0x57, 0xa0, 0x84,
	0x7c, 0x49, 0x65, 0x41, 0x6a, 0x71, 0x12, 0x1b, 0x38, 0xc0, 0x8d, 0x01, 0x01, 0x00, 0x00, 0xff,
	0xff, 0x2d, 0xa6, 0x51, 0x9a, 0xf5, 0x01, 0x00, 0x00,
}

func (m *IBCCallback) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *IBCCallback) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *IBCCallback) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Args) > 0 {
		i -= len(m.Args)
		copy(dAtA[i:], m.Args)
		i = encodeVarintCallback(dAtA, i, uint64(len(m.Args)))
		i--
		dAtA[i] = 0x12
	}
	if m.CallType != 0 {
		i = encodeVarintCallback(dAtA, i, uint64(m.CallType))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *DelegateCallbackArgs) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *DelegateCallbackArgs) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *DelegateCallbackArgs) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Validators) > 0 {
		for iNdEx := len(m.Validators) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Validators[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintCallback(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *UnbondCallbackArgs) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *UnbondCallbackArgs) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *UnbondCallbackArgs) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.ChainID) > 0 {
		i -= len(m.ChainID)
		copy(dAtA[i:], m.ChainID)
		i = encodeVarintCallback(dAtA, i, uint64(len(m.ChainID)))
		i--
		dAtA[i] = 0x1a
	}
	if m.Epoch != 0 {
		i = encodeVarintCallback(dAtA, i, uint64(m.Epoch))
		i--
		dAtA[i] = 0x10
	}
	if len(m.Validators) > 0 {
		for iNdEx := len(m.Validators) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Validators[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintCallback(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func encodeVarintCallback(dAtA []byte, offset int, v uint64) int {
	offset -= sovCallback(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *IBCCallback) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.CallType != 0 {
		n += 1 + sovCallback(uint64(m.CallType))
	}
	l = len(m.Args)
	if l > 0 {
		n += 1 + l + sovCallback(uint64(l))
	}
	return n
}

func (m *DelegateCallbackArgs) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Validators) > 0 {
		for _, e := range m.Validators {
			l = e.Size()
			n += 1 + l + sovCallback(uint64(l))
		}
	}
	return n
}

func (m *UnbondCallbackArgs) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Validators) > 0 {
		for _, e := range m.Validators {
			l = e.Size()
			n += 1 + l + sovCallback(uint64(l))
		}
	}
	if m.Epoch != 0 {
		n += 1 + sovCallback(uint64(m.Epoch))
	}
	l = len(m.ChainID)
	if l > 0 {
		n += 1 + l + sovCallback(uint64(l))
	}
	return n
}

func sovCallback(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozCallback(x uint64) (n int) {
	return sovCallback(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *IBCCallback) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowCallback
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
			return fmt.Errorf("proto: IBCCallback: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: IBCCallback: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field CallType", wireType)
			}
			m.CallType = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCallback
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.CallType |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Args", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCallback
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
				return ErrInvalidLengthCallback
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthCallback
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Args = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipCallback(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthCallback
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
func (m *DelegateCallbackArgs) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowCallback
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
			return fmt.Errorf("proto: DelegateCallbackArgs: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: DelegateCallbackArgs: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Validators", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCallback
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
				return ErrInvalidLengthCallback
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthCallback
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Validators = append(m.Validators, Validator{})
			if err := m.Validators[len(m.Validators)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipCallback(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthCallback
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
func (m *UnbondCallbackArgs) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowCallback
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
			return fmt.Errorf("proto: UnbondCallbackArgs: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: UnbondCallbackArgs: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Validators", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCallback
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
				return ErrInvalidLengthCallback
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthCallback
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Validators = append(m.Validators, Validator{})
			if err := m.Validators[len(m.Validators)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Epoch", wireType)
			}
			m.Epoch = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCallback
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Epoch |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ChainID", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCallback
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
				return ErrInvalidLengthCallback
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthCallback
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ChainID = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipCallback(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthCallback
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
func skipCallback(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowCallback
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
					return 0, ErrIntOverflowCallback
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
					return 0, ErrIntOverflowCallback
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
				return 0, ErrInvalidLengthCallback
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupCallback
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthCallback
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthCallback        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowCallback          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupCallback = fmt.Errorf("proto: unexpected end of group")
)
