// Copyright 2018 luckin coffee.
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
// see: http://hessian.caucho.com/doc/hessian-serialization.html##long
//
// Long Grammar
//
// long ::= L b7 b6 b5 b4 b3 b2 b1 b0
// ::= [xd8-xef]
// ::= [xf0-xff] b0
// ::= [x38-x3f] b1 b0
// ::= x4c b3 b2 b1 b0
//
// A 64-bit signed integer. An long is represented by the octet x4c ('L' ) followed by the 8-bytes of the integer in big-endian order.
//
// single octet longs
// Longs between -8 and 15 are represented by a single octet in the range xd8 to xef.
// 	value = (code - 0xe0)
//
// two octet longs
// Longs between -2048 and 2047 are encoded in two octets with the leading byte in the range xf0 to xff.
// 	value = ((code - 0xf8) << 8) + b0
//
// three octet longs
// Longs between -262144 and 262143 are encoded in three octets with the leading byte in the range x38 to x3f.
// 	value = ((code - 0x3c) << 16) + (b1 << 8) + b0
//
// four octet longs
// Longs between which fit into 32-bits are encoded in five octets with the leading byte x4c.
// 	value = (b3 << 24) + (b2 << 16) + (b1 << 8) + b0

package hessian

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"unsafe"
)

const (
	LongStartTag = byte('L') // 64-bit signed integer

	Long1ByteTagMin    = 0xd8
	Long1ByteTagMax    = 0xef
	Long1ByteZero      = byte(0xe0)
	Long1ByteZeroInt64 = int64(Long1ByteZero)
	Long1ByteValueMin  = -0x08      // -8
	Long1ByteValueMax  = byte(0x0f) // 15
	Long1ByteMinInt64  = int64(Long1ByteValueMin)
	Long1ByteMaxInt64  = int64(Long1ByteValueMax)

	Long2ByteTagMin        = 0xf0
	Long2ByteTagMax        = 0xff
	Long2ByteZero          = byte(0xf8)
	Long2ByteZeroInt64     = int64(Long2ByteZero)
	Long2ByteValueMin      = -0x800 // -2048
	Long2ByteValueMax      = 0x7ff  // 2047
	Long2ByteValueMinInt64 = int64(Long2ByteValueMin)
	Long2ByteValueMaxInt64 = int64(Long2ByteValueMax)

	Long3ByteTagMin        = 0x38
	Long3ByteTagMax        = 0x3f
	Long3ByteValueMin      = -0x40000 // -262144
	Long3ByteValueMax      = 0x3ffff  // 262143
	Long3ByteZero          = byte(0x3c)
	Long3ByteValueMinInt64 = int64(Long3ByteValueMin)
	Long3ByteValueMaxInt64 = int64(Long3ByteValueMax)
	Long3ByteZeroInt64     = int64(Long3ByteZero)

	Long4ByteStartTag = byte(0x59)
)

// see: http://hessian.caucho.com/doc/hessian-serialization.html##long
func encodeLong(value int64) []byte {
	// 1 octet longs
	if Long1ByteMinInt64 <= value && value <= Long1ByteMaxInt64 {
		return []byte{byte(Long1ByteZeroInt64 + value)}
	}

	// 2 octet longs
	if Long2ByteValueMinInt64 <= value && value <= Long2ByteValueMaxInt64 {
		return []byte{
			byte(Long2ByteZeroInt64 + (value >> 8)),
			byte(value)}
	}

	// 3 octet longs
	if Long3ByteValueMinInt64 <= value && value <= Long3ByteValueMaxInt64 {
		return []byte{
			byte(Long3ByteZeroInt64 + (value >> 16)),
			byte(value >> 8),
			byte(value)}
	}

	// 4 octet longs
	if math.MinInt32 <= value && value <= math.MaxInt32 {
		return []byte{
			Long4ByteStartTag,
			byte(value >> 24),
			byte(value >> 16),
			byte(value >> 8),
			byte(value)}
	}

	// 8 octet longs
	return []byte{
		LongStartTag,
		byte(value >> 56),
		byte(value >> 48),
		byte(value >> 40),
		byte(value >> 32),
		byte(value >> 24),
		byte(value >> 16),
		byte(value >> 8),
		byte(value)}
}

func LongTag(tag byte) bool {
	return (tag >= Long1ByteTagMin && tag <= Long1ByteTagMax) ||
		(tag >= Long2ByteTagMin && tag <= Long2ByteTagMax) ||
		(tag >= Long3ByteTagMin && tag <= Long3ByteTagMax) ||
		(tag == Long4ByteStartTag) ||
		(tag == LongStartTag)
}

func decodeLong(reader io.Reader) (int64, error) {
	return decodeLongValue(reader, TagRead)
}

func decodeLongValue(reader io.Reader, flag int32) (int64, error) {
	tag, err := getTag(reader, flag)
	if err != nil {
		return 0, err
	}

	// 1 octet longs
	if tag >= Long1ByteTagMin && tag <= Long1ByteTagMax {
		u8 := uint8(tag - Long1ByteZero)
		i8 := *(*int8)(unsafe.Pointer(&u8))
		return int64(i8), nil
	}

	// 2 octet longs
	if tag >= Long2ByteTagMin && tag <= Long2ByteTagMax {
		bf := make([]byte, 1)
		if _, err := io.ReadFull(reader, bf); err != nil {
			return 0, err
		}

		by := []byte{byte(tag - Long2ByteZero), bf[0]}
		u16 := binary.BigEndian.Uint16(by)
		i16 := *(*int16)(unsafe.Pointer(&u16))

		return int64(i16), nil
	}

	// 3 octet longs
	if tag >= Long3ByteTagMin && tag <= Long3ByteTagMax {
		bf := make([]byte, 2)
		if _, err := io.ReadFull(reader, bf); err != nil {
			return 0, err
		}

		b := byte(tag - Long3ByteZero)
		var fb byte
		if b&0x08 > 0 {
			fb = 0xFF
		} else {
			fb = 0x00
		}
		by := []byte{fb, b, bf[0], bf[1]}
		u32 := binary.BigEndian.Uint32(by)
		i32 := *(*int32)(unsafe.Pointer(&u32))

		return int64(i32), nil
	}

	// 4 octet longs
	if tag == Long4ByteStartTag {
		bf := make([]byte, 4)
		if _, err := io.ReadFull(reader, bf); err != nil {
			return 0, err
		}

		by := []byte{bf[0], bf[1], bf[2], bf[3]}
		u32 := binary.BigEndian.Uint32(by)
		i32 := *(*int32)(unsafe.Pointer(&u32))

		return int64(i32), nil
	}

	// 8 octet longs
	if tag == LongStartTag {
		bf := make([]byte, 8)
		if _, err := io.ReadFull(reader, bf); err != nil {
			return 0, err
		}

		by := []byte{bf[0], bf[1], bf[2], bf[3], bf[4], bf[5], bf[6], bf[7]}
		u64 := binary.BigEndian.Uint64(by)
		i64 := *(*int64)(unsafe.Pointer(&u64))

		return i64, nil
	}

	return 0, fmt.Errorf("wrong long tag: %x", tag)
}
