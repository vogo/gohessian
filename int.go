// Copyright 2018 vogo.
// Author: wongoo
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

// ----------------------------
// see: http://hessian.caucho.com/doc/hessian-serialization.html##int
//
// int ::= 'I' b3 b2 b1 b0
// ::= [x80-xbf]
// ::= [xc0-xcf] b0
// ::= [xd0-xd7] b1 b0
//
// single octet integers
// Integers between -16 and 47 can be encoded by a single octet in the range x80 to xbf.
// 	    value = code - 0x90
//
//
// two octet integers
// Integers between -2048 and 2047 can be encoded in two octets with the leading byte in the range xc0 to xcf.
// 		value = ((code - 0xc8) << 8) + b0;
//
//
// three octet integers
// Integers between -262144 and 262143 can be encoded in three bytes with the leading byte in the range xd0 to xd7.
// 	    value = ((code - 0xd4) << 16) + (b1 << 8) + b0;
//
// four octet integers
// A 32-bit signed integer. An integer is represented by the octet x49 ('I') followed by the 4 octets of the integer in big-endian order.
// 	    value = (b3 << 24) + (b2 << 16) + (b1 << 8) + b0;

package hessian

import (
	"encoding/binary"
	"fmt"
	"io"
	"unsafe"
)

const (
	Int1ByteTagMin    = 0x80
	Int1ByteTagMax    = 0xbf
	Int1ByteZero      = byte(0x90)
	Int1ByteZeroInt32 = int32(Int1ByteZero)
	Int1ByteValueMin  = int32(-16)
	Int1ByteValueMax  = int32(47)

	Int2ByteTagMin    = 0xc0
	Int2ByteTagMax    = 0xcf
	Int2ByteZero      = byte(0xc8)
	Int2ByteZeroInt32 = int32(Int2ByteZero)
	Int2ByteValueMin  = int32(-2048)
	Int2ByteValueMax  = int32(2047)

	Int3ByteTagMin    = 0xd0
	Int3ByteTagMax    = 0xd7
	Int3ByteZero      = byte(0xd4)
	Int3ByteZeroInt32 = int32(Int3ByteZero)
	Int3ByteValueMin  = int32(-262144)
	Int3ByteValueMax  = int32(262143)

	Int4ByteStartTag = byte('I')
)

func IntTag(tag byte) bool {
	return (tag >= Int1ByteTagMin && tag <= Int1ByteTagMax) ||
		(tag >= Int2ByteTagMin && tag <= Int2ByteTagMax) ||
		(tag >= Int3ByteTagMin && tag <= Int3ByteTagMax) ||
		(tag == Int4ByteStartTag)
}

func encodeInt(value int32) []byte {
	if Int1ByteValueMin <= value && value <= Int1ByteValueMax {
		return []byte{byte(Int1ByteZeroInt32 + value)}
	}

	if Int2ByteValueMin <= value && value <= Int2ByteValueMax {
		return []byte{
			byte(Int2ByteZeroInt32 + value>>8),
			byte(value)}
	}

	if Int3ByteValueMin <= value && value <= Int3ByteValueMax {
		return []byte{
			byte(Int3ByteZeroInt32 + value>>16),
			byte(value >> 8),
			byte(value)}
	}

	return []byte{
		Int4ByteStartTag,
		byte(value >> 24),
		byte(value >> 16),
		byte(value >> 8),
		byte(value)}
}

func decodeInt(reader io.Reader) (int32, error) {
	return decodeIntValue(reader, TagRead)
}

func decodeIntValue(reader io.Reader, flag int32) (int32, error) {
	tag, err := getTag(reader, flag)
	if err != nil {
		return 0, err
	}

	if tag >= Int1ByteTagMin && tag <= Int1ByteTagMax {
		u8 := uint8(tag - Int1ByteZero)
		i8 := *(*int8)(unsafe.Pointer(&u8))
		return int32(i8), nil
	}

	if tag >= Int2ByteTagMin && tag <= Int2ByteTagMax {
		bf, err := readBytes(reader, 1)
		if err != nil {
			return 0, err
		}
		by := []byte{byte(tag - Int2ByteZero), bf[0]}
		u16 := binary.BigEndian.Uint16(by)
		i16 := *(*int16)(unsafe.Pointer(&u16))
		return int32(i16), nil
	}

	if tag >= Int3ByteTagMin && tag <= Int3ByteTagMax {
		bf, err := readBytes(reader, 2)
		if err != nil {
			return 0, err
		}
		b := byte(tag - Int3ByteZero)
		var fb byte
		if b&0x08 > 0 {
			fb = 0xFF
		} else {
			fb = 0x00
		}
		by := []byte{fb, b, bf[0], bf[1]}
		u32 := binary.BigEndian.Uint32(by)
		i32 := *(*int32)(unsafe.Pointer(&u32))
		return i32, nil
	}

	if tag == Int4ByteStartTag {
		buf, err := readBytes(reader, 4)
		if err != nil {
			return 0, err
		}
		by := []byte{buf[0], buf[1], buf[2], buf[3]}
		u32 := binary.BigEndian.Uint32(by)
		i32 := *(*int32)(unsafe.Pointer(&u32))
		return i32, nil
	}

	return 0, fmt.Errorf("wrong int tag: %x", tag)
}
