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
	"bufio"
	"bytes"
	"testing"
)

// BenchmarkParallelPool benchmark for pool
// the performance with pool is not obviously better than that with no pool, so it's not recommended right now.
// Create and cache the serializer in goroutine is a better choice.
func BenchmarkParallelPool(b *testing.B) {
	c := buildSingleCircularObject()
	typeMap, nameMap := ExtractTypeNameMap(c)

	pool := NewSerializerPool(100, typeMap, nameMap)

	b.RunParallel(func(pb *testing.PB) {
		buffer := bytes.NewBuffer(nil)
		reader := bufio.NewReader(buffer)
		for pb.Next() {
			serializer := pool.Get().(Serializer)
			serializer.WriteTo(buffer, c)
			serializer.ReadFrom(reader)
			pool.Return(serializer)
		}
	})
}

func BenchmarkParallelNoPool(b *testing.B) {
	c := buildSingleCircularObject()
	typeMap, nameMap := ExtractTypeNameMap(c)

	b.RunParallel(func(pb *testing.PB) {
		buffer := bytes.NewBuffer(nil)
		reader := bufio.NewReader(buffer)
		for pb.Next() {
			serializer := NewSerializer(typeMap, nameMap)
			serializer.WriteTo(buffer, c)
			serializer.ReadFrom(reader)
		}
	})
}
