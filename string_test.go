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
	"bytes"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"reflect"
	"testing"
)

func TestString(t *testing.T) {
	stringTest(t, "")

	stringLengthTest(t, _stringShortMaxLen)
	stringLengthTest(t, _stringShortMaxLen-5)
	stringLengthTest(t, _stringShortMaxLen+5)

	stringLengthTest(t, _stringChunkSize)
	stringLengthTest(t, _stringChunkSize*2)
	stringLengthTest(t, _stringChunkSize*3+123)
	stringLengthTest(t, _stringChunkSize*4+1234)

	stringLengthTest(t, _stringChunkSize+_stringShortMaxLen)
	stringLengthTest(t, _stringChunkSize+_stringShortMaxLen-5)
	stringLengthTest(t, _stringChunkSize+_stringShortMaxLen+5)

	stringLengthTest(t, _stringChunkSize+_stringMiddleMaxLen)
	stringLengthTest(t, _stringChunkSize+_stringMiddleMaxLen-5)
	stringLengthTest(t, _stringChunkSize+_stringMiddleMaxLen+5)
}

// Returns an int >= min, < max
func randomInt(min, max int) int {
	return min + rand.Intn(max-min)
}

// Generate a random string of A-Z chars with len = l
func randomString(len int) string {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		bytes[i] = byte(randomInt(65, 90))
	}
	return string(bytes)
}

func stringLengthTest(t *testing.T, length int) {
	stringTest(t, randomString(length))
}

func stringTest(t *testing.T, str string) {
	encodeString := encodeString(str)
	assert.NotNil(t, encodeString)

	reader := bufio.NewReader(bytes.NewReader(encodeString))
	decodeString, err := decodeString(reader)
	assert.Nil(t, err)

	equal := reflect.DeepEqual(str, decodeString)
	assert.True(t, equal)
	if !equal {
		t.Logf("expect: %s, got: %s", str, decodeString)
	}
}

func TestRuneString(t *testing.T) {
	stringTest(t, "hello world 你好世界...")
}
