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
	"crypto/rand"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestBinaryLength(t *testing.T) {
	chunkSize := BinaryChunkSize
	b1 := byte(chunkSize >> 8)
	b0 := byte(chunkSize)

	size := int(b1)<<8 + int(b0)
	assert.Equal(t, chunkSize, size)
}

func TestBinary(t *testing.T) {
	binaryTest(t, nil)
	binaryTest(t, []byte{})

	binaryLengthTest(t, BinaryShortTagMaxLen)
	binaryLengthTest(t, BinaryShortTagMaxLen-5)
	binaryLengthTest(t, BinaryShortTagMaxLen+5)

	binaryLengthTest(t, BinaryChunkSize)
	binaryLengthTest(t, BinaryChunkSize+BinaryShortTagMaxLen)
	binaryLengthTest(t, BinaryChunkSize+BinaryShortTagMaxLen-5)
	binaryLengthTest(t, BinaryChunkSize+BinaryShortTagMaxLen+5)
	binaryLengthTest(t, BinaryChunkSize*2)
	binaryLengthTest(t, BinaryChunkSize*3+123)
	binaryLengthTest(t, BinaryChunkSize*4+1234)
}

func binaryLengthTest(t *testing.T, length int) {
	buf := make([]byte, length)
	_, err := rand.Read(buf)
	assert.Nil(t, err)

	binaryTest(t, buf)
}

func binaryTest(t *testing.T, buf []byte) {
	encodeBt := encodeBinary(buf)
	assert.NotNil(t, encodeBt)

	reader := bufio.NewReader(bytes.NewReader(encodeBt))
	decodeBt, err := decodeBinary(reader)
	assert.Nil(t, err)

	if len(buf) > 0 {
		assert.True(t, reflect.DeepEqual(buf, decodeBt))
	}
}
