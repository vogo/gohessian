// Copyright 2018-2019 vogo.
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
	"fmt"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestIsZero(t *testing.T) {
	var p *string
	v := reflect.ValueOf(p)

	assert.True(t, v.IsValid())
	assert.True(t, IsZero(v))

	assert.Equal(t, uintptr(0), v.Pointer())

	v = v.Elem()
	assert.False(t, v.IsValid())
	assert.True(t, IsZero(v))

	var b *bool
	bv := reflect.ValueOf(b)
	assert.True(t, bv.IsValid())
	assert.True(t, IsZero(bv))

	bv = bv.Elem()
	assert.False(t, bv.IsValid())
	assert.True(t, IsZero(bv))
}

func TestExtractTypeMap(t *testing.T) {
	type ServerApi struct {
		ApiName string
	}

	type ServerNode struct {
		Name     string
		Channels []string
		ApiList  []ServerApi
		ApiMap   map[string]ServerApi
	}

	m := TypeMapFrom(ServerNode{})
	assert.NotNil(t, m)
	t.Log(m)

	_, found := m["ServerNode"]
	assert.True(t, found)

	_, found = m["ServerApi"]
	assert.True(t, found)

}

func TestTypeName(t *testing.T) {
	i := make([]interface{}, 2, 2)
	i[0] = 1
	i[1] = "hello"

	typ := reflect.TypeOf(i)
	fmt.Println(typ)
	assert.Equal(t, _interfaceTypeName, arrayRootElemName(TypeName(typ)))
}
