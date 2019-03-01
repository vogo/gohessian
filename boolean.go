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
	"bufio"
)

const (
	BoolTrueTag  = byte('T')
	BoolFalseTag = byte('F')
)

// see: http://hessian.caucho.com/doc/hessian-serialization.html##boolean
func encodeBoolean(value bool) []byte {
	buf := make([]byte, 1)
	if value {
		buf[0] = BoolTrueTag
	} else {
		buf[0] = BoolFalseTag
	}
	return buf
}

func decodeBooleanValue(reader *bufio.Reader, flag int32) (bool, error) {
	tag, err := getTag(reader, flag)
	if err != nil {
		return false, err
	}
	switch tag {
	case BoolTrueTag:
		return true, nil
	case BoolFalseTag:
		return false, nil
	}
	return false, newCodecError("decodeBooleanValue", "wrong boolean tag: %x", tag)
}
