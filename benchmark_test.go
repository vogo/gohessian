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
	"testing"
)

func buildBenchmarkSerializer(c interface{}, b assert.TestingT) Serializer {
	buffer := bytes.NewBuffer(nil)
	reader := bufio.NewReader(buffer)
	serializer := NewSerializer(ExtractTypeNameMap(c))
	err := serializer.WriteTo(buffer, c)
	_, err = serializer.ReadFrom(reader)
	assert.Nil(b, err)
	err = serializer.Write(c)
	assert.Nil(b, err)
	_, err = serializer.Read()
	assert.Nil(b, err)
	return serializer
}

func TestBuildSerializer(t *testing.T) {
	buildBenchmarkSerializer(buildComplexLevelPerson(), t)
}

func doBenchmarkTest(b *testing.B, c interface{}) {
	serializer := buildBenchmarkSerializer(c, b)
	for i := 0; i < b.N; i++ {
		serializer.Write(c)
		serializer.Read()
	}
}

func BenchmarkComplexLevelObject(b *testing.B) {
	doBenchmarkTest(b, buildComplexLevelPerson())
}

func BenchmarkSingleCircularObject(b *testing.B) {
	doBenchmarkTest(b, buildSingleCircularObject())
}

func BenchmarkCircularObject(b *testing.B) {
	c, _, _, _ := buildCircularObject()
	doBenchmarkTest(b, c)
}
