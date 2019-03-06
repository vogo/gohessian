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

func initBenchmarkTest(b *testing.B) (Serializer, *personT) {
	c := complexLevelPerson()

	buffer := bytes.NewBuffer(nil)
	reader := bufio.NewReader(buffer)

	goHessian := NewGoHessian(ExtractTypeNameMap(c))
	err := goHessian.WriteObject(buffer, c)

	_, err = goHessian.ReadObject(reader)
	assert.Nil(b, err)

	err = goHessian.Write(c)
	assert.Nil(b, err)
	_, err = goHessian.Read()
	assert.Nil(b, err)

	return goHessian, c
}

func BenchmarkEncodeAndDecode(b *testing.B) {
	serializer, c := initBenchmarkTest(b)
	for i := 0; i < b.N; i++ {
		serializer.Write(c)
		serializer.Read()
	}

}
