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
	"encoding/binary"
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
	"unsafe"
)

func TestInt(t *testing.T) {
	doIntTest(t, -17, 2)
	doIntTest(t, -16, 1)
	doIntTest(t, -15, 1)
	doIntTest(t, -11, 1)
	doIntTest(t, 46, 1)
	doIntTest(t, 47, 1)
	doIntTest(t, 48, 2)
	doIntTest(t, 21, 1)

	doIntTest(t, -2047, 2)
	doIntTest(t, -2048, 2)
	doIntTest(t, -2049, 3)
	doIntTest(t, -1025, 2)
	doIntTest(t, 2046, 2)
	doIntTest(t, 2047, 2)
	doIntTest(t, 2048, 3)
	doIntTest(t, 1023, 2)

	doIntTest(t, -262143, 3)
	doIntTest(t, -262144, 3)
	doIntTest(t, -262145, 5)
	doIntTest(t, -162144, 3)
	doIntTest(t, 262142, 3)
	doIntTest(t, 262143, 3)
	doIntTest(t, 262144, 5)
	doIntTest(t, 162143, 3)

	doIntTest(t, -362143, 5)
	doIntTest(t, -362144, 5)
	doIntTest(t, -362145, 5)
	doIntTest(t, -462144, 5)
	doIntTest(t, 362142, 5)
	doIntTest(t, 362143, 5)
	doIntTest(t, 362144, 5)
	doIntTest(t, 462143, 5)

	doIntTest(t, math.MinInt32, 5)
	doIntTest(t, math.MaxInt32, 5)
}

func doIntTest(t *testing.T, i32 int32, length int) {
	t.Log("--------------")
	//t.Logf("i32: %d , %x", i32, i32)

	u32 := *(*uint32)(unsafe.Pointer(&i32))
	tb := make([]byte, 4)
	binary.BigEndian.PutUint32(tb, u32)
	t.Logf("u32: %d , %x", u32, u32)

	bt := encodeInt(i32)
	t.Logf(" bt: %x", bt)
	assert.Equal(t, length, len(bt))

	reader := bufio.NewReader(bytes.NewReader(bt))
	d32, err := decodeInt(reader)
	assert.Nil(t, err)
	//t.Logf("d32: %d , %x", d32, d32)

	assert.Equal(t, i32, d32)
}

func TestIntTagValue(t *testing.T) {
	var l8 int64 = -8
	var l15 int64 = 15
	var zero int64 = 0xe0

	assert.True(t, byte(l8+zero) == 0xd8)
	assert.True(t, byte(l15+zero) == 0xef)
}
