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
	"io"
	"reflect"
	"strings"
)

//Encoder type
type Encoder struct {
	writer     io.Writer
	clsDefList []ClassDef
	nameMap    map[string]string
}

//NewEncoder new
func NewEncoder(w io.Writer, np map[string]string) *Encoder {
	if w == nil {
		return nil
	}
	if np == nil {
		np = make(map[string]string, 17)
	}
	encoder := &Encoder{w, make([]ClassDef, 0, 17), np}
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

//WriteObject write object
func (e *Encoder) WriteObject(data interface{}) (int, error) {
	if data == nil {
		io.WriteString(e.writer, "N")
		return 1, nil
	}
	typ := reflect.TypeOf(data)
	for typ.Kind() == reflect.Ptr {
		data = reflect.ValueOf(data).Elem().Interface()
		typ = typ.Elem()
	}

	switch typ.Kind() {
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
	case reflect.Slice, reflect.Array:
		return e.writeList(data)
	case reflect.Float32:
		value := data.(float32)
		return e.writeDouble(float64(value))
	case reflect.Float64:
		value := data.(float64)
		return e.writeDouble(value)
	case reflect.Map:
		return e.writeMap(data)
	case reflect.Bool:
		value := data.(bool)
		return e.writeBoolean(value)
	case reflect.Struct:
		return e.writeInstance(data)
	}
	return 0, fmt.Errorf("unsupported object:%v, kind:%v, type:%v", data, typ.Kind(), typ)
}

// see: http://hessian.caucho.com/doc/hessian-serialization.html##map
func (e *Encoder) writeMap(data interface{}) (int, error) {
	vv := reflect.ValueOf(data)
	typ := vv.Type()

	mapName, ok := e.nameMap[typ.Name()]
	if ok {
		e.writeBT(BcMap)
		e.writeString(mapName)
	} else {
		e.writeBT(BcMapUntyped)
	}

	count := 0

	if typ.Kind() == reflect.Map {
		// -------> untyped map
		keys := vv.MapKeys()
		count = len(keys)
		for i := 0; i < count; i++ {
			k := keys[i]
			_, err := e.WriteObject(k.Interface())
			if err != nil {
				return 0, err
			}
			_, err = e.WriteObject(vv.MapIndex(keys[i]).Interface())
			if err != nil {
				return 0, err
			}
		}
	} else {
		// -------> typed map
		count = vv.NumField()
		for i := 0; i < count; i++ {
			f := vv.Field(i)
			e.writeString(f.Type().Name())
			_, err := e.WriteObject(f.Interface())
			if err != nil {
				return 0, err
			}
		}
	}

	e.writeBT(BcEnd)

	return count, nil
}

// see: http://hessian.caucho.com/doc/hessian-serialization.html##list
func (e *Encoder) writeList(data interface{}) (int, error) {
	vv := reflect.ValueOf(data)
	e.writeBT(BcListFixedUntyped)
	e.writeInt(int32(vv.Len()))
	for i := 0; i < vv.Len(); i++ {
		e.WriteObject(vv.Index(i).Interface())
	}
	return vv.Len(), nil
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

func (e *Encoder) writeBytes(value []byte) (int, error) {
	return e.writer.Write(encodeBinary(value))
}

func (e *Encoder) writeBT(bs ...byte) (int, error) {
	return e.writer.Write(bs)
}

//see: http://hessian.caucho.com/doc/hessian-serialization.html##object
func (e *Encoder) writeInstance(data interface{}) (int, error) {
	typ := reflect.TypeOf(data)
	vv := reflect.ValueOf(data)

	clsName, ok := e.nameMap[typ.Name()]
	if !ok {
		clsName = typ.Name()
		e.nameMap[clsName] = clsName
	}
	l, ok := e.existClassDef(clsName)
	if !ok {
		l, _ = e.writeClsDef(typ, clsName)
	}
	if byte(l) <= ObjectDirectMax {
		e.writeBT(byte(l) + BcObjectDirect)
	} else {
		e.writeBT(BcObject)
		e.writeInt(int32(l))
	}
	for i := 0; i < vv.NumField(); i++ {
		_, err := e.WriteObject(vv.Field(i).Interface())
		if err != nil {
			return 0, err
		}
	}
	return vv.NumField(), nil
}

func (e *Encoder) writeClsDef(typ reflect.Type, clsName string) (int, error) {
	e.writeBT(BcObjectDef)
	e.writeString(clsName)
	fldList := make([]string, typ.NumField())
	e.writeInt(int32(len(fldList)))
	for i := 0; i < len(fldList); i++ {
		str, _ := lowerName(typ.Field(i).Name)
		fldList[i] = str
		e.writeString(fldList[i])
	}
	clsDef := ClassDef{clsName, fldList}
	l := len(e.clsDefList)
	e.clsDefList = append(e.clsDefList, clsDef)
	return l, nil
}

func (e *Encoder) existClassDef(clsName string) (int, bool) {
	for i := 0; i < len(e.clsDefList); i++ {
		if strings.Compare(clsName, e.clsDefList[i].FullClassName) == 0 {
			return i, true
		}
	}
	return 0, false
}
