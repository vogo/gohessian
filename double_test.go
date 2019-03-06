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
	"fmt"
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

func TestDoubleBoundaryValue(t *testing.T) {
	assert.Equal(t, "-80", fmt.Sprintf("%x", -128))
	assert.Equal(t, "7f", fmt.Sprintf("%x", 127))
	assert.Equal(t, "-8000", fmt.Sprintf("%x", -32768))
	assert.Equal(t, "7fff", fmt.Sprintf("%x", 32767))
}

func TestDouble(t *testing.T) {
	doubleTest(t, 0, 1)
	doubleTest(t, 1, 1)
	doubleTest(t, 1.0, 1)

	doubleTest(t, _doubleOneByteMin, 2)
	doubleTest(t, _doubleOneByteMin+1, 2)
	doubleTest(t, _doubleOneByteMax, 2)

	doubleTest(t, _doubleTwoByteMin, 3)
	doubleTest(t, _doubleTwoByteMin+1, 3)
	doubleTest(t, _doubleTwoByteMax, 3)

	doubleTest(t, math.MaxFloat32, 5)
	doubleTest(t, math.MaxFloat32-1, 5)
	doubleTest(t, math.MaxFloat32+1, 5)

	doubleTest(t, math.MaxFloat64, 9)
	doubleTest(t, math.MaxFloat64-1, 9)
}

func doubleTest(t *testing.T, f64 float64, length int) {
	t.Log("--------------")

	bt, err := encodeDouble(f64)
	assert.Nil(t, err)
	t.Logf("%f ==> bt: 0x%x", f64, bt)
	assert.Equal(t, length, len(bt))

	reader := bufio.NewReader(bytes.NewReader(bt))
	d64, err := decodeDouble(reader)
	assert.Nil(t, err)

	assert.Equal(t, f64, d64)
}
