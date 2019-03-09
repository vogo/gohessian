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

var (
	_zeroBoolPinter *bool
	_zeroValue      = reflect.ValueOf(_zeroBoolPinter).Elem()
)

//CodecNamable to define codec name for hessian
type CodecNamable interface {
	HessianCodecName() string
}

//ValueExtractor extract info from struct value
// return true to continue extracting process
// return false to return current extracting process
type ValueExtractor func(v reflect.Value) bool

//TypeMapFrom instance
func TypeMapFrom(v interface{}) map[string]reflect.Type {
	typMap, _ := ExtractTypeNameMap(v)
	return typMap
}

//NameMapFrom instance
func NameMapFrom(v interface{}) map[string]string {
	_, nameMap := ExtractTypeNameMap(v)
	return nameMap
}

//ExtractTypeNameMap from reflect value
func ExtractTypeNameMap(v interface{}) (map[string]reflect.Type, map[string]string) {
	value := reflect.ValueOf(v)
	typMap := make(map[string]reflect.Type)
	nameMap := make(map[string]string)
	ExtractValue(value, func(v reflect.Value) bool {
		if !v.IsValid() {
			return false
		}
		typ := v.Type()
		name := TypeName(typ)
		if _, ok := typMap[name]; ok {
			return false
		}

		typMap[name] = typ
		nameMap[name] = name

		if v.CanInterface() {
			if n, ok := v.Interface().(CodecNamable); ok {
				nameMap[name] = n.HessianCodecName()
				typMap[n.HessianCodecName()] = typ
			}
		}
		return true
	})

	for k, v := range nameMap {
		if v[0] == '[' {
			v = formatArrayTypeName(v)
			elemName := arrayRootElemName(v)
			replaceName, ok := nameMap[elemName]
			if !ok {
				replaceName, ok = _buildInTypeNameMap[elemName]
			}
			if ok {
				v = strings.Replace(v, elemName, replaceName, -1)
			}

			nameMap[k] = v
			nameMap[v] = v

			typ, _ := typMap[k]
			typMap[v] = typ
		}
	}

	return typMap, nameMap
}

// remove pointer '*' and right bracket ']'
func formatArrayTypeName(v string) string {
	v = strings.Replace(v, "]", "", -1)
	v = strings.Replace(v, "*", "", -1)
	return v
}

//ExtractValue info
func ExtractValue(v reflect.Value, extractor ValueExtractor) {
	v = RawValue(v)

	if !extractor(v) {
		return
	}

	if v.Kind() == reflect.Array || v.Kind() == reflect.Slice {
		itemTyp := UnpackPtrType(v.Type().Elem())

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
			keyTyp := UnpackPtrType(v.Type().Key())
			valueTyp := UnpackPtrType(v.Type().Elem())
			ExtractValue(reflect.New(keyTyp), extractor)
			ExtractValue(reflect.New(valueTyp), extractor)
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
			ExtractValue(v.Field(i), extractor)
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
	typ = UnpackPtrType(typ)

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

//TypeName return the name of type
func TypeName(t reflect.Type) string {
	n := t.Name()
	if n == "" {
		n = t.String()
	}
	return n
}

// RawValue unpack value to raw value.
// NOTE: it may be the zero value
// return value of pointer pointed if it's pointer
func RawValue(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v
}

//UnpackPtrType unpack pointer type to original type
func UnpackPtrType(typ reflect.Type) reflect.Type {
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	return typ
}

// UnpackPtrValue unpack pointer value to original value
// return the pointer if its elem is zero value, because lots of operations on zero value is invalid
func UnpackPtrValue(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr && v.Elem().IsValid() {
		v = v.Elem()
	}
	return v
}

// UnpackPtr unpack pointer value to original value
func UnpackPtr(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v
}

//PackPtr pack a Ptr value
func PackPtr(v reflect.Value) reflect.Value {
	vv := reflect.New(v.Type())
	vv.Elem().Set(v)
	return vv
}

//EnsureRawValue pack the interface with value, and make sure it's not a ref holder
func EnsureRawValue(in interface{}) reflect.Value {
	if v, ok := in.(reflect.Value); ok {
		if v.IsValid() {
			if r, ok := v.Interface().(*_refHolder); ok {
				return r.value
			}
		}
		return v
	}
	if v, ok := in.(*_refHolder); ok {
		in = v.value
	}
	return reflect.ValueOf(in)
}

//EnsurePackValue pack the interface with value
func EnsurePackValue(in interface{}) reflect.Value {
	if v, ok := in.(reflect.Value); ok {
		return v
	}
	return reflect.ValueOf(in)
}

// EnsureInterface get value of reflect.Value
// return original value if not reflect.Value
func EnsureInterface(in interface{}, err error) (interface{}, error) {
	if err != nil {
		return in, err
	}
	if v, ok := in.(reflect.Value); ok {
		in = v.Interface()
	}
	if v, ok := in.(*_refHolder); ok {
		in = v.value.Interface()
	}
	if v, ok := in.(reflect.Value); ok {
		in = v.Interface()
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

//ElemKind check whether k is elem kind, a value of which kind can call func Elem()
func ElemKind(k reflect.Kind) bool {
	switch k {
	case reflect.Array, reflect.Ptr, reflect.Interface:
		return true
	default:
		return false
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
func SetSlice(dest reflect.Value, objects interface{}) error {
	if objects == nil {
		return nil
	}

	dest = UnpackPtrValue(dest)
	destTyp := UnpackPtrType(dest.Type())
	elemKind := destTyp.Elem().Kind()
	if elemKind == reflect.Uint8 {
		// for binary
		dest.Set(EnsureRawValue(objects))
		return nil
	}

	if ref, ok := objects.(*_refHolder); ok {
		v, err := ConvertSliceValueType(destTyp, ref.value)
		if err != nil {
			return err
		}
		SetValue(dest, v)
		ref.change(v) // change finally
		ref.notify()
		return nil
	}

	v := EnsurePackValue(objects)
	if h, ok := v.Interface().(*_refHolder); ok {
		h.add(dest)
		return nil
	}

	v, err := ConvertSliceValueType(destTyp, v)
	if err != nil {
		return err
	}
	SetValue(dest, v)
	return nil
}

func ConvertSliceValueType(destTyp reflect.Type, v reflect.Value) (reflect.Value, error) {
	if destTyp == v.Type() {
		return v, nil
	}

	k := v.Type().Kind()
	if k != reflect.Slice && k != reflect.Array {
		return _zeroValue, newCodecError("ConvertSliceValueType", "expect slice type, but get %v, objects: %v", k, v)
	}

	if v.Len() <= 0 {
		return _zeroValue, nil
	}

	elemKind := destTyp.Elem().Kind()
	elemPtrType := elemKind == reflect.Ptr
	elemFloatType := FloatKind(elemKind)
	elemIntType := IntKind(elemKind)
	elemUintType := UintKind(elemKind)

	sl := reflect.MakeSlice(destTyp, v.Len(), v.Len())
	var itemValue reflect.Value
	for i := 0; i < v.Len(); i++ {
		item := v.Index(i).Interface()
		if cv, ok := item.(reflect.Value); ok {
			itemValue = cv
		} else {
			itemValue = reflect.ValueOf(item)
		}

		if !elemPtrType && itemValue.Kind() == reflect.Ptr {
			itemValue = UnpackPtrValue(itemValue)
		}

		switch {
		case elemFloatType:
			sl.Index(i).SetFloat(EnsureFloat64(itemValue.Interface()))
		case elemIntType:
			sl.Index(i).SetInt(EnsureInt64(itemValue.Interface()))
		case elemUintType:
			sl.Index(i).SetUint(EnsureUint64(itemValue.Interface()))
		default:
			SetValue(sl.Index(i), itemValue)
		}
	}

	return sl, nil
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

// SetValue set the value to dest.
// It will auto check the Ptr pack level and unpack/pack to the right level.
// It make sure success to set value
func SetValue(dest, v reflect.Value) {
	// check whether the v is a ref holder
	if v.IsValid() {
		if h, ok := v.Interface().(*_refHolder); ok {
			h.add(dest)
			return
		}
	}

	// if the kind of dest is Ptr, the original value will be zero value
	// set value on zero value is not allowed
	// unpack to one-level pointer
	for dest.Kind() == reflect.Ptr && dest.Elem().Kind() == reflect.Ptr {
		dest = dest.Elem()
	}

	// if the kind of dest is Ptr, change the v to a Ptr value too.
	if dest.Kind() == reflect.Ptr {

		// unpack to one-level pointer
		for v.IsValid() && v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Ptr {
			v = v.Elem()
		}

		// zero value not need to set
		if !v.IsValid() {
			return
		}

		if v.Kind() != reflect.Ptr {
			// change the v to a Ptr value
			v = PackPtr(v)
		}
	} else {
		v = UnpackPtrValue(v)
	}

	// zero value not need to set
	if !v.IsValid() {
		return
	}

	// set value as required type
	if dest.Type() == v.Type() {
		dest.Set(v)
		return
	}

	// unpack ptr so that to special check for float,int,uint kind
	if dest.Kind() == reflect.Ptr {
		dest = UnpackPtrValue(dest)
		v = UnpackPtrValue(v)
	}

	kind := dest.Kind()
	switch kind {
	case reflect.Float32, reflect.Float64:
		dest.SetFloat(EnsureFloat64(v.Interface()))
		return
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		dest.SetInt(EnsureInt64(v.Interface()))
		return
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		dest.SetUint(EnsureUint64(v.Interface()))
		return
	}

	dest.Set(v)
}

func AddrEqual(x, y interface{}) bool {
	if x == nil || y == nil {
		return x == y
	}
	v1 := reflect.ValueOf(x)
	v2 := reflect.ValueOf(y)
	if v1.Type() != v2.Type() {
		return false
	}

	if v1.Kind() != reflect.Ptr {
		v1 = PackPtr(v1)
		v2 = PackPtr(v2)
	}
	return v1.Pointer() == v2.Pointer()
}

// arrayRootElemName format may be '[[[string' or '[][][]string'
// this func will remove the bracket prefix
func arrayRootElemName(arrayName string) string {
	idx := strings.LastIndex(arrayName, "]")
	if idx < 0 {
		idx = strings.LastIndex(arrayName, "[")
	}
	if idx >= 0 {
		arrayName = arrayName[idx+1:]
	}
	idx = strings.LastIndex(arrayName, "*")
	if idx >= 0 {
		arrayName = arrayName[idx+1:]
	}
	return arrayName
}
