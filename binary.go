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

import "bytes"

const (
	BcBinary        = byte('B')  // final chunk
	BcBinaryChunk   = byte('A')  // non-final chunk
	BcBinaryDirect  = byte(0x20) // 1-byte length binary
	BinaryDirectMax = byte(0x0f)
	BcBinaryShort   = byte(0x34) // 2-byte length binary
	BinaryShortMax  = 0x3ff      // 0-1023 binary
)

// see: http://hessian.caucho.com/doc/hessian-serialization.html##binary
func encodeBinary(value []byte) []byte {
	length := len(value)
	if length == 0 {
		return nil
	}

	byteBuf := bytes.NewBuffer(nil)
	sub := ChunkSize
	begin := 0

	for length > sub {
		byteBuf.WriteByte(byte(BcBinaryChunk))
		byteBuf.WriteByte(byte(sub >> 8))
		byteBuf.WriteByte(byte(sub))

		byteBuf.Write(value[begin : begin+ChunkSize])

		length -= ChunkSize
		begin += ChunkSize
	}

	if length <= int(BinaryDirectMax) {
		byteBuf.WriteByte(byte(int(BcBinaryDirect) + length))
	} else if length <= int(BinaryShortMax) {
		byteBuf.WriteByte(byte(int(BcBinaryShort) + length>>8))
		byteBuf.WriteByte(byte(length))
	} else {
		byteBuf.WriteByte(byte(BcBinary))
		byteBuf.WriteByte(byte(length >> 8))
		byteBuf.WriteByte(byte(length))
	}
	byteBuf.Write(value[begin:])

	return byteBuf.Bytes()
}
