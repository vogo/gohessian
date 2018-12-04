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
	"fmt"
	"reflect"
	"testing"
)

type UserName struct {
	FirstName string
	LastName  string
}
type Person struct {
	UserName
	Age int32
	Sex bool
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
		Age:      18,
		Sex:      true,
	}
	worker := Worker{
		Person: person,
		Job: []JOB{
			JOB{Title: "manager", Company: "google"},
			JOB{Title: "ceo", Company: "microsoft"},
		},
	}

	fmt.Println("worker:", worker)
	typ := reflect.TypeOf(worker)
	typMap := make(map[string]reflect.Type)
	InitTypeMap(typ, typMap)

	gh := NewGoHessian(typMap, nil)
	fmt.Println("hessian:", gh)

	bt, err := gh.ToBytes(worker)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	fmt.Println("bytes:", string(bt))
	fmt.Println("bytes len:", len(bt))
	res, err := gh.ToObject(bt)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	fmt.Println(res)
	fmt.Println(reflect.TypeOf(res))
}

//InitTypeMap init
func InitTypeMap(typ reflect.Type, typMap map[string]reflect.Type) {
	fmt.Println("--------> ", typ.Name())
	typMap[typ.Name()] = typ
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		fmt.Println(f.Type.Kind())
		switch f.Type.Kind() {
		case reflect.Struct:
			InitTypeMap(f.Type, typMap)
		case reflect.Array:
		case reflect.Slice:
			if f.Type.Elem().Kind() == reflect.Struct {
				InitTypeMap(f.Type.Elem(), typMap)
			}
		}
	}
}
