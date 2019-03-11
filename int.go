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
	"unsafe"
)

const (
	_int1ByteTagMin    = 0x80
	_int1ByteTagMax    = 0xbf
	_int1ByteZero      = byte(0x90)
	_int1ByteZeroInt32 = int32(_int1ByteZero)
	_int1ByteValueMin  = int32(-16)
	_int1ByteValueMax  = int32(47)

	_int2ByteTagMin    = 0xc0
	_int2ByteTagMax    = 0xcf
	_int2ByteZero      = byte(0xc8)
	_int2ByteZeroInt32 = int32(_int2ByteZero)
	_int2ByteValueMin  = int32(-2048)
	_int2ByteValueMax  = int32(2047)

	_int3ByteTagMin    = 0xd0
	_int3ByteTagMax    = 0xd7
	_int3ByteZero      = byte(0xd4)
	_int3ByteZeroInt32 = int32(_int3ByteZero)
	_int3ByteValueMin  = int32(-262144)
	_int3ByteValueMax  = int32(262143)

	_int4ByteStartTag = byte('I')
)

func intTag(tag byte) bool {
	return (tag >= _int1ByteTagMin && tag <= _int1ByteTagMax) ||
		(tag >= _int2ByteTagMin && tag <= _int2ByteTagMax) ||
		(tag >= _int3ByteTagMin && tag <= _int3ByteTagMax) ||
		(tag == _int4ByteStartTag)
}

func encodeInt(value int32) []byte {
	if _int1ByteValueMin <= value && value <= _int1ByteValueMax {
		return []byte{byte(_int1ByteZeroInt32 + value)}
	}

	if _int2ByteValueMin <= value && value <= _int2ByteValueMax {
		return []byte{
			byte(_int2ByteZeroInt32 + value>>8),
			byte(value)}
	}

	if _int3ByteValueMin <= value && value <= _int3ByteValueMax {
		return []byte{
			byte(_int3ByteZeroInt32 + value>>16),
			byte(value >> 8),
			byte(value)}
	}

	return []byte{
		_int4ByteStartTag,
		byte(value >> 24),
		byte(value >> 16),
		byte(value >> 8),
		byte(value)}
}

func decodeInt(reader ByteRuneReader) (int32, error) {
	return decodeIntValue(reader, _tagRead)
}

func decodeIntValue(reader ByteRuneReader, flag int32) (int32, error) {
	tag, err := getTag(reader, flag)
	if err != nil {
		return 0, err
	}

	if tag >= _int1ByteTagMin && tag <= _int1ByteTagMax {
		u8 := uint8(tag - _int1ByteZero)
		i8 := *(*int8)(unsafe.Pointer(&u8))
		return int32(i8), nil
	}

	if tag >= _int2ByteTagMin && tag <= _int2ByteTagMax {
		bf, err := readBytes(reader, 1)
		if err != nil {
			return 0, err
		}
		by := []byte{byte(tag - _int2ByteZero), bf[0]}
		u16 := binary.BigEndian.Uint16(by)
		i16 := *(*int16)(unsafe.Pointer(&u16))
		return int32(i16), nil
	}

	if tag >= _int3ByteTagMin && tag <= _int3ByteTagMax {
		bf, err := readBytes(reader, 2)
		if err != nil {
			return 0, err
		}
		b := byte(tag - _int3ByteZero)
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

	if tag == _int4ByteStartTag {
		buf, err := readBytes(reader, 4)
		if err != nil {
			return 0, err
		}
		by := []byte{buf[0], buf[1], buf[2], buf[3]}
		u32 := binary.BigEndian.Uint32(by)
		i32 := *(*int32)(unsafe.Pointer(&u32))
		return i32, nil
	}

	return 0, newCodecError("decodeIntValue", "error int tag: 0x%x", tag)
}
