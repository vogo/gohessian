// Copyright 2018 vogo.
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
	"errors"
	"fmt"
	"reflect"
	"strings"
)

//CodecNamable to define codec name for hessian
type CodecNamable interface {
	HessianCodecName() string
}

//ValueExtractor extract info from struct value
type ValueExtractor func(v reflect.Value)

//TypeMapFrom instance
func TypeMapFrom(v interface{}) map[string]reflect.Type {
	return ExtractTypeMap(reflect.ValueOf(v))
}

//ExtractTypeMap from reflect value
func ExtractTypeMap(value reflect.Value) map[string]reflect.Type {
	typMap := make(map[string]reflect.Type)
	ExtractValue(value, func(v reflect.Value) {
		typ := v.Type()
		typMap[typ.Name()] = typ

		if n, ok := v.Interface().(CodecNamable); ok {
			typMap[n.HessianCodecName()] = typ
		}

	})
	return typMap
}

//NameMapFrom instance
func NameMapFrom(v interface{}) map[string]string {
	return ExtractNameMap(reflect.ValueOf(v))
}

//ExtractNameMap from reflect value
func ExtractNameMap(value reflect.Value) map[string]string {
	nameMap := make(map[string]string)
	ExtractValue(value, func(v reflect.Value) {
		typ := v.Type()
		if n, ok := v.Interface().(CodecNamable); ok {
			nameMap[typ.Name()] = n.HessianCodecName()
		}
	})
	return nameMap
}

//ExtractValue info
func ExtractValue(v reflect.Value, extractor ValueExtractor) {
	v = OriginalValue(v)

	if IsRawKind(v.Kind()) {
		return
	}

	extractor(v)

	if v.Kind() == reflect.Array || v.Kind() == reflect.Slice {
		itemTyp := OriginalType(v.Type().Elem())
		if IsRawKind(itemTyp.Kind()) {
			return
		}

		if v.Len() == 0 {
			ExtractValue(reflect.New(itemTyp), extractor)
			return
		}

		for i := 0; i < v.Len(); i++ {
			ExtractValue(v.Index(i), extractor)
		}
		return
	}

	if v.Kind() == reflect.Map {
		if v.Len() == 0 {
			keyTyp := OriginalType(v.Type().Key())
			valueTyp := OriginalType(v.Type().Elem())
			if !IsRawKind(keyTyp.Kind()) {
				ExtractValue(reflect.New(keyTyp), extractor)
			}
			if !IsRawKind(valueTyp.Kind()) {
				ExtractValue(reflect.New(valueTyp), extractor)
			}
			return
		}

		for _, keyValue := range v.MapKeys() {
			ExtractValue(keyValue, extractor)
			ExtractValue(v.MapIndex(keyValue), extractor)
		}
		return
	}

	if v.Kind() == reflect.Struct {
		for i := 0; i < v.NumField(); i++ {
			f := v.Field(i)
			ExtractValue(f, extractor)
		}
	}
}

//TypeMapOf type
func TypeMapOf(typ reflect.Type) map[string]reflect.Type {
	typMap := make(map[string]reflect.Type)
	FetchType(typ, typMap)
	return typMap
}

//FetchType map
func FetchType(typ reflect.Type, typMap map[string]reflect.Type) {
	typ = OriginalType(typ)

	if IsRawKind(typ.Kind()) {
		return
	}

	if typ.Kind() == reflect.Array || typ.Kind() == reflect.Slice {
		FetchType(typ.Elem(), typMap)
		return
	}

	if typ.Kind() == reflect.Map {
		FetchType(typ.Key(), typMap)
		FetchType(typ.Elem(), typMap)
		return
	}

	if typ.Kind() != reflect.Struct {
		return
	}

	typMap[typ.Name()] = typ
	for i := 0; i < typ.NumField(); i++ {
		FetchType(typ.Field(i).Type, typMap)
	}

}

//OriginalType unpack pointer type to original type
func OriginalType(typ reflect.Type) reflect.Type {
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	return typ
}

//OriginalValue unpack pointer value to original value
func OriginalValue(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v
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

//IsRawKind check whether k is raw kind
func IsRawKind(k reflect.Kind) bool {
	switch k {
	case reflect.Struct, reflect.Interface, reflect.Map, reflect.Array, reflect.Slice, reflect.Ptr:
		return false
	default:
		return true
	}
}

//IntKind check whether k is int kind
func IntKind(k reflect.Kind) bool {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	default:
		return false
	}
}

//UintKind check whether k is uint kind
func UintKind(k reflect.Kind) bool {
	switch k {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	default:
		return false
	}
}

//FloatKind check whether k is float kind
func FloatKind(k reflect.Kind) bool {
	switch k {
	case reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

//EnsureFloat64 convert i to float64
func EnsureFloat64(i interface{}) float64 {
	if i64, ok := i.(float64); ok {
		return i64
	}
	if i32, ok := i.(float32); ok {
		return float64(i32)
	}
	panic(fmt.Errorf("can't convert to float64: %v, type:%v", i, reflect.TypeOf(i)))
}

//EnsureInt64 convert i to int64
func EnsureInt64(i interface{}) int64 {
	if i64, ok := i.(int64); ok {
		return i64
	}
	if i32, ok := i.(int32); ok {
		return int64(i32)
	}
	panic(fmt.Errorf("can't convert to int64: %v, type:%v", i, reflect.TypeOf(i)))
}

//EnsureUint64 convert i to uint64
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

//SetSlice set value into slice object
func SetSlice(value reflect.Value, objects interface{}) error {
	if objects == nil {
		return nil
	}

	v := reflect.ValueOf(objects)
	k := v.Type().Kind()
	if k != reflect.Slice && k != reflect.Array {
		return fmt.Errorf("expect slice type, but get %v, value: %v", k, objects)
	}
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

		switch {
		case elemFloatType:
			sl.Index(i).SetFloat(EnsureFloat64(itemValue.Interface()))
		case elemIntType:
			sl.Index(i).SetInt(EnsureInt64(itemValue.Interface()))
		case elemUintType:
			sl.Index(i).SetUint(EnsureUint64(itemValue.Interface()))
		default:
			sl.Index(i).Set(itemValue)
		}
	}

	value.Set(sl)
	return nil
}

func findField(name string, typ reflect.Type) (int, error) {
	for i := 0; i < typ.NumField(); i++ {
		str := typ.Field(i).Name
		if strings.Compare(str, name) == 0 {
			return i, nil
		}
		str1 := capitalizeName(name)
		if strings.Compare(str, str1) == 0 {
			return i, nil
		}
	}
	return 0, errors.New("no field " + name)
}
