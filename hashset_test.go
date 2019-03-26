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
	"encoding/base64"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashSet(t *testing.T) {
	var hashsetJavaClassName = "java.util.HashSet"
	var hessianHashsetBase64 = "chFqYXZhLnV0aWwuSGFzaFNldAZjY2NkZGQGYWFhYmJi"

	data, err := base64.StdEncoding.DecodeString(hessianHashsetBase64)
	if err != nil {
		t.Error(err)
	}

	var hashsetType reflect.Type
	var hashsetHessianTypeMap map[string]reflect.Type
	var hashsetHessianNameMap map[string]string

	hashset := []string{}
	hashsetType = reflect.TypeOf(hashset)

	hashsetHessianTypeMap = TypeMapOf(hashsetType)
	hashsetHessianTypeMap[hashsetJavaClassName] = hashsetType

	hashsetHessianNameMap = make(map[string]string)
	hashsetHessianNameMap[hashsetType.Name()] = hashsetJavaClassName

	obj, err := ToObject(data, hashsetHessianTypeMap)
	if err != nil {
		t.Error(err)
	}

	t.Log(obj)

	arr, ok := obj.([]string)
	if !ok {
		t.Error("result not []string")
	}

	t.Logf("arr length:%d", len(arr))
	t.Logf("arr:%v", arr)
	assert.Equal(t, 2, len(arr))
	assert.Equal(t, "aaabbb", arr[1])
	assert.Equal(t, "cccddd", arr[0])
}
