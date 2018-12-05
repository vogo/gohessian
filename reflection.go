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

// ServerAPI server api
type ServerApi struct {
	ApiName string `json:"apiName"`
	ApiDesc string `json:"apiDesc"`
	AppRoot string `json:"appRoot"`
}

//ServerNode server node
type ServerNode struct {
	Name     string       `json:"name"`
	Version  string       `json:"version"`
	Desc     string       `json:"desc"`
	Address  string       `json:"address"`
	Channels []string     `json:"channels"`
	ApiList  []*ServerApi `json:"apiList"`
}

//TypeMapFrom value
func TypeMapFrom(v interface{}) map[string]reflect.Type {
	return TypeMapOf(reflect.TypeOf(v))
}

//TypeMapOf type
func TypeMapOf(typ reflect.Type) map[string]reflect.Type {
	typMap := make(map[string]reflect.Type)
	fetchTypeMap(typ, typMap)
	return typMap
}

func fetchTypeMap(typ reflect.Type, typMap map[string]reflect.Type) {
	typMap[typ.Name()] = typ
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		switch f.Type.Kind() {
		case reflect.Struct:
			fetchTypeMap(f.Type, typMap)
		case reflect.Array:
		case reflect.Slice:
			if f.Type.Elem().Kind() == reflect.Struct {
				fetchTypeMap(f.Type.Elem(), typMap)
			}
		}
	}
}
