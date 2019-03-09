// Copyright 2019 vogo.
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

package main

import (
	"fmt"
	"github.com/vogo/gohessian"
)

type circular struct {
	Num      int
	Previous *circular
	Next     *circular
}

func main() {
	c := &circular{}
	c.Num = 12345
	c.Previous = c
	c.Next = c

	// create hessian serializer
	serializer := hessian.NewSerializer(hessian.ExtractTypeNameMap(c))

	fmt.Println("source object: ", c)

	// encode to bytes
	bytes, err := serializer.ToBytes(c)
	if err != nil {
		panic(err)
	}

	// decode from bytes
	decoded, err := serializer.ToObject(bytes)
	if err != nil {
		panic(err)
	}
	fmt.Println("decode object: ", decoded)
}
