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

func equalName(n1, n2 *UserName) bool {
	return n1 != nil && n2 != nil && n1.FirstName == n2.FirstName && n1.LastName == n2.LastName
}

func TestHessian(t *testing.T) {
	name := UserName{
		FirstName: "John",
		LastName:  "Doe",
	}
	name2 := encodeDecode(name, t)
	fmt.Println(reflect.TypeOf(name2))
	fmt.Println(name2)

	//	if !reflect.DeepEqual(&name, name2) {
	//		t.Error("not equal from bytes")
	//		t.FailNow()
	//	}

	fmt.Println("--------------")
	person := Person{
		UserName: name,
		Age:      18,
		Sex:      true,
	}
	person2 := encodeDecode(person, t)
	fmt.Println(person2)

	//	if !cmp.Equal(&person, person2) {
	//		t.Error("not equal from bytes")
	//		t.FailNow()
	//	}

	fmt.Println("--------------")
	worker := Worker{
		Person: person,
		Job: []JOB{
			JOB{Title: "manager", Company: "google"},
			JOB{Title: "ceo", Company: "microsoft"},
		},
	}
	worker2 := encodeDecode(worker, t)
	fmt.Println(worker2)

	//	if !cmp.Equal(&worker, worker2) {
	//		t.Error("not equal from bytes")
	//		t.FailNow()
	//	}

}

func GetObjectTypeMap(typ reflect.Type, typMap map[string]reflect.Type) {

	fmt.Println("--------> ", typ.Name())
	typMap[typ.Name()] = typ
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		fmt.Println(f.Type.Kind())
		switch f.Type.Kind() {
		case reflect.Struct:
			GetObjectTypeMap(f.Type, typMap)
		case reflect.Array:
		case reflect.Slice:
			if f.Type.Elem().Kind() == reflect.Struct {
				GetObjectTypeMap(f.Type.Elem(), typMap)
			}
		}
	}
}
func encodeDecode(p interface{}, t *testing.T) interface{} {
	fmt.Println("object:", p)
	typ := reflect.TypeOf(p)
	typMap := make(map[string]reflect.Type)
	GetObjectTypeMap(typ, typMap)

	gh := NewGoHessian(typMap, nil)
	bt, err := gh.ToBytes(p)
	fmt.Println("hessian:", gh)
	if err != nil {
		t.Error(err)
		t.FailNow()
		return nil
	}
	fmt.Println("bytes:", string(bt))
	fmt.Println("bytes len:", len(bt))
	pnew, err := gh.ToObject(bt)
	if err != nil {
		t.Error(err)
		t.FailNow()
		return nil
	}
	return pnew
}
