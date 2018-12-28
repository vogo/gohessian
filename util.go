// Copyright 2018 luckin coffee.
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

import "reflect"

func strTag(tag byte) bool {
	return (tag >= BC_STRING_DIRECT && tag <= STRING_DIRECT_MAX) || (tag >= 0x30 && tag <= 0x34) || (tag == BC_STRING || tag == BC_STRING_CHUNK)
}

func isBuildInType(typeStr string) bool {
	switch typeStr {
	case ARRAY_STRING:
		return true
	case ARRAY_INT:
		return true
	case ARRAY_FLOAT:
		return true
	case ARRAY_DOUBLE:
		return true
	case ARRAY_BOOL:
		return true
	case ARRAY_LONG:
		return true
	default:
		return false
	}
}

func buildKey(key reflect.Value, typ reflect.Type) interface{} {
	switch typ.Kind() {
	case reflect.String:
		return key.String()
	case reflect.Bool:
		return key.Bool()
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Uint16:
		return int32(key.Int())
	case reflect.Int64:
		return key.Int()
	case reflect.Uint8:
		return byte(key.Uint())
	case reflect.Uint, reflect.Uint32, reflect.Uint64:
		return int64(key.Uint())
	default:
		return key.Interface()
	}
	return newCodecError("unsupported key kind " + typ.Kind().String())
}
