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

// --------------------------
// see: http://hessian.caucho.com/doc/hessian-serialization.html##string
//
// String Grammar
//
// string ::= x52 b1 b0 <utf8-data> string
//        ::= S b1 b0 <utf8-data>
//        ::= [x00-x1f] <utf8-data>
//        ::= [x30-x33] b0 <utf8-data>
//
// A 16-bit unicode character string encoded in UTF-8.
// Strings are encoded in chunks.
// x53 ('S') represents the final chunk and x52 ('R') represents any non-final chunk.
// Each chunk has a 16-bit unsigned integer length value.
//
// The length is the number of 16-bit characters, which may be different than the number of bytes.
//
// String chunks may not split surrogate pairs.
//
// short strings
// Strings with length less than 32 may be encoded with a single octet length [x00-x1f].
// 	value = code
//
// x00                 # "", empty string
// x05 hello           # "hello"
// x01 xc3 x83         # "\u00c3"
//
// S x00 x05 hello     # "hello" in long form
//
// x52 x00 x07 hello,  # "hello, world" split into two chunks
//   x05 world

package hessian

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

const (
	StringChunkSize  = 2048
	StringFinalChunk = byte('S') // final string
	StringChunk      = byte('R') // non-final string

	StringShortLenMin = byte(0x00)
	StringShortLenMax = byte(0x1f)
	StringShortMaxLen = 31

	StringMiddleLenMin = byte(0x30)
	StringMiddleLenMax = byte(0x33)
	StringMiddleMaxLen = 0x3ff
)

var (
	strChunkSize         = StringChunkSize
	StringChunkSizeBytes = []byte{byte(strChunkSize >> 8), byte(strChunkSize)}
)

func encodeString(value string) []byte {
	dataBys := []byte(value)
	length := len(dataBys)
	if length == 0 {
		return []byte{BcNull}
	}

	byteBuf := bytes.NewBuffer(nil)

	begin := 0
	// ----> chunk string
	for length > StringChunkSize {
		byteBuf.WriteByte(StringChunk)
		byteBuf.Write(StringChunkSizeBytes)

		byteBuf.Write(dataBys[begin : begin+StringChunkSize])

		length -= StringChunkSize
		begin += StringChunkSize
	}

	// ----> short string
	if length <= StringShortMaxLen {
		byteBuf.WriteByte(byte(int(StringShortLenMin) + length))
		byteBuf.Write(dataBys[begin:])
		return byteBuf.Bytes()
	}

	// ----> middle string
	if length <= StringMiddleMaxLen {
		byteBuf.WriteByte(byte((length >> 8) + int(StringMiddleLenMin)))
		byteBuf.WriteByte(byte(length))
		byteBuf.Write(dataBys[begin:])
		return byteBuf.Bytes()
	}

	// ----> final chunk string
	byteBuf.WriteByte(StringFinalChunk)
	byteBuf.WriteByte(byte(length >> 8))
	byteBuf.WriteByte(byte(length))
	byteBuf.Write(dataBys[begin:])
	return byteBuf.Bytes()
}

func decodeString(reader io.Reader) (string, error) {
	return decodeStringValue(reader, TagRead)
}

func decodeStringValue(reader io.Reader, flag int32) (string, error) {
	tag, err := getTag(reader, flag)
	if err != nil {
		return "", err
	}

	// ----> nil string
	if tag == BcNull {
		return "", nil
	}

	length, err := getStringLen(reader, tag)
	if err != nil {
		return "", err
	}

	byteBuf := bytes.NewBuffer(nil)
	buf := make([]byte, length)

	for {
		read, err := io.ReadFull(reader, buf)
		if err != nil && !strings.Contains(err.Error(), "EOF") {
			return "", err
		}
		byteBuf.Write(buf[:read])

		if stringEndTag(tag) {
			break
		}

		// ---> read next chunk
		tag, err = readTag(reader)
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}

		if !stringTag(tag) {
			return "", fmt.Errorf("error string tag: %x", tag)
		}

		newLength, err := getStringLen(reader, tag)
		if err != nil {
			return "", err
		}
		if newLength < length {
			buf = buf[:newLength]
			length = newLength
		}
	}

	return string(byteBuf.Bytes()), nil
}

func stringShortTag(tag byte) bool {
	return tag >= StringShortLenMin && tag <= StringShortLenMax
}

func stringMiddleTag(tag byte) bool {
	return tag >= StringMiddleLenMin && tag <= StringMiddleLenMax
}

func stringChunkTag(tag byte) bool {
	return tag == StringChunk || tag == StringFinalChunk
}

func stringTag(tag byte) bool {
	return stringShortTag(tag) ||
		stringMiddleTag(tag) ||
		stringChunkTag(tag)
}

func stringEndTag(tag byte) bool {
	return tag == StringFinalChunk || stringShortTag(tag) || stringMiddleTag(tag)
}

func getStringLen(reader io.Reader, tag byte) (int, error) {
	if stringShortTag(tag) {
		return int(tag - StringShortLenMin), nil
	}

	if stringMiddleTag(tag) {
		buf := make([]byte, 1)
		_, err := io.ReadFull(reader, buf)
		if err != nil {
			return -1, err
		}
		len := int(tag-StringMiddleLenMin)<<8 + int(buf[0])
		return len, nil
	}

	if stringChunkTag(tag) {
		buf, err := readBytes(reader, 2)
		if err != nil {
			return -1, err
		}
		len := int(buf[0])<<8 + int(buf[1])
		return len, nil
	}

	return -1, fmt.Errorf("err string tag: %x", tag)
}
