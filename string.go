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

package hessian

import (
	"bytes"
)

const (
	BcString        = byte('S') // final string
	BcStringChunk   = byte('R') // non-final string
	BcStringDirect  = byte(0x00)
	StringDirectMax = byte(0x1f)
	BcStringShort   = byte(0x30)
	StringShortMax  = 0x3ff
)

func strTag(tag byte) bool {
	return (tag >= BcStringDirect && tag <= StringDirectMax) || (tag >= 0x30 && tag <= 0x34) || (tag == BcString || tag == BcStringChunk)
}

// see: http://hessian.caucho.com/doc/hessian-serialization.html##string
func encodeString(value string) []byte {
	bytesBuf := bytes.NewBuffer(nil)
	dataBys := []byte(value)
	length := len(dataBys)
	sub := 0x8000
	begin := 0

	for length > sub {
		bytesBuf.WriteByte(BcStringChunk)
		bytesBuf.WriteByte(byte(sub >> 8))
		bytesBuf.WriteByte(byte(sub))

		bytesBuf.Write(dataBys[begin : begin+sub])

		length -= sub
		begin += sub
	}

	if length == 0 {
		bytesBuf.WriteByte(BcNull)
		return bytesBuf.Bytes()
	} else if length <= int(StringDirectMax) {
		bytesBuf.WriteByte(byte(length + int(BcStringDirect)))
	} else if length <= int(StringShortMax) {
		bytesBuf.WriteByte(byte((length >> 8) + int(BcStringShort)))
		bytesBuf.WriteByte(byte(length))
	} else {
		bytesBuf.WriteByte(BcString)
		bytesBuf.WriteByte(byte(length >> 8))
		bytesBuf.WriteByte(byte(length))
	}
	bytesBuf.Write(dataBys[begin:])

	return bytesBuf.Bytes()
}
