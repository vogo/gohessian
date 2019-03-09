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
//
// ------------- Date Grammar
//
// date ::= x4a b7 b6 b5 b4 b3 b2 b1 b0
//      ::= x4b b3 b2 b1 b0
//
// Date represented by a 64-bit long of milliseconds since Jan 1 1970 00:00H, UTC.
//
// -------------- Compact: date in minutes

// The second form contains a 32-bit int of minutes since Jan 1 1970 00:00H, UTC.
//
//
// -------------- Date Examples
//
// x4a x00 x00 x00 xd0 x4b x92 x84 xb8   # 09:51:31 May 8, 1998 UTC
//
// x4b x4b x92 x0b xa0                 # 09:51:00 May 8, 1998 UTC

package hessian

import (
	"bufio"
	"encoding/binary"
	"io"
	"reflect"
	"time"
	"unsafe"
)

const (
	_dateMillisStartTag = byte(0x4a)
	_dateSecondStartTag = byte(0x4b)
)

var _zeroDate time.Time
var _dateType = reflect.TypeOf(time.Now())

func dateTag(tag byte) bool {
	return tag == _dateMillisStartTag || tag == _dateSecondStartTag
}

func encodeDate(date time.Time) []byte {
	if date.IsZero() {
		return []byte{_nilTag}
	}
	if date.UnixNano()%int64(time.Second) > 0 {
		value := date.UnixNano() / int64(time.Millisecond)

		// 8 octet longs
		return []byte{
			_dateMillisStartTag,
			byte(value >> 56),
			byte(value >> 48),
			byte(value >> 40),
			byte(value >> 32),
			byte(value >> 24),
			byte(value >> 16),
			byte(value >> 8),
			byte(value)}
	}

	value := date.Unix()
	return []byte{
		_dateSecondStartTag,
		byte(value >> 24),
		byte(value >> 16),
		byte(value >> 8),
		byte(value)}
}

func decodeDate(reader *bufio.Reader) (time.Time, error) {
	return decodeDateValue(reader, _tagRead)
}

func decodeDateValue(reader *bufio.Reader, flag int32) (time.Time, error) {
	tag, err := getTag(reader, flag)
	if err != nil {
		return _zeroDate, err
	}

	switch tag {
	case _dateMillisStartTag:
		bf := make([]byte, 8)
		if _, err := io.ReadFull(reader, bf); err != nil {
			return _zeroDate, err
		}

		by := []byte{bf[0], bf[1], bf[2], bf[3], bf[4], bf[5], bf[6], bf[7]}
		u64 := binary.BigEndian.Uint64(by)
		i64 := *(*int64)(unsafe.Pointer(&u64))
		return time.Unix(0, i64*int64(time.Millisecond)), nil
	case _dateSecondStartTag:
		buf, err := readBytes(reader, 4)
		if err != nil {
			return _zeroDate, err
		}
		by := []byte{buf[0], buf[1], buf[2], buf[3]}
		u32 := binary.BigEndian.Uint32(by)
		i32 := *(*int32)(unsafe.Pointer(&u32))
		return time.Unix(int64(i32), 0), nil
	}

	return _zeroDate, newCodecError("decodeDateValue", "error date tag: 0x%x", tag)
}
