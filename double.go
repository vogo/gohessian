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

// see: http://hessian.caucho.com/doc/hessian-serialization.html##double
//
// Double Grammar
//
// double ::= D b7 b6 b5 b4 b3 b2 b1 b0
//        ::= x5b
//        ::= x5c
//        ::= x5d b0
//        ::= x5e b1 b0
//        ::= x5f b3 b2 b1 b0
//
// A 64-bit IEEE floating pointer number.
//
//
// ===> double zero
// The double 0.0 can be represented by the octet x5b
//
// ===> double one
// The double 1.0 can be represented by the octet x5c
//
// ===> double octet
// Doubles between -128.0 and 127.0 with no fractional component can be represented in two octets by casting the byte value to a double.
// 	value = (double) b0
//
// ===> double short
// Doubles between -32768.0 and 32767.0 with no fractional component can be represented in three octets by casting the short value to a double.
// 	value = (double) (256 * b1 + b0)
//
// ===> double float
// Doubles which are equivalent to their 32-bit float representation can be represented as the 4-octet float and then cast to double.

package hessian

import (
	"bufio"
	"encoding/binary"
	"math"
)

const (
	_doubleStartTag = byte('D') // IEEE 64-bit double
	_doubleZeroTag  = byte(0x5b)
	_doubleOneTag   = byte(0x5c)
	_doubleByteTag  = byte(0x5d)
	_doubleShortTag = byte(0x5e)
	_doubleMillTag  = byte(0x5f)
)

// see: http://hessian.caucho.com/doc/hessian-serialization.html##double
func encodeDouble(value float64) ([]byte, error) {
	v := float64(int64(value))
	if v == value {
		iv := int64(value)
		if iv == 0 {
			return []byte{_doubleZeroTag}, nil
		}
		if iv == 1 {
			return []byte{_doubleOneTag}, nil
		}

		if iv >= -0x80 && iv < 0x80 {
			return []byte{_doubleByteTag, byte(iv)}, nil
		}

		if iv >= -0x8000 && iv < 0x8000 {
			return []byte{_doubleByteTag, byte(iv >> 8), byte(iv)}, nil
		}
		return nil, newCodecError("encodeDouble", "unsupported double range: %v", iv)
	}

	bits := uint64(math.Float64bits(value))
	return []byte{_doubleStartTag,
		byte(bits >> 56),
		byte(bits >> 48),
		byte(bits >> 40),
		byte(bits >> 32),
		byte(bits >> 24),
		byte(bits >> 16),
		byte(bits >> 8),
		byte(bits)}, nil
}

func doubleTag(tag byte) bool {
	switch tag {
	case _long4ByteStartTag, _doubleMillTag, _doubleZeroTag, _doubleOneTag, _doubleByteTag, _doubleShortTag, _doubleStartTag:
		return true
	default:
		return false
	}
}

func decodeDoubleValue(reader *bufio.Reader, flag int32) (float64, error) {
	tag, err := getTag(reader, flag)
	if err != nil {
		return 0, err
	}

	switch tag {
	case _long4ByteStartTag, _doubleMillTag:
		i32, err := decodeInt(reader)
		if err != nil {
			return 0, err
		}
		return float64(i32), nil
	case _doubleZeroTag:
		return float64(0), nil
	case _doubleOneTag:
		return float64(1), nil
	case _doubleByteTag:
		bt, err := readTag(reader)
		if err != nil {
			return 0, err
		}
		return float64(bt), nil
	case _doubleShortTag:
		bf, err := readBytes(reader, 2)
		if err != nil {
			return 0, err
		}
		return float64(int(bf[0])*256 + int(bf[1])), nil
	case _doubleStartTag:
		buf, err := readBytes(reader, 8)
		if err != nil {
			return 0, err
		}
		bits := binary.BigEndian.Uint64(buf)
		datum := math.Float64frombits(bits)
		return datum, nil
	}

	return 0, newCodecError("decodeDoubleValue", "error double tag: 0x%x", tag)
}
