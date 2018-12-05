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

package hessian

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
	Job []JOB
}

func TestHessian(t *testing.T) {
	name := UserName{
		FirstName: "John",
		LastName:  "Doe",
	}
	person := Person{
		UserName: name,
		Tags:     []string{"rich", "handsome"},
		Age:      18,
		Sex:      true,
	}

	encodeDecode(t, person, func(res interface{}) {
		t.Log(res)
		t.Log(reflect.TypeOf(res).Name())
		if value, ok := res.(reflect.Value); ok {
			decodeObject := value.Interface().(*Person)
			assert.True(t, reflect.DeepEqual(person, *decodeObject))
			return
		}
		assert.True(t, reflect.DeepEqual(person, res))
	})

	//TODO
	worker := Worker{
		Person: person,
		Job: []JOB{
			JOB{Title: "manager", Company: "google"},
			JOB{Title: "ceo", Company: "microsoft"},
		},
	}

	encodeDecode(t, worker, func(res interface{}) {
		t.Log(res)
		t.Log(reflect.TypeOf(res).Name())
		if value, ok := res.(reflect.Value); ok {
			decodeObject := value.Interface().(*Worker)
			assert.True(t, reflect.DeepEqual(worker, *decodeObject))
			return
		}
		assert.True(t, reflect.DeepEqual(worker, res))
	})
}

func encodeDecode(t *testing.T, object interface{}, testFunc func(res interface{})) {
	t.Log("object:", object)
	bt, err := Encode(object)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Log("bytes:", string(bt))
	t.Log("bytes len:", len(bt))

	typ := reflect.TypeOf(object)
	res, err := Decode(bt, typ)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	testFunc(res)
}
