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
// see: http://hessian.caucho.com/doc/hessian-serialization.html##binary
//
// Binary Grammar
//
// binary ::= b b1 b0 <binary-data> binary
//        ::= B b1 b0 <binary-data>
//        ::= [x20-x2f] <binary-data>
//
// Binary data is encoded in chunks. The octet x42 ('B') encodes the final chunk
// and x62 ('b') represents any non-final chunk. Each chunk has a 16-bit // length value.
// 	len = 256 * b1 + b0
//
// short binary
// Binary data with length less than 15 may be encoded by a single octet length [x20-x2f].
// 	len = code - 0x20

package hessian

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

const (
	BinaryChunkSize    = 4096
	BcBinaryFinalChunk = byte('B')  // final chunk
	BcBinaryChunk      = byte('b')  // non-final chunk
	ShortBinaryLenMin  = byte(0x20) // 1-byte length binary min
	ShortBinaryLenMax  = byte(0x2f) // 1-byte length binary max
	ShortBinaryMaxLen  = 15
)

var (
	size                 = BinaryChunkSize
	BinaryChunkSizeBytes = []byte{byte(size >> 8), byte(size)}
)

func encodeBinary(value []byte) []byte {
	length := len(value)
	if length == 0 {
		return []byte{ShortBinaryLenMin}
	}

	byteBuf := bytes.NewBuffer(nil)

	// ----> short binary
	if length <= ShortBinaryMaxLen {
		byteBuf.WriteByte(byte(int(ShortBinaryLenMin) + length))
		byteBuf.Write(value)
		return byteBuf.Bytes()
	}

	// ----> chunk binary
	begin := 0
	for length > BinaryChunkSize {
		byteBuf.WriteByte(BcBinaryChunk)
		byteBuf.Write(BinaryChunkSizeBytes)

		byteBuf.Write(value[begin : begin+BinaryChunkSize])

		length -= BinaryChunkSize
		begin += BinaryChunkSize
	}

	byteBuf.WriteByte(byte(BcBinaryFinalChunk))
	byteBuf.WriteByte(byte(length >> 8))
	byteBuf.WriteByte(byte(length))
	byteBuf.Write(value[begin:])

	return byteBuf.Bytes()
}

func decodeBinary(reader io.Reader) ([]byte, error) {
	return decodeBinaryTag(reader, TagRead)
}

func decodeBinaryTag(reader io.Reader, flag int32) ([]byte, error) {
	tag, err := getTag(reader, flag)
	if err != nil {
		return nil, err
	}

	// ----> nil binary
	if tag == ShortBinaryLenMin {
		return nil, nil
	}

	// ----> short binary
	if tag > ShortBinaryLenMin && tag <= ShortBinaryLenMax {
		length := int(tag - ShortBinaryLenMin)
		buf := make([]byte, length)
		_, err := io.ReadFull(reader, buf)
		if err != nil {
			return nil, err
		}
		return buf, nil
	}

	// ----> chunk binary
	if !binaryChunkTag(tag) {
		return nil, fmt.Errorf("error binary tag: %x", tag)
	}

	length, err := getBinaryLen(reader, tag)
	if err != nil {
		return nil, err
	}

	byteBuf := bytes.NewBuffer(nil)
	buf := make([]byte, length)

	for {
		read, err := io.ReadFull(reader, buf)
		if err != nil && !strings.Contains(err.Error(), "EOF") {
			return nil, err
		}
		byteBuf.Write(buf[:read])

		if tag == BcBinaryFinalChunk {
			break
		}

		// ---> read next chunk
		tag, err = readTag(reader)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if !binaryChunkTag(tag) {
			return nil, fmt.Errorf("error binary tag: %x", tag)
		}

		length, err = getBinaryLen(reader, tag)
		if err != nil {
			return nil, err
		}
	}

	return byteBuf.Bytes(), nil
}

func binaryTag(tag byte) bool {
	return (tag >= ShortBinaryLenMin && tag <= ShortBinaryLenMax) || (tag == BcBinaryFinalChunk || tag == BcBinaryChunk)
}

func binaryChunkTag(tag byte) bool {
	return tag == BcBinaryFinalChunk || tag == BcBinaryChunk
}

func getBinaryLen(reader io.Reader, tag byte) (int, error) {
	if tag >= ShortBinaryLenMin && tag <= ShortBinaryLenMax {
		return int(tag - ShortBinaryLenMin), nil
	}
	bs := make([]byte, 2)
	_, err := io.ReadFull(reader, bs)
	if err != nil {
		return 0, err
	}
	return int(bs[0])<<8 + int(bs[1]), nil
}
