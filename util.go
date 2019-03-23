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

package hessian

import (
	"io"
)

func lowerName(name string) (string, error) {
	if name[0] >= 'a' && name[0] <= 'z' {
		return name, nil
	}
	if name[0] >= 'A' && name[0] <= 'Z' {
		bs := make([]byte, len(name))
		bs[0] = byte(name[0] + _asciiGap)
		copy(bs[1:], name[1:])
		return string(bs), nil
	}
	return name, nil
}

func capitalizeName(name string) string {
	if name[0] >= 'A' && name[0] <= 'Z' {
		return name
	}
	if name[0] >= 'a' && name[0] <= 'z' {
		bs := make([]byte, len(name))
		bs[0] = byte(name[0] - _asciiGap)
		copy(bs[1:], name[1:])
		return string(bs)
	}
	return name
}

func getTag(reader ByteRuneReader, flag int32) (byte, error) {
	if flag != _tagRead {
		return byte(flag), nil
	}
	return readTag(reader)
}

func readTag(reader ByteRuneReader) (byte, error) {
	bt, err := readBytes(reader, 1)
	if err != nil {
		return 0, err
	}
	// fmt.Printf("### read tag: %x\n", bt[0])
	return bt[0], nil
}

func readBytes(reader ByteRuneReader, length int) ([]byte, error) {
	buf := make([]byte, length)
	_, err := io.ReadFull(reader, buf)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func readRunes(reader io.RuneReader, buf []rune) (int, error) {
	i := 0
	for ; i < len(buf); i++ {
		r, _, err := reader.ReadRune()
		if err != nil {
			return i, err
		}
		buf[i] = r
	}
	return i, nil
}
