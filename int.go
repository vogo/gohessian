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
	BcInt = byte('I') // 32-bit int

	IntDirectMax   = byte(47)
	BcIntZero      = byte(0x90)
	BcIntByteZero  = byte(0xc8)
	BcIntShortZero = byte(0xd4)

	Int32BcIntZero      = int32(BcIntZero)
	Int32BcIntByteZero  = int32(BcIntByteZero)
	Int32BcIntShortZero = int32(BcIntShortZero)
)

func encodeInt(value int32) []byte {
	if -16 <= value && value <= 47 {
		return []byte{byte(Int32BcIntZero + value)}
	}

	if -2048 <= value && value <= 2047 {
		return []byte{
			byte(Int32BcIntByteZero + value>>8),
			byte(value)}
	}

	if -262144 <= value && value <= 262143 {
		return []byte{
			byte(Int32BcIntShortZero + value>>16),
			byte(value >> 8),
			byte(value)}
	}

	return []byte{
		BcInt,
		byte(value >> 24),
		byte(value >> 16),
		byte(value >> 8),
		byte(value)}
}

func decodeInt(reader io.Reader) (int32, error) {
	return decodeIntTag(reader, TagRead)
}

func decodeIntTag(reader io.Reader, flag int32) (int32, error) {
	tag, err := getTag(reader, flag)
	if err != nil {
		return 0, err
	}

	if tag >= 0x80 && tag <= 0xbf {
		u8 := uint8(tag - BcIntZero)
		i8 := *(*int8)(unsafe.Pointer(&u8))
		return int32(i8), nil
	}

	if tag >= 0xc0 && tag <= 0xcf {
		bf := make([]byte, 1)
		if _, err := io.ReadFull(reader, bf); err != nil {
			return 0, err
		}
		by := []byte{byte(tag - BcIntByteZero), bf[0]}
		u16 := binary.BigEndian.Uint16(by)
		i16 := *(*int16)(unsafe.Pointer(&u16))
		return int32(i16), nil
	}

	if tag >= 0xd0 && tag <= 0xd7 {
		bf := make([]byte, 2)
		if _, err := io.ReadFull(reader, bf); err != nil {
			return 0, err
		}
		b := byte(tag - BcIntShortZero)
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

	if tag == BcInt {
		buf := make([]byte, 4)
		if _, err := io.ReadFull(reader, buf); err != nil {
			return 0, err
		}
		by := []byte{buf[0], buf[1], buf[2], buf[3]}
		u32 := binary.BigEndian.Uint32(by)
		i32 := *(*int32)(unsafe.Pointer(&u32))
		return i32, nil
	}

	return 0, fmt.Errorf("wrong int tag: %x", tag)

}
