// Copyright 2018 luckincoffee.
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
	"reflect"
)

//TypeMapFrom value
func TypeMapFrom(v interface{}) map[string]reflect.Type {
	return TypeMapOf(reflect.TypeOf(v))
}

//TypeMapOf type
func TypeMapOf(typ reflect.Type) map[string]reflect.Type {
	typMap := make(map[string]reflect.Type)
	FetchTypeMap(typ, typMap)
	return typMap
}

//FetchTypeMap map
func FetchTypeMap(typ reflect.Type, typMap map[string]reflect.Type) {
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return
	}
	typMap[typ.Name()] = typ
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		ft := f.Type
		if f.Type.Kind() == reflect.Array || f.Type.Kind() == reflect.Slice {
			ft = f.Type.Elem()
		}
		FetchTypeMap(ft, typMap)
	}
}
