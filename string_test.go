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
	"crypto/rand"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestString(t *testing.T) {
	stringTest(t, "")

	stringLengthTest(t, StringShortMaxLen)
	stringLengthTest(t, StringShortMaxLen-5)
	stringLengthTest(t, StringShortMaxLen+5)

	stringLengthTest(t, StringChunkSize)
	stringLengthTest(t, StringChunkSize*2)
	stringLengthTest(t, StringChunkSize*3+123)
	stringLengthTest(t, StringChunkSize*4+1234)

	stringLengthTest(t, StringChunkSize+StringShortMaxLen)
	stringLengthTest(t, StringChunkSize+StringShortMaxLen-5)
	stringLengthTest(t, StringChunkSize+StringShortMaxLen+5)

	stringLengthTest(t, StringChunkSize+StringMiddleMaxLen)
	stringLengthTest(t, StringChunkSize+StringMiddleMaxLen-5)
	stringLengthTest(t, StringChunkSize+StringMiddleMaxLen+5)
}

func stringLengthTest(t *testing.T, length int) {
	buf := make([]byte, length)
	_, err := rand.Read(buf)
	assert.Nil(t, err)
	str := string(buf)

	stringTest(t, str)
}

func stringTest(t *testing.T, str string) {
	encodeString := encodeString(str)
	assert.NotNil(t, encodeString)

	reader := bytes.NewReader(encodeString)
	decodeString, err := decodeString(reader)
	assert.Nil(t, err)

	assert.True(t, reflect.DeepEqual(str, decodeString))
}
