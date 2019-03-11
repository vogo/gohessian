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
	"encoding/binary"
	"math"
	"unsafe"
)

const (
	_doubleLongStartTag = byte('D') // IEEE 64-bit double
	_doubleZeroTag      = byte(0x5b)
	_doubleOneTag       = byte(0x5c)
	_doubleOneByteTag   = byte(0x5d)
	_doubleTwoByteTag   = byte(0x5e)
	_doubleFourByteTag  = byte(0x5f)
	_doubleOneByteMin   = -0x80   // -128
	_doubleOneByteMax   = -0x7f   // 127
	_doubleTwoByteMin   = -0x8000 // -32768.0
	_doubleTwoByteMax   = -0x7fff // 32767.0
)

func doubleTag(tag byte) bool {
	switch tag {
	case _long4ByteStartTag, _doubleFourByteTag, _doubleZeroTag, _doubleOneTag, _doubleOneByteTag, _doubleTwoByteTag, _doubleLongStartTag:
		return true
	default:
		return false
	}
}

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

		if iv >= _doubleOneByteMin && iv <= _doubleOneByteMax {
			return []byte{_doubleOneByteTag, byte(int8(iv))}, nil
		}

		if iv >= _doubleTwoByteMin && iv <= _doubleTwoByteMax {
			return []byte{_doubleTwoByteTag, byte(iv >> 8), byte(iv)}, nil
		}
		return nil, newCodecError("encodeDouble", "unsupported double range: %v", iv)
	}

	f32 := float32(value)
	f3264 := float64(f32)
	if f3264 == value {
		bits := math.Float32bits(f32)
		return []byte{_doubleFourByteTag,
			byte(bits >> 24),
			byte(bits >> 16),
			byte(bits >> 8),
			byte(bits)}, nil
	}

	bits := uint64(math.Float64bits(value))
	return []byte{_doubleLongStartTag,
		byte(bits >> 56),
		byte(bits >> 48),
		byte(bits >> 40),
		byte(bits >> 32),
		byte(bits >> 24),
		byte(bits >> 16),
		byte(bits >> 8),
		byte(bits)}, nil
}

func decodeDouble(reader ByteRuneReader) (float64, error) {
	return decodeDoubleValue(reader, _tagRead)
}

func decodeDoubleValue(reader ByteRuneReader, flag int32) (float64, error) {
	tag, err := getTag(reader, flag)
	if err != nil {
		return 0, err
	}

	switch tag {
	case _doubleZeroTag:
		return float64(0), nil
	case _doubleOneTag:
		return float64(1), nil
	case _doubleOneByteTag:
		bt, err := readTag(reader)
		if err != nil {
			return 0, err
		}
		return float64(int8(bt)), nil
	case _doubleTwoByteTag:
		bf, err := readBytes(reader, 2)
		if err != nil {
			return 0, err
		}
		u16 := binary.BigEndian.Uint16(bf)
		i16 := *(*int16)(unsafe.Pointer(&u16))
		return float64(i16), nil
	case _doubleFourByteTag:
		buf, err := readBytes(reader, 4)
		if err != nil {
			return 0, err
		}
		bits := binary.BigEndian.Uint32(buf)
		f32 := math.Float32frombits(bits)
		return float64(f32), nil
	case _doubleLongStartTag:
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
