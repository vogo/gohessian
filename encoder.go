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
	"io"
	"reflect"
)

//Encoder type
type Encoder struct {
	writer     io.Writer
	clsDefList []ClassDef
	nameMap    map[string]string
	refMap     map[uintptr]int
}

//NewEncoder new
func NewEncoder(w io.Writer, np map[string]string) *Encoder {
	if w == nil {
		return nil
	}
	if np == nil {
		np = make(map[string]string, 17)
	}
	encoder := &Encoder{
		writer:     w,
		clsDefList: make([]ClassDef, 0, 17),
		nameMap:    np,
		refMap:     make(map[uintptr]int, 17),
	}
	return encoder
}

//RegisterNameType register name type
func (e *Encoder) RegisterNameType(key string, javaClsName string) {
	e.nameMap[key] = javaClsName
}

//RegisterNameMap register name map
func (e *Encoder) RegisterNameMap(mp map[string]string) {
	e.nameMap = mp
}

//Reset reset
func (e *Encoder) Reset() {
	e.nameMap = make(map[string]string, 17)
	e.clsDefList = make([]ClassDef, 0, 17)
}

//WriteData write object
func (e *Encoder) WriteData(data interface{}) (int, error) {
	if data == nil {
		e.writeBT(_nilTag)
		return 1, nil
	}
	source := data
	v := UnpackPtrValue(reflect.ValueOf(data))

	if !IsRawKind(v.Kind()) && IsZero(v) {
		e.writeBT(_nilTag)
		return 1, nil
	}

	data = v.Interface()

	switch v.Kind() {
	case reflect.Bool:
		value := data.(bool)
		return e.writeBoolean(value)
	case reflect.String:
		value := data.(string)
		return e.writeString(value)
	case reflect.Int8: // as int
		value := int32(data.(int8))
		return e.writeInt(value)
	case reflect.Int16: // as int
		value := int32(data.(int16))
		return e.writeInt(value)
	case reflect.Int32: // as int
		value := data.(int32)
		return e.writeInt(value)
	case reflect.Int: // as int
		value := int32(data.(int))
		return e.writeInt(value)
	case reflect.Uint8: // as int
		value := int32(data.(uint8))
		return e.writeInt(value)
	case reflect.Uint16: // as int
		value := int32(data.(uint16))
		return e.writeInt(value)
	case reflect.Int64: // as long
		value := data.(int64)
		return e.writeLong(value)
	case reflect.Uint: // as long
		value := int64(data.(uint))
		return e.writeLong(value)
	case reflect.Uint32: // as long
		value := int64(data.(uint32))
		return e.writeLong(value)
	case reflect.Uint64: // as long
		value := int64(data.(uint64))
		return e.writeLong(value)
	case reflect.Float32:
		value := data.(float32)
		return e.writeDouble(float64(value))
	case reflect.Float64:
		value := data.(float64)
		return e.writeDouble(value)
	case reflect.Slice, reflect.Array:
		return e.writeList(source)
	case reflect.Map:
		return e.writeMap(source)
	case reflect.Struct:
		return e.writeObject(source)
	}
	return 0, newCodecError("WriteData", "unsupported object:%v, kind:%v, type:%v", data, v.Kind(), v.Kind())
}

func (e *Encoder) writeString(value string) (int, error) {
	return e.writer.Write(encodeString(value))
}

func (e *Encoder) writeInt(value int32) (int, error) {
	return e.writer.Write(encodeInt(value))
}

func (e *Encoder) writeLong(value int64) (int, error) {
	return e.writer.Write(encodeLong(value))
}

func (e *Encoder) writeDouble(value float64) (int, error) {
	bytes, err := encodeDouble(value)
	if err != nil {
		return 0, err
	}
	return e.writer.Write(bytes)
}

func (e *Encoder) writeBoolean(value bool) (int, error) {
	return e.writer.Write(encodeBoolean(value))
}

func (e *Encoder) writeBinary(value []byte) (int, error) {
	return e.writer.Write(encodeBinary(value))
}

func (e *Encoder) writeBT(bs ...byte) (int, error) {
	return e.writer.Write(bs)
}
