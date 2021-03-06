// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: github.com/rkcloudchain/rksync/protos/identity.proto

package protos

import (
	fmt "fmt"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/golang/protobuf/proto"
	io "io"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type SerializedIdentity struct {
	NodeId               string   `protobuf:"bytes,1,opt,name=node_id,json=nodeId,proto3" json:"node_id,omitempty"`
	IdBytes              []byte   `protobuf:"bytes,2,opt,name=id_bytes,json=idBytes,proto3" json:"id_bytes,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SerializedIdentity) Reset()         { *m = SerializedIdentity{} }
func (m *SerializedIdentity) String() string { return proto.CompactTextString(m) }
func (*SerializedIdentity) ProtoMessage()    {}
func (*SerializedIdentity) Descriptor() ([]byte, []int) {
	return fileDescriptor_bfba01b5c7988b92, []int{0}
}
func (m *SerializedIdentity) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *SerializedIdentity) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_SerializedIdentity.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalTo(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *SerializedIdentity) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SerializedIdentity.Merge(m, src)
}
func (m *SerializedIdentity) XXX_Size() int {
	return m.Size()
}
func (m *SerializedIdentity) XXX_DiscardUnknown() {
	xxx_messageInfo_SerializedIdentity.DiscardUnknown(m)
}

var xxx_messageInfo_SerializedIdentity proto.InternalMessageInfo

func init() {
	proto.RegisterType((*SerializedIdentity)(nil), "protos.SerializedIdentity")
}

func init() {
	proto.RegisterFile("github.com/rkcloudchain/rksync/protos/identity.proto", fileDescriptor_bfba01b5c7988b92)
}

var fileDescriptor_bfba01b5c7988b92 = []byte{
	// 192 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x32, 0x49, 0xcf, 0x2c, 0xc9,
	0x28, 0x4d, 0xd2, 0x4b, 0xce, 0xcf, 0xd5, 0x2f, 0xca, 0x4e, 0xce, 0xc9, 0x2f, 0x4d, 0x49, 0xce,
	0x48, 0xcc, 0xcc, 0xd3, 0x2f, 0xca, 0x2e, 0xae, 0xcc, 0x4b, 0xd6, 0x2f, 0x28, 0xca, 0x2f, 0xc9,
	0x2f, 0xd6, 0xcf, 0x4c, 0x49, 0xcd, 0x2b, 0xc9, 0x2c, 0xa9, 0xd4, 0x03, 0xf3, 0x85, 0xd8, 0x20,
	0xc2, 0x52, 0x22, 0xe9, 0xf9, 0xe9, 0xf9, 0x60, 0xb6, 0x3e, 0x88, 0x05, 0x91, 0x55, 0xf2, 0xe0,
	0x12, 0x0a, 0x4e, 0x2d, 0xca, 0x4c, 0xcc, 0xc9, 0xac, 0x4a, 0x4d, 0xf1, 0x84, 0xea, 0x14, 0x12,
	0xe7, 0x62, 0xcf, 0xcb, 0x4f, 0x49, 0x8d, 0xcf, 0x4c, 0x91, 0x60, 0x54, 0x60, 0xd4, 0xe0, 0x0c,
	0x62, 0x03, 0x71, 0x3d, 0x53, 0x84, 0x24, 0xb9, 0x38, 0x32, 0x53, 0xe2, 0x93, 0x2a, 0x4b, 0x52,
	0x8b, 0x25, 0x98, 0x14, 0x18, 0x35, 0x78, 0x82, 0xd8, 0x33, 0x53, 0x9c, 0x40, 0x5c, 0x27, 0xfb,
	0x13, 0x0f, 0xe5, 0x18, 0x2e, 0x3c, 0x94, 0x63, 0x3c, 0xf1, 0x48, 0x8e, 0xf1, 0xc2, 0x23, 0x39,
	0xc6, 0x07, 0x8f, 0xe4, 0x18, 0x67, 0x3c, 0x96, 0x63, 0x88, 0x52, 0x25, 0xca, 0xdd, 0x49, 0x10,
	0x87, 0x1a, 0x03, 0x02, 0x00, 0x00, 0xff, 0xff, 0x71, 0xcd, 0xba, 0x9b, 0xe7, 0x00, 0x00, 0x00,
}

func (m *SerializedIdentity) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SerializedIdentity) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.NodeId) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintIdentity(dAtA, i, uint64(len(m.NodeId)))
		i += copy(dAtA[i:], m.NodeId)
	}
	if len(m.IdBytes) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintIdentity(dAtA, i, uint64(len(m.IdBytes)))
		i += copy(dAtA[i:], m.IdBytes)
	}
	if m.XXX_unrecognized != nil {
		i += copy(dAtA[i:], m.XXX_unrecognized)
	}
	return i, nil
}

func encodeVarintIdentity(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *SerializedIdentity) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.NodeId)
	if l > 0 {
		n += 1 + l + sovIdentity(uint64(l))
	}
	l = len(m.IdBytes)
	if l > 0 {
		n += 1 + l + sovIdentity(uint64(l))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func sovIdentity(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozIdentity(x uint64) (n int) {
	return sovIdentity(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *SerializedIdentity) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowIdentity
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
			return fmt.Errorf("proto: SerializedIdentity: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SerializedIdentity: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field NodeId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowIdentity
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
				return ErrInvalidLengthIdentity
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthIdentity
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.NodeId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field IdBytes", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowIdentity
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthIdentity
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthIdentity
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.IdBytes = append(m.IdBytes[:0], dAtA[iNdEx:postIndex]...)
			if m.IdBytes == nil {
				m.IdBytes = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipIdentity(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthIdentity
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthIdentity
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipIdentity(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowIdentity
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
					return 0, ErrIntOverflowIdentity
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowIdentity
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
				return 0, ErrInvalidLengthIdentity
			}
			iNdEx += length
			if iNdEx < 0 {
				return 0, ErrInvalidLengthIdentity
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowIdentity
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipIdentity(dAtA[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
				if iNdEx < 0 {
					return 0, ErrInvalidLengthIdentity
				}
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthIdentity = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowIdentity   = fmt.Errorf("proto: integer overflow")
)
