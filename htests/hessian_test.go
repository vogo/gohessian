/*
 *
 *  * Copyright 2012-2016 Viant.
 *  *
 *  * Licensed under the Apache License, Version 2.0 (the "License"); you may not
 *  * use this file except in compliance with the License. You may obtain a copy of
 *  * the License at
 *  *
 *  * http://www.apache.org/licenses/LICENSE-2.0
 *  *
 *  * Unless required by applicable law or agreed to in writing, software
 *  * distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
 *  * WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
 *  * License for the specific language governing permissions and limitations under
 *  * the License.
 *
 */

package htests

import (
	"reflect"
	"testing"

	hessian "github.com/luckincoffee/gohessian"
	"github.com/stretchr/testify/assert"
)

func TestComplexStruct(t *testing.T) {
	type UserName struct {
		FirstName string
		LastName  string
	}
	type Person struct {
		UserName
		Tags []string
		Age  int32
		Sex  bool
	}

	type JOB struct {
		Title   string
		Company string
	}

	type Worker struct {
		Person
		Job        JOB   // pointer not supported
		HistoryJob []JOB // pointer not supported
	}

	name := UserName{
		FirstName: "John",
		LastName:  "Doe",
	}

	person := Person{
		UserName: name,
	}

	encodeDecode(t, person, func(res interface{}) {
		t.Log("decode person:", res)
		t.Log("type of decode person:", reflect.TypeOf(res))
		decodeObject := res.(*Person)
		assert.True(t, reflect.DeepEqual(person, *decodeObject))
	})

	person = Person{
		UserName: name,
		Tags:     []string{"rich", "handsome"},
		Age:      18,
		Sex:      true,
	}

	worker := Worker{
		Person: person,
		Job:    JOB{Title: "cto", Company: "facebook"},
		HistoryJob: []JOB{
			JOB{Title: "manager", Company: "google"},
			JOB{Title: "ceo", Company: "microsoft"},
		},
	}

	encodeDecode(t, worker, func(res interface{}) {
		t.Log("decode object:", res)
		t.Log("type of decode object:", reflect.TypeOf(res))
		decodeObject := res.(*Worker)
		assert.True(t, reflect.DeepEqual(worker, *decodeObject))
	})
}

func encodeDecode(t *testing.T, object interface{}, testFunc func(res interface{})) {
	t.Log("object:", object)
	bt, err := hessian.Encode(object)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Log("bytes:", string(bt))
	t.Log("bytes len:", len(bt))

	typ := reflect.TypeOf(object)
	t.Log("type map:", hessian.TypeMapOf(typ))
	res, err := hessian.Decode(bt, typ)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	testFunc(res)
}
