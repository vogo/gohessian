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
	BcLong          = byte('L') // 64-bit signed integer
	LongDirectMin   = -0x08
	LongDirectMax   = byte(0x0f)
	BcLongZero      = byte(0xe0)
	LongByteMin     = -0x800
	LongByteMax     = 0x7ff
	BcLongByteZero  = byte(0xf8)
	LongShortMin    = -0x40000
	LongShortMax    = 0x3ffff
	BcLongShortZero = byte(0x3c)
	BcLongInt       = byte(0x59)

	Int64LongDirectMin   = int64(LongDirectMin)
	Int64LongDirectMax   = int64(LongDirectMax)
	Int64BcLongZero      = int64(BcLongZero)
	Int64LongByteMin     = int64(LongByteMin)
	Int64LongByteMax     = int64(LongByteMax)
	Int64BcLongByteZero  = int64(BcLongByteZero)
	Int64LongShortMin    = int64(LongShortMin)
	Int64LongShortMax    = int64(LongShortMax)
	Int64BcLongShortZero = int64(BcLongShortZero)
)

// see: http://hessian.caucho.com/doc/hessian-serialization.html##long
func encodeLong(value int64) []byte {
	// 1 octet longs
	if Int64LongDirectMin <= value && value <= Int64LongDirectMax {
		return []byte{byte(Int64BcLongZero + value)}
	}

	// 2 octet longs
	if Int64LongByteMin <= value && value <= Int64LongByteMax {
		return []byte{
			byte(Int64BcLongByteZero + (value >> 8)),
			byte(value)}
	}

	// 3 octet longs
	if Int64LongShortMin <= value && value <= Int64LongShortMax {
		return []byte{
			byte(Int64BcLongShortZero + (value >> 16)),
			byte(value >> 8),
			byte(value)}
	}

	// 4 octet longs
	if math.MinInt32 <= value && value <= math.MaxInt32 {
		return []byte{
			BcLongInt,
			byte(value >> 24),
			byte(value >> 16),
			byte(value >> 8),
			byte(value)}
	}

	// 8 octet longs
	return []byte{
		'L',
		byte(value >> 64),
		byte(value >> 56),
		byte(value >> 48),
		byte(value >> 32),
		byte(value >> 24),
		byte(value >> 16),
		byte(value >> 8),
		byte(value)}
}

func decodeLong(reader io.Reader) (int64, error) {
	return decodeLongTag(reader, TagRead)
}

func decodeLongTag(reader io.Reader, flag int32) (int64, error) {
	var tag byte
	if flag == TagRead {
		bf := make([]byte, 1)
		_, err := reader.Read(bf)
		if err != nil {
			return 0, err
		}
		tag = bf[0]
	} else {
		tag = byte(flag)
	}

	// 1 octet longs
	if tag >= 0xd8 && tag <= 0xef {
		u8 := uint8(tag - BcLongZero)
		i8 := *(*int8)(unsafe.Pointer(&u8))
		return int64(i8), nil
	}

	// 2 octet longs
	if tag >= 0xf0 && tag <= 0xff {
		bf := make([]byte, 1)
		if _, err := io.ReadFull(reader, bf); err != nil {
			return 0, err
		}

		by := []byte{byte(tag - BcLongByteZero), bf[0]}
		u16 := binary.BigEndian.Uint16(by)
		i16 := *(*int16)(unsafe.Pointer(&u16))

		return int64(i16), nil
	}

	// 3 octet longs
	if tag >= 0x38 && tag <= 0x3f {
		bf := make([]byte, 2)
		if _, err := io.ReadFull(reader, bf); err != nil {
			return 0, err
		}

		b := byte(tag - BcLongShortZero)
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
	if tag == BcLongInt {
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
	if tag == BcLong {
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
