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
	"reflect"
	"testing"
	"time"
)

func TestDateType(t *testing.T) {
	date := time.Now()
	dateType := reflect.TypeOf(date)
	assert.Equal(t, reflect.Struct, dateType.Kind())

	assert.Equal(t, _dateType, dateType)
}

func TestDate(t *testing.T) {
	date := time.Now()
	bt := encodeDate(date)
	assert.Equal(t, 9, len(bt))
	reader := bufio.NewReader(bytes.NewReader(bt))
	d, err := decodeDate(reader)
	assert.Nil(t, err)
	assert.Equal(t, (date.UnixNano()/int64(time.Millisecond))*int64(time.Millisecond), d.UnixNano())

	nano := date.UnixNano()
	date = time.Unix(nano/int64(time.Second), 0)
	bt = encodeDate(date)
	assert.Equal(t, 5, len(bt))
	reader = bufio.NewReader(bytes.NewReader(bt))
	d, err = decodeDate(reader)
	assert.Nil(t, err)
	assert.Equal(t, date.UnixNano(), d.UnixNano())
}
