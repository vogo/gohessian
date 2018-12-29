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

func lowerName(name string) (string, error) {
	if name[0] >= 'a' && name[0] <= 'c' {
		return name, nil
	}
	if name[0] >= 'A' && name[0] <= 'Z' {
		bs := make([]byte, len(name))
		bs[0] = byte(name[0] + AsciiGap)
		copy(bs[1:], name[1:])
		return string(bs), nil
	}
	return name, nil
}

func strTag(tag byte) bool {
	return (tag >= BcStringDirect && tag <= StringDirectMax) || (tag >= 0x30 && tag <= 0x34) || (tag == BcString || tag == BcStringChunk)
}

func isBuildInType(typeStr string) bool {
	switch typeStr {
	case ArrayString:
		return true
	case ArrayInt:
		return true
	case ArrayFloat:
		return true
	case ArrayDouble:
		return true
	case ArrayBool:
		return true
	case ArrayLong:
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
