// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: processedBlockNonce.proto

package esdtSupply

import (
	fmt "fmt"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	io "io"
	math "math"
	math_bits "math/bits"
	reflect "reflect"
	strings "strings"
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

// ProcessedBlock is used to store nonce of the latest processed block
type ProcessedBlockNonce struct {
	Nonce uint64 `protobuf:"varint,1,opt,name=Nonce,proto3" json:"Nonce,omitempty"`
}

func (m *ProcessedBlockNonce) Reset()      { *m = ProcessedBlockNonce{} }
func (*ProcessedBlockNonce) ProtoMessage() {}
func (*ProcessedBlockNonce) Descriptor() ([]byte, []int) {
	return fileDescriptor_5937e333d7eca260, []int{0}
}
func (m *ProcessedBlockNonce) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *ProcessedBlockNonce) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	b = b[:cap(b)]
	n, err := m.MarshalToSizedBuffer(b)
	if err != nil {
		return nil, err
	}
	return b[:n], nil
}
func (m *ProcessedBlockNonce) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ProcessedBlockNonce.Merge(m, src)
}
func (m *ProcessedBlockNonce) XXX_Size() int {
	return m.Size()
}
func (m *ProcessedBlockNonce) XXX_DiscardUnknown() {
	xxx_messageInfo_ProcessedBlockNonce.DiscardUnknown(m)
}

var xxx_messageInfo_ProcessedBlockNonce proto.InternalMessageInfo

func (m *ProcessedBlockNonce) GetNonce() uint64 {
	if m != nil {
		return m.Nonce
	}
	return 0
}

func init() {
	proto.RegisterType((*ProcessedBlockNonce)(nil), "proto.ProcessedBlockNonce")
}

func init() { proto.RegisterFile("processedBlockNonce.proto", fileDescriptor_5937e333d7eca260) }

var fileDescriptor_5937e333d7eca260 = []byte{
	// 185 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x92, 0x2c, 0x28, 0xca, 0x4f,
	0x4e, 0x2d, 0x2e, 0x4e, 0x4d, 0x71, 0xca, 0xc9, 0x4f, 0xce, 0xf6, 0xcb, 0xcf, 0x4b, 0x4e, 0xd5,
	0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x05, 0x53, 0x52, 0xba, 0xe9, 0x99, 0x25, 0x19, 0xa5,
	0x49, 0x7a, 0xc9, 0xf9, 0xb9, 0xfa, 0xe9, 0xf9, 0xe9, 0xf9, 0xfa, 0x60, 0xe1, 0xa4, 0xd2, 0x34,
	0x30, 0x0f, 0xcc, 0x01, 0xb3, 0x20, 0xba, 0x94, 0xb4, 0xb9, 0x84, 0x03, 0x30, 0x8d, 0x14, 0x12,
	0xe1, 0x62, 0x05, 0x33, 0x24, 0x18, 0x15, 0x18, 0x35, 0x58, 0x82, 0x20, 0x1c, 0x27, 0x97, 0x0b,
	0x0f, 0xe5, 0x18, 0x6e, 0x3c, 0x94, 0x63, 0xf8, 0xf0, 0x50, 0x8e, 0xb1, 0xe1, 0x91, 0x1c, 0xe3,
	0x8a, 0x47, 0x72, 0x8c, 0x27, 0x1e, 0xc9, 0x31, 0x5e, 0x78, 0x24, 0xc7, 0x78, 0xe3, 0x91, 0x1c,
	0xe3, 0x83, 0x47, 0x72, 0x8c, 0x2f, 0x1e, 0xc9, 0x31, 0x7c, 0x78, 0x24, 0xc7, 0x38, 0xe1, 0xb1,
	0x1c, 0xc3, 0x85, 0xc7, 0x72, 0x0c, 0x37, 0x1e, 0xcb, 0x31, 0x44, 0x71, 0xa5, 0x16, 0xa7, 0x94,
	0x04, 0x97, 0x16, 0x14, 0xe4, 0x54, 0x26, 0xb1, 0x81, 0x6d, 0x36, 0x06, 0x04, 0x00, 0x00, 0xff,
	0xff, 0xdc, 0xa4, 0x39, 0x57, 0xcc, 0x00, 0x00, 0x00,
}

func (this *ProcessedBlockNonce) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*ProcessedBlockNonce)
	if !ok {
		that2, ok := that.(ProcessedBlockNonce)
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
	if this.Nonce != that1.Nonce {
		return false
	}
	return true
}
func (this *ProcessedBlockNonce) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 5)
	s = append(s, "&esdtSupply.ProcessedBlockNonce{")
	s = append(s, "Nonce: "+fmt.Sprintf("%#v", this.Nonce)+",\n")
	s = append(s, "}")
	return strings.Join(s, "")
}
func valueToGoStringProcessedBlockNonce(v interface{}, typ string) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("func(v %v) *%v { return &v } ( %#v )", typ, typ, pv)
}
func (m *ProcessedBlockNonce) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ProcessedBlockNonce) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *ProcessedBlockNonce) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Nonce != 0 {
		i = encodeVarintProcessedBlockNonce(dAtA, i, uint64(m.Nonce))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintProcessedBlockNonce(dAtA []byte, offset int, v uint64) int {
	offset -= sovProcessedBlockNonce(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *ProcessedBlockNonce) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Nonce != 0 {
		n += 1 + sovProcessedBlockNonce(uint64(m.Nonce))
	}
	return n
}

func sovProcessedBlockNonce(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozProcessedBlockNonce(x uint64) (n int) {
	return sovProcessedBlockNonce(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (this *ProcessedBlockNonce) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&ProcessedBlockNonce{`,
		`Nonce:` + fmt.Sprintf("%v", this.Nonce) + `,`,
		`}`,
	}, "")
	return s
}
func valueToStringProcessedBlockNonce(v interface{}) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("*%v", pv)
}
func (m *ProcessedBlockNonce) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowProcessedBlockNonce
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
			return fmt.Errorf("proto: ProcessedBlockNonce: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ProcessedBlockNonce: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Nonce", wireType)
			}
			m.Nonce = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProcessedBlockNonce
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Nonce |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipProcessedBlockNonce(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthProcessedBlockNonce
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthProcessedBlockNonce
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
func skipProcessedBlockNonce(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowProcessedBlockNonce
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
					return 0, ErrIntOverflowProcessedBlockNonce
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
					return 0, ErrIntOverflowProcessedBlockNonce
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
				return 0, ErrInvalidLengthProcessedBlockNonce
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupProcessedBlockNonce
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthProcessedBlockNonce
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthProcessedBlockNonce        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowProcessedBlockNonce          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupProcessedBlockNonce = fmt.Errorf("proto: unexpected end of group")
)
