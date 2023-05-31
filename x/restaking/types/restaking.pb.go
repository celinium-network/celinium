// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: celinium/restaking/v1/restaking.proto

package types

import (
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
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

type ValidatorSetChangePacketData struct {
	ValidatorUpdates []TendermintABCIValidatorUpdate `protobuf:"bytes,1,rep,name=validator_updates,json=validatorUpdates,proto3,customtype=TendermintABCIValidatorUpdate" json:"validator_updates" yaml:"validator_updates"`
	ValsetUpdateId   uint64                          `protobuf:"varint,2,opt,name=valset_update_id,json=valsetUpdateId,proto3" json:"valset_update_id,omitempty"`
	SlashAcks        []string                        `protobuf:"bytes,3,rep,name=slash_acks,json=slashAcks,proto3" json:"slash_acks,omitempty"`
}

func (m *ValidatorSetChangePacketData) Reset()         { *m = ValidatorSetChangePacketData{} }
func (m *ValidatorSetChangePacketData) String() string { return proto.CompactTextString(m) }
func (*ValidatorSetChangePacketData) ProtoMessage()    {}
func (*ValidatorSetChangePacketData) Descriptor() ([]byte, []int) {
	return fileDescriptor_e4ee5b8db36aa17a, []int{0}
}
func (m *ValidatorSetChangePacketData) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *ValidatorSetChangePacketData) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ValidatorSetChangePacketData.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *ValidatorSetChangePacketData) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ValidatorSetChangePacketData.Merge(m, src)
}
func (m *ValidatorSetChangePacketData) XXX_Size() int {
	return m.Size()
}
func (m *ValidatorSetChangePacketData) XXX_DiscardUnknown() {
	xxx_messageInfo_ValidatorSetChangePacketData.DiscardUnknown(m)
}

var xxx_messageInfo_ValidatorSetChangePacketData proto.InternalMessageInfo

func (m *ValidatorSetChangePacketData) GetValsetUpdateId() uint64 {
	if m != nil {
		return m.ValsetUpdateId
	}
	return 0
}

func (m *ValidatorSetChangePacketData) GetSlashAcks() []string {
	if m != nil {
		return m.SlashAcks
	}
	return nil
}

type CounterPartyVersion struct {
	Version      string       `protobuf:"bytes,1,opt,name=version,proto3" json:"version,omitempty"`
	ValidatorSet ValidatorSet `protobuf:"bytes,2,opt,name=validator_set,json=validatorSet,proto3,customtype=ValidatorSet" json:"validator_set"`
}

func (m *CounterPartyVersion) Reset()         { *m = CounterPartyVersion{} }
func (m *CounterPartyVersion) String() string { return proto.CompactTextString(m) }
func (*CounterPartyVersion) ProtoMessage()    {}
func (*CounterPartyVersion) Descriptor() ([]byte, []int) {
	return fileDescriptor_e4ee5b8db36aa17a, []int{1}
}
func (m *CounterPartyVersion) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *CounterPartyVersion) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_CounterPartyVersion.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *CounterPartyVersion) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CounterPartyVersion.Merge(m, src)
}
func (m *CounterPartyVersion) XXX_Size() int {
	return m.Size()
}
func (m *CounterPartyVersion) XXX_DiscardUnknown() {
	xxx_messageInfo_CounterPartyVersion.DiscardUnknown(m)
}

var xxx_messageInfo_CounterPartyVersion proto.InternalMessageInfo

func (m *CounterPartyVersion) GetVersion() string {
	if m != nil {
		return m.Version
	}
	return ""
}

func init() {
	proto.RegisterType((*ValidatorSetChangePacketData)(nil), "celinium.restaking.v1.ValidatorSetChangePacketData")
	proto.RegisterType((*CounterPartyVersion)(nil), "celinium.restaking.v1.CounterPartyVersion")
}

func init() {
	proto.RegisterFile("celinium/restaking/v1/restaking.proto", fileDescriptor_e4ee5b8db36aa17a)
}

var fileDescriptor_e4ee5b8db36aa17a = []byte{
	// 349 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x64, 0x91, 0x4f, 0x4f, 0xc2, 0x30,
	0x18, 0xc6, 0x57, 0x31, 0x9a, 0x35, 0x68, 0x74, 0x62, 0x32, 0x89, 0x0c, 0xb2, 0xc4, 0x64, 0x27,
	0x08, 0xd1, 0x8b, 0xde, 0x18, 0x5e, 0xb8, 0x91, 0xa9, 0x1c, 0xbc, 0x2c, 0x75, 0x6b, 0x46, 0xdd,
	0xd6, 0x92, 0xb6, 0x2c, 0xf2, 0x2d, 0xfc, 0x58, 0x1c, 0xb9, 0x69, 0x3c, 0x10, 0x03, 0xdf, 0xc0,
	0x4f, 0x60, 0x58, 0xf9, 0x33, 0xe3, 0xed, 0x7d, 0x9e, 0xf7, 0xd7, 0xb7, 0x6f, 0xfb, 0xc0, 0xab,
	0x00, 0x27, 0x84, 0x92, 0x71, 0xda, 0xe2, 0x58, 0x48, 0x14, 0x13, 0x1a, 0xb5, 0xb2, 0xf6, 0x4e,
	0x34, 0x47, 0x9c, 0x49, 0x66, 0x9c, 0x6f, 0xb0, 0xe6, 0xae, 0x93, 0xb5, 0xab, 0x17, 0x01, 0x13,
	0x29, 0x13, 0x7e, 0x0e, 0xb5, 0x94, 0x50, 0x27, 0xaa, 0x95, 0x88, 0x45, 0x4c, 0xf9, 0xab, 0x4a,
	0xb9, 0xf6, 0x07, 0x80, 0x97, 0x03, 0x94, 0x90, 0x10, 0x49, 0xc6, 0x1f, 0xb0, 0xec, 0x0e, 0x11,
	0x8d, 0x70, 0x1f, 0x05, 0x31, 0x96, 0xf7, 0x48, 0x22, 0x83, 0xc2, 0xd3, 0x6c, 0xd3, 0xf7, 0xc7,
	0xa3, 0x10, 0x49, 0x2c, 0x4c, 0xd0, 0x28, 0x39, 0xba, 0xdb, 0x99, 0xce, 0xeb, 0xda, 0xd7, 0xbc,
	0x5e, 0x7b, 0xc4, 0x34, 0xc4, 0x3c, 0x25, 0x54, 0x76, 0xdc, 0x6e, 0x6f, 0x3b, 0xee, 0x29, 0xa7,
	0x7f, 0xe6, 0x75, 0x73, 0x82, 0xd2, 0xe4, 0xce, 0xfe, 0x37, 0xc7, 0xf6, 0x4e, 0xb2, 0xbf, 0xb0,
	0x30, 0x1c, 0xb8, 0xf2, 0x04, 0x96, 0x6b, 0xc8, 0x27, 0xa1, 0xb9, 0xd7, 0x00, 0xce, 0xbe, 0x77,
	0xac, 0x7c, 0x05, 0xf6, 0x42, 0xa3, 0x06, 0xa1, 0x48, 0x90, 0x18, 0xfa, 0x28, 0x88, 0x85, 0x59,
	0x5a, 0xad, 0xe4, 0xe9, 0xb9, 0xd3, 0x09, 0x62, 0x61, 0xbf, 0xc2, 0xb3, 0x2e, 0x1b, 0x53, 0x89,
	0x79, 0x1f, 0x71, 0x39, 0x19, 0x60, 0x2e, 0x08, 0xa3, 0x86, 0x09, 0x0f, 0x33, 0x55, 0x9a, 0xa0,
	0x01, 0x1c, 0xdd, 0xdb, 0x48, 0xe3, 0x16, 0x1e, 0xed, 0x36, 0x14, 0x58, 0xe6, 0xd7, 0xea, 0x6e,
	0x65, 0xfd, 0xca, 0x72, 0xf1, 0x9b, 0xbc, 0x72, 0x56, 0x50, 0xee, 0xcd, 0x74, 0x61, 0x81, 0xd9,
	0xc2, 0x02, 0xdf, 0x0b, 0x0b, 0xbc, 0x2f, 0x2d, 0x6d, 0xb6, 0xb4, 0xb4, 0xcf, 0xa5, 0xa5, 0x3d,
	0x57, 0xb7, 0x71, 0xbe, 0x15, 0x02, 0x95, 0x93, 0x11, 0x16, 0x2f, 0x07, 0x79, 0x04, 0xd7, 0xbf,
	0x01, 0x00, 0x00, 0xff, 0xff, 0x6c, 0x3d, 0x4a, 0xca, 0xf3, 0x01, 0x00, 0x00,
}

func (m *ValidatorSetChangePacketData) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ValidatorSetChangePacketData) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *ValidatorSetChangePacketData) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.SlashAcks) > 0 {
		for iNdEx := len(m.SlashAcks) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.SlashAcks[iNdEx])
			copy(dAtA[i:], m.SlashAcks[iNdEx])
			i = encodeVarintRestaking(dAtA, i, uint64(len(m.SlashAcks[iNdEx])))
			i--
			dAtA[i] = 0x1a
		}
	}
	if m.ValsetUpdateId != 0 {
		i = encodeVarintRestaking(dAtA, i, uint64(m.ValsetUpdateId))
		i--
		dAtA[i] = 0x10
	}
	if len(m.ValidatorUpdates) > 0 {
		for iNdEx := len(m.ValidatorUpdates) - 1; iNdEx >= 0; iNdEx-- {
			{
				size := m.ValidatorUpdates[iNdEx].Size()
				i -= size
				if _, err := m.ValidatorUpdates[iNdEx].MarshalTo(dAtA[i:]); err != nil {
					return 0, err
				}
				i = encodeVarintRestaking(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *CounterPartyVersion) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *CounterPartyVersion) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *CounterPartyVersion) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size := m.ValidatorSet.Size()
		i -= size
		if _, err := m.ValidatorSet.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintRestaking(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	if len(m.Version) > 0 {
		i -= len(m.Version)
		copy(dAtA[i:], m.Version)
		i = encodeVarintRestaking(dAtA, i, uint64(len(m.Version)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintRestaking(dAtA []byte, offset int, v uint64) int {
	offset -= sovRestaking(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *ValidatorSetChangePacketData) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.ValidatorUpdates) > 0 {
		for _, e := range m.ValidatorUpdates {
			l = e.Size()
			n += 1 + l + sovRestaking(uint64(l))
		}
	}
	if m.ValsetUpdateId != 0 {
		n += 1 + sovRestaking(uint64(m.ValsetUpdateId))
	}
	if len(m.SlashAcks) > 0 {
		for _, s := range m.SlashAcks {
			l = len(s)
			n += 1 + l + sovRestaking(uint64(l))
		}
	}
	return n
}

func (m *CounterPartyVersion) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Version)
	if l > 0 {
		n += 1 + l + sovRestaking(uint64(l))
	}
	l = m.ValidatorSet.Size()
	n += 1 + l + sovRestaking(uint64(l))
	return n
}

func sovRestaking(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozRestaking(x uint64) (n int) {
	return sovRestaking(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *ValidatorSetChangePacketData) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowRestaking
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
			return fmt.Errorf("proto: ValidatorSetChangePacketData: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ValidatorSetChangePacketData: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ValidatorUpdates", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRestaking
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
				return ErrInvalidLengthRestaking
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthRestaking
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			var v TendermintABCIValidatorUpdate
			m.ValidatorUpdates = append(m.ValidatorUpdates, v)
			if err := m.ValidatorUpdates[len(m.ValidatorUpdates)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ValsetUpdateId", wireType)
			}
			m.ValsetUpdateId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRestaking
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ValsetUpdateId |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SlashAcks", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRestaking
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
				return ErrInvalidLengthRestaking
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthRestaking
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.SlashAcks = append(m.SlashAcks, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipRestaking(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthRestaking
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
func (m *CounterPartyVersion) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowRestaking
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
			return fmt.Errorf("proto: CounterPartyVersion: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: CounterPartyVersion: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Version", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRestaking
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
				return ErrInvalidLengthRestaking
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthRestaking
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Version = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ValidatorSet", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRestaking
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
				return ErrInvalidLengthRestaking
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthRestaking
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.ValidatorSet.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipRestaking(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthRestaking
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
func skipRestaking(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowRestaking
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
					return 0, ErrIntOverflowRestaking
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
					return 0, ErrIntOverflowRestaking
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
				return 0, ErrInvalidLengthRestaking
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupRestaking
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthRestaking
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthRestaking        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowRestaking          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupRestaking = fmt.Errorf("proto: unexpected end of group")
)
