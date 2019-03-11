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
	_binaryChunkSize      = 4096
	_binaryFinalChunk     = byte('B')  // final chunk
	_binaryChunk          = byte('b')  // non-final chunk
	_binaryShortLenTagMin = byte(0x20) // 1-byte length binary min
	_binaryShortLenTagMax = byte(0x2f) // 1-byte length binary max
	_binaryShortTagMaxLen = int(_binaryShortLenTagMax - _binaryShortLenTagMin)
)

var (
	_binChunkSize         = _binaryChunkSize
	_binaryChunkSizeBytes = []byte{byte(_binChunkSize >> 8), byte(_binChunkSize)}
)

func encodeBinary(value []byte) []byte {
	length := len(value)
	if length == 0 {
		return []byte{_binaryShortLenTagMin}
	}

	byteBuf := bytes.NewBuffer(nil)

	// ----> chunk binary
	begin := 0
	for length > _binaryChunkSize {
		byteBuf.WriteByte(_binaryChunk)
		byteBuf.Write(_binaryChunkSizeBytes)

		byteBuf.Write(value[begin : begin+_binaryChunkSize])

		length -= _binaryChunkSize
		begin += _binaryChunkSize
	}

	// ----> short binary
	if length <= _binaryShortTagMaxLen {
		byteBuf.WriteByte(byte(int(_binaryShortLenTagMin) + length))
		byteBuf.Write(value[begin:])
		return byteBuf.Bytes()
	}

	// ----> final chunk binary
	byteBuf.WriteByte(byte(_binaryFinalChunk))
	byteBuf.WriteByte(byte(length >> 8))
	byteBuf.WriteByte(byte(length))
	byteBuf.Write(value[begin:])
	return byteBuf.Bytes()
}

func decodeBinary(reader ByteRuneReader) ([]byte, error) {
	return decodeBinaryValue(reader, _tagRead)
}

func decodeBinaryValue(reader ByteRuneReader, flag int32) ([]byte, error) {
	tag, err := getTag(reader, flag)
	if err != nil {
		return nil, err
	}

	// ----> nil binary
	if tag == _binaryShortLenTagMin {
		return nil, nil
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

		if binaryEndTag(tag) {
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
		if !binaryTag(tag) {
			return nil, fmt.Errorf("error binary tag: 0x%x", tag)
		}

		newLength, err := getBinaryLen(reader, tag)
		if err != nil {
			return nil, err
		}
		if newLength < length {
			buf = buf[:newLength]
			length = newLength
		}
	}

	return byteBuf.Bytes(), nil
}

func binaryShortTag(tag byte) bool {
	return tag >= _binaryShortLenTagMin && tag <= _binaryShortLenTagMax
}

func binaryChunkTag(tag byte) bool {
	return tag == _binaryFinalChunk || tag == _binaryChunk
}

func binaryEndTag(tag byte) bool {
	return tag == _binaryFinalChunk || binaryShortTag(tag)
}

func binaryTag(tag byte) bool {
	return binaryShortTag(tag) || binaryChunkTag(tag)
}

func getBinaryLen(reader ByteRuneReader, tag byte) (int, error) {
	if binaryShortTag(tag) {
		return int(tag - _binaryShortLenTagMin), nil
	}

	bs := make([]byte, 2)
	_, err := io.ReadFull(reader, bs)
	if err != nil {
		return 0, err
	}
	return int(bs[0])<<8 + int(bs[1]), nil
}
