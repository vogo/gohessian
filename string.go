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
	"bufio"
	"bytes"
	"io"
	"strings"
)

const (
	_stringChunkSize  = 2048
	_stringFinalChunk = byte('S') // final string
	_stringChunk      = byte('R') // non-final string

	_stringShortLenMin = byte(0x00)
	_stringShortLenMax = byte(0x1f)
	_stringShortMaxLen = 31

	_stringMiddleLenMin = byte(0x30)
	_stringMiddleLenMax = byte(0x33)
	_stringMiddleMaxLen = 0x3ff
)

var (
	strChunkSize         = _stringChunkSize
	StringChunkSizeBytes = []byte{byte(strChunkSize >> 8), byte(strChunkSize)}
)

func encodeString(value string) []byte {

	if value == "" {
		return []byte{_nilTag}
	}

	dataBys := []rune(value)
	length := len(dataBys)
	byteBuf := bytes.NewBuffer(nil)

	begin := 0
	// ----> chunk string
	for length > _stringChunkSize {
		byteBuf.WriteByte(_stringChunk)
		byteBuf.Write(StringChunkSizeBytes)

		byteBuf.Write([]byte(string(dataBys[begin:begin+_stringChunkSize])))

		length -= _stringChunkSize
		begin += _stringChunkSize
	}

	// ----> short string
	if length <= _stringShortMaxLen {
		byteBuf.WriteByte(byte(int(_stringShortLenMin) + length))
		byteBuf.Write([]byte(string(dataBys[begin:])))
		return byteBuf.Bytes()
	}

	// ----> middle string
	if length <= _stringMiddleMaxLen {
		byteBuf.WriteByte(byte((length >> 8) + int(_stringMiddleLenMin)))
		byteBuf.WriteByte(byte(length))
		byteBuf.Write([]byte(string(dataBys[begin:])))
		return byteBuf.Bytes()
	}

	// ----> final chunk string
	byteBuf.WriteByte(_stringFinalChunk)
	byteBuf.WriteByte(byte(length >> 8))
	byteBuf.WriteByte(byte(length))
	byteBuf.Write([]byte(string(dataBys[begin:])))
	return byteBuf.Bytes()
}

func decodeString(reader *bufio.Reader) (string, error) {
	return decodeStringValue(reader, _tagRead)
}

func decodeStringValue(reader *bufio.Reader, flag int32) (string, error) {
	tag, err := getTag(reader, flag)
	if err != nil {
		return "", err
	}

	// ----> nil string
	if tag == _nilTag {
		return "", nil
	}

	length, err := getStringLen(reader, tag)
	if err != nil {
		return "", err
	}

	byteBuf := bytes.NewBuffer(nil)
	buf := make([]rune, length)
	for {
		read, err := readRunes(reader, buf)
		if err != nil && !strings.Contains(err.Error(), "EOF") {
			return "", err
		}
		byteBuf.Write([]byte(string(buf[:read])))

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
			return "", newCodecError("decodeStringValue", "error string tag: 0x%x", tag)
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
	return tag >= _stringShortLenMin && tag <= _stringShortLenMax
}

func stringMiddleTag(tag byte) bool {
	return tag >= _stringMiddleLenMin && tag <= _stringMiddleLenMax
}

func stringChunkTag(tag byte) bool {
	return tag == _stringChunk || tag == _stringFinalChunk
}

func stringTag(tag byte) bool {
	return stringShortTag(tag) ||
		stringMiddleTag(tag) ||
		stringChunkTag(tag)
}

func stringEndTag(tag byte) bool {
	return tag == _stringFinalChunk || stringShortTag(tag) || stringMiddleTag(tag)
}

func getStringLen(reader *bufio.Reader, tag byte) (int, error) {
	if stringShortTag(tag) {
		return int(tag - _stringShortLenMin), nil
	}

	if stringMiddleTag(tag) {
		buf := make([]byte, 1)
		_, err := io.ReadFull(reader, buf)
		if err != nil {
			return -1, err
		}
		len := int(tag-_stringMiddleLenMin)<<8 + int(buf[0])
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

	return -1, newCodecError("getStringLen", "err string tag: 0x%x", tag)

}
