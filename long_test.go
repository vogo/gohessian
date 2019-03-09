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

func TestLong(t *testing.T) {
	doLongTest(t, -7, 1)
	doLongTest(t, -8, 1)
	doLongTest(t, -9, 2)
	doLongTest(t, -3, 1)
	doLongTest(t, 14, 1)
	doLongTest(t, 15, 1)
	doLongTest(t, 16, 2)
	doLongTest(t, 7, 1)

	doLongTest(t, -2047, 2)
	doLongTest(t, -2048, 2)
	doLongTest(t, -2049, 3)
	doLongTest(t, -1025, 2)
	doLongTest(t, 2046, 2)
	doLongTest(t, 2047, 2)
	doLongTest(t, 2048, 3)
	doLongTest(t, 1023, 2)

	doLongTest(t, -262143, 3)
	doLongTest(t, -262144, 3)
	doLongTest(t, -262145, 5)
	doLongTest(t, -162144, 3)
	doLongTest(t, 262142, 3)
	doLongTest(t, 262143, 3)
	doLongTest(t, 262144, 5)
	doLongTest(t, 162143, 3)

	doLongTest(t, -362143, 5)
	doLongTest(t, -362144, 5)
	doLongTest(t, -362145, 5)
	doLongTest(t, -462144, 5)
	doLongTest(t, 362142, 5)
	doLongTest(t, 362143, 5)
	doLongTest(t, 362144, 5)
	doLongTest(t, 462143, 5)
	doLongTest(t, math.MinInt32, 5)
	doLongTest(t, math.MaxInt32, 5)
	doLongTest(t, int64(math.MinInt32)-1, 9)
	doLongTest(t, int64(math.MaxInt32)+1, 9)

	doLongTest(t, -3621447777, 9)
	doLongTest(t, -4621447777, 9)
	doLongTest(t, 3621437777, 9)
	doLongTest(t, 4621437777, 9)

	doLongTest(t, math.MinInt64, 9)
	doLongTest(t, math.MaxInt64, 9)
}

func doLongTest(t *testing.T, i64 int64, length int) {
	t.Log("--------------")
	// t.Logf("i64: %d , %x", i64, i64)

	u64 := *(*uint64)(unsafe.Pointer(&i64))
	tb := make([]byte, 8)
	binary.BigEndian.PutUint64(tb, u64)
	t.Logf("u64: %d , %x", u64, u64)

	bt := encodeLong(i64)
	t.Logf(" bt: %x", bt)
	assert.Equal(t, length, len(bt))

	reader := bufio.NewReader(bytes.NewReader(bt))
	d64, err := decodeLong(reader)
	assert.Nil(t, err)
	// t.Logf("d64: %d , %x", d64, d64)

	assert.Equal(t, i64, d64)
}

//Long Grammar
//
//long ::= L b7 b6 b5 b4 b3 b2 b1 b0
//::= [xd8-xef]
//::= [xf0-xff] b0
//::= [x38-x3f] b1 b0
//::= x4c b3 b2 b1 b0
// -------------------------------------
//4.7.1.  Compact: single octet longs
//Longs between -8 and 15 are represented by a single octet in the range xd8 to xef.
//
//value = (code - 0xe0)
// -------------------------------------
func TestLongTagValue(t *testing.T) {
	var l8 int64 = -8
	var l15 int64 = 15
	var zero int64 = 0xe0

	assert.True(t, byte(l8+zero) == 0xd8)
	assert.True(t, byte(l15+zero) == 0xef)
}
