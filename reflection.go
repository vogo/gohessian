// Copyright 2018 luckincoffee.
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
	"fmt"
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

//UnpackValue unpack reflect.Value
func UnpackValue(in interface{}, err error) (interface{}, error) {
	if err != nil {
		return in, err
	}
	if v, ok := in.(reflect.Value); ok {
		return v.Interface(), nil
	}
	return in, nil
}

func IsRawKind(k reflect.Kind) bool {
	switch k {
	case reflect.Struct, reflect.Interface, reflect.Map, reflect.Array, reflect.Slice, reflect.Ptr:
		return false
	default:
		return true
	}
}

func IntKind(k reflect.Kind) bool {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	default:
		return false
	}
}

func UintKind(k reflect.Kind) bool {
	switch k {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	default:
		return false
	}
}

func FloatKind(k reflect.Kind) bool {
	switch k {
	case reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

func EnsureFloat64(i interface{}) float64 {
	if i64, ok := i.(float64); ok {
		return i64
	}
	if i32, ok := i.(float32); ok {
		return float64(i32)
	}
	panic(fmt.Errorf("can't convert to float64: %v, type:%v", i, reflect.TypeOf(i)))
}

func EnsureInt64(i interface{}) int64 {
	if i64, ok := i.(int64); ok {
		return i64
	}
	if i32, ok := i.(int32); ok {
		return int64(i32)
	}
	panic(fmt.Errorf("can't convert to int64: %v, type:%v", i, reflect.TypeOf(i)))
}

func EnsureUint64(i interface{}) uint64 {
	if i64, ok := i.(uint64); ok {
		return i64
	}
	if i64, ok := i.(int64); ok {
		return uint64(i64)
	}
	if i32, ok := i.(int32); ok {
		return uint64(i32)
	}
	if i32, ok := i.(uint32); ok {
		return uint64(i32)
	}
	panic(fmt.Errorf("can't convert to uint64: %v, type:%v", i, reflect.TypeOf(i)))
}


func SetSlice(value reflect.Value, objects interface{}) error {
	v := reflect.ValueOf(objects)
	elemKind := value.Type().Elem().Kind()
	if objects == nil && v.Len() <= 0 {
		return nil
	}
	if elemKind == reflect.Uint8 {
		// for binary
		value.Set(v)
		return nil
	}
	elemPtrType := elemKind == reflect.Ptr
	elemFloatType := FloatKind(elemKind)
	elemIntType := IntKind(elemKind)
	elemUintType := UintKind(elemKind)

	sl := reflect.MakeSlice(value.Type(), v.Len(), v.Len())
	for i := 0; i < v.Len(); i++ {
		item := v.Index(i).Interface()
		itemValue := reflect.ValueOf(item)
		if cv, ok := itemValue.Interface().(reflect.Value); ok {
			itemValue = cv
		}
		if !elemPtrType && itemValue.Kind() == reflect.Ptr {
			itemValue = itemValue.Elem()
		}

		if elemFloatType {
			sl.Index(i).SetFloat(EnsureFloat64(itemValue.Interface()))
		} else if elemIntType {
			sl.Index(i).SetInt(EnsureInt64(itemValue.Interface()))
		} else if elemUintType {
			sl.Index(i).SetUint(EnsureUint64(itemValue.Interface()))
		} else {
			sl.Index(i).Set(itemValue)
		}
	}

	value.Set(sl)
	return nil
}
