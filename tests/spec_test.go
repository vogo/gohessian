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

package tests

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

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
func TestLong(t *testing.T) {
	var l8 int64 = -8
	var l15 int64 = 15
	var zero int64 = 0xe0

	assert.True(t, byte(l8+zero) == 0xd8)
	assert.True(t, byte(l15+zero) == 0xef)
}

func TestInt(t *testing.T) {
	var l8 int64 = -8
	var l15 int64 = 15
	var zero int64 = 0xe0

	assert.True(t, byte(l8+zero) == 0xd8)
	assert.True(t, byte(l15+zero) == 0xef)
}
