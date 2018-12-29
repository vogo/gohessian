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
	"encoding/binary"
	"github.com/stretchr/testify/assert"
	"testing"
	"unsafe"
)

func TestLong(t *testing.T) {
	LongTest(t, -8, 1)
	LongTest(t, -3, 1)
	LongTest(t, 15, 1)
	LongTest(t, 7, 1)

	LongTest(t, -2048, 2)
	LongTest(t, -1025, 2)
	LongTest(t, 2047, 2)
	LongTest(t, 1023, 2)

	LongTest(t, -262144, 3)
	LongTest(t, -162144, 3)
	LongTest(t, 262143, 3)
	LongTest(t, 162143, 3)

	LongTest(t, -362144, 5)
	LongTest(t, -462144, 5)
	LongTest(t, 362143, 5)
	LongTest(t, 462143, 5)

	LongTest(t, -3621447777, 9)
	LongTest(t, -4621447777, 9)
	LongTest(t, 3621437777, 9)
	LongTest(t, 4621437777, 9)
}

func LongTest(t *testing.T, i64 int64, length int) {
	t.Log("--------------")
	t.Logf("i64: %d , %x", i64, i64)

	u64 := *(*uint64)(unsafe.Pointer(&i64))
	tb := make([]byte, 8)
	binary.BigEndian.PutUint64(tb, u64)
	t.Logf("u64: %d , %x", u64, u64)

	bt := encodeLong(i64)
	t.Logf(" bt: %x", bt)
	assert.Equal(t, length, len(bt))

	reader := bytes.NewReader(bt)
	d64, err := decodeLong(reader)
	assert.Nil(t, err)
	t.Logf("d64: %d , %x", d64, d64)

	assert.Equal(t, i64, d64)
}
