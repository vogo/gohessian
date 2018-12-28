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
	"math"
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

// see: http://hessian.caucho.com/doc/hessian-serialization.html##double
func (e *Encoder) writeDouble(value float64) (int, error) {
	v := float64(int64(value))
	if v == value {
		iv := int64(value)
		switch iv {
		case 0:
			return e.writeBT(BC_DOUBLE_ZERO)
		case 1:
			return e.writeBT(BC_DOUBLE_ONE)
		}
		if iv >= -0x80 && iv < 0x80 {
			e.writeBT(BC_DOUBLE_BYTE, byte(iv))
		} else if iv >= -0x8000 && iv < 0x8000 {
			e.writeBT(BC_DOUBLE_BYTE, byte(iv>>8), byte(iv))
		}
	} else {
		bits := uint64(math.Float64bits(value))
		e.writeBT(BC_DOUBLE)
		e.writeBT(byte(bits>>56), byte(bits>>48), byte(bits>>40), byte(bits>>32), byte(bits>>24),
			byte(bits>>16), byte(bits>>8), byte(bits))
	}
	return 8, nil

}

// see: http://hessian.caucho.com/doc/hessian-serialization.html##map
func (e *Encoder) writeMap(data interface{}) (int, error) {
	vv := reflect.ValueOf(data)
	typ := vv.Type()

	mapName, ok := e.nameMap[typ.Name()]
	if ok {
		e.writeBT(BC_MAP)
		e.writeString(mapName)
	} else {
		e.writeBT(BC_MAP_UNTYPED)
	}

	count := 0

	if typ.Kind() == reflect.Map {
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

	e.writeBT(BC_END)

	return count, nil
}

// see: http://hessian.caucho.com/doc/hessian-serialization.html##string
func (e *Encoder) writeString(value string) (int, error) {
	dataBys := []byte(value)
	l := len(dataBys)
	sub := 0x8000
	begin := 0
	for l > sub {
		buf := make([]byte, 3)
		buf[0] = BC_STRING_CHUNK
		buf[1] = byte(sub >> 8)
		buf[2] = byte(sub)
		_, err := e.writer.Write(buf)
		if err != nil {
			return 0, newCodecError("writeString", err)
		}
		buf = make([]byte, sub)
		copy(buf, dataBys[begin:begin+sub])
		_, err = e.writer.Write(buf)
		if err != nil {
			return 0, newCodecError("writeString", err)
		}
		l -= sub
		begin += sub
	}
	var buf []byte
	if l == 0 {
		e.writer.Write([]byte{BC_NULL})
		return len(dataBys), nil
	} else if l <= int(STRING_DIRECT_MAX) {
		buf = make([]byte, 1)
		buf[0] = byte(l + int(BC_STRING_DIRECT))
	} else if l <= int(STRING_SHORT_MAX) {
		buf = make([]byte, 2)
		buf[0] = byte((l >> 8) + int(BC_STRING_SHORT))
		buf[1] = byte(l)
	} else {
		buf = make([]byte, 3)
		buf[0] = BC_STRING
		buf[0] = byte(l >> 8)
		buf[1] = byte(l)
	}
	bs := make([]byte, l+len(buf))
	copy(bs[0:], buf)
	copy(bs[len(buf):], dataBys[begin:])
	_, err := e.writer.Write(bs)
	if err != nil {
		return 0, newCodecError("writeString", err)
	}
	return l, nil
}

// see: http://hessian.caucho.com/doc/hessian-serialization.html##list
func (e *Encoder) writeList(data interface{}) (int, error) {
	vv := reflect.ValueOf(data)
	e.writeBT(BC_LIST_FIXED_UNTYPED)
	e.writeInt(int32(vv.Len()))
	for i := 0; i < vv.Len(); i++ {
		e.WriteObject(vv.Index(i).Interface())
	}
	return vv.Len(), nil
}

// see: http://hessian.caucho.com/doc/hessian-serialization.html##int
func (e *Encoder) writeInt(value int32) (int, error) {
	var buf []byte
	if Int32DirectMin <= value && value <= Int32DirectMax {
		buf = make([]byte, 1)
		buf[0] = byte(Int32BcIntZero + value)
	} else if Int32ByteMin <= value && value <= Int32ByteMax {
		buf = make([]byte, 2)
		buf[0] = byte(Int32BcIntByteZero + value>>8)
		buf[1] = byte(value)
	} else if Int32ShortMin <= value && value <= Int32ShortMax {
		buf = make([]byte, 3)
		buf[0] = byte(Int32BcIntShortZero + value>>16)
		buf[1] = byte(value >> 8)
		buf[2] = byte(value)
	} else {
		buf = make([]byte, 5)
		buf[0] = byte('I')
		buf[1] = byte(value >> 24)
		buf[2] = byte(value >> 16)
		buf[3] = byte(value >> 8)
		buf[4] = byte(value)
	}
	l, err := e.writer.Write(buf)
	if err != nil {
		return 0, newCodecError("WriteInt", err)
	}
	return l, nil
}

// see: http://hessian.caucho.com/doc/hessian-serialization.html##long
func (e *Encoder) writeLong(value int64) (int, error) {
	var buf []byte
	if Int64LongDirectMin <= value && value <= Int64LongDirectMax {
		buf = make([]byte, 1)
		buf[0] = byte(Int64BcLongZero + value)
	} else if Int64LongByteMin <= value && value <= Int64LongByteMax {
		buf = make([]byte, 2)
		buf[0] = byte(Int64BcLongByteZero + (value >> 8))
		buf[1] = byte(value)
	} else if Int64LongShortMin <= value && value <= Int64LongShortMax {
		buf = make([]byte, 3)
		buf[0] = byte(Int64BcLongShortZero + (value >> 16))
		buf[1] = byte(value >> 8)
		buf[2] = byte(value)
	} else if 0x80000000 <= value && value <= 0x7fffffff {
		buf = make([]byte, 5)
		buf[0] = BC_LONG_INT
		buf[1] = byte(value >> 24)
		buf[2] = byte(value >> 16)
		buf[3] = byte(value >> 8)
		buf[4] = byte(value)
	} else {
		buf = make([]byte, 9)
		buf[0] = 'L'
		buf[1] = byte(value >> 56)
		buf[2] = byte(value >> 48)
		buf[3] = byte(value >> 32)
		buf[4] = byte(value >> 24)
		buf[5] = byte(value >> 16)
		buf[6] = byte(value >> 8)
		buf[7] = byte(value)
	}
	l, err := e.writer.Write(buf)
	if err != nil {
		return 0, newCodecError("WriteLong", err)
	}
	return l, nil
}

// see: http://hessian.caucho.com/doc/hessian-serialization.html##boolean
func (e *Encoder) writeBoolean(value bool) (int, error) {
	buf := make([]byte, 1)
	if value {
		buf[0] = BC_TRUE
	} else {
		buf[0] = BC_FALSE
	}
	l, err := e.writer.Write(buf)
	if err != nil {
		return 0, newCodecError("WriteBoolean", err)
	}
	return l, nil
}

// see: http://hessian.caucho.com/doc/hessian-serialization.html##binary
func (e *Encoder) writeBytes(value []byte) (int, error) {
	sub := CHUNK_SIZE
	l := len(value)
	begin := 0

	for l > sub {
		buf := make([]byte, 3+CHUNK_SIZE)
		buf[0] = byte(BC_BINARY_CHUNK)
		buf[1] = byte(sub >> 8)
		buf[2] = byte(sub)

		copy(buf[3:], value[begin:begin+CHUNK_SIZE])
		_, err := e.writer.Write(buf)
		if err != nil {
			return 0, newCodecError("WriteBytes", err)
		}
		l -= CHUNK_SIZE
		begin += CHUNK_SIZE
	}
	var buf []byte
	if l == 0 {
		return len(value), nil
	} else if l <= int(BINARY_DIRECT_MAX) {
		buf := make([]byte, 1)
		buf[0] = byte(int(BC_BINARY_DIRECT) + l)
	} else if l <= int(BINARY_SHORT_MAX) {
		buf := make([]byte, 2)
		buf[0] = byte(int(BC_BINARY_SHORT) + l>>8)
		buf[1] = byte(l)
	} else {
		buf := make([]byte, 3)
		buf[0] = byte(BC_BINARY)
		buf[1] = byte(l >> 8)
		buf[2] = byte(l)
	}
	bs := make([]byte, l+len(buf))
	copy(bs[0:], buf)
	copy(bs[len(buf):], value[begin:])
	_, err := e.writer.Write(bs)
	if err != nil {
		return 0, newCodecError("WriteBytes", err)
	}
	return len(value), nil
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
	if byte(l) <= OBJECT_DIRECT_MAX {
		e.writeBT(byte(l) + BC_OBJECT_DIRECT)
	} else {
		e.writeBT(BC_OBJECT)
		e.writeInt(int32(l))
	}
	for i := 0; i < vv.NumField(); i++ {
		//err := e.writeField(vv.Field(i), vv.Field(i).Type().Kind())
		_, err := e.WriteObject(vv.Field(i).Interface())
		if err != nil {
			return 0, err
		}
	}
	return vv.NumField(), nil
}

func (e *Encoder) writeField(field reflect.Value, kind reflect.Kind) error {
	switch kind {
	case reflect.Struct:
		e.WriteObject(field.Interface())
	case reflect.String:
		v := field.String()
		e.writeString(v)
	case reflect.Int32:
		v := int32(field.Int())
		e.writeInt(v)
	case reflect.Int64:
		v := int64(field.Int())
		e.writeLong(v)
	case reflect.Bool:
		v := field.Bool()
		e.writeBoolean(v)
	case reflect.Map:
		v := field.Interface()
		e.writeMap(v)
	case reflect.Array, reflect.Slice:
		v := field.Interface()
		e.WriteObject(v)
	case reflect.Int:
		v := int32(field.Int())
		e.writeInt(v)
	}
	return newCodecError("writeField unsupported kind " + kind.String())
}

func (e *Encoder) writeClsDef(typ reflect.Type, clsName string) (int, error) {
	e.writeBT(BC_OBJECT_DEF)
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

func lowerName(name string) (string, error) {
	if name[0] >= 'a' && name[0] <= 'c' {
		return name, nil
	}
	if name[0] >= 'A' && name[0] <= 'Z' {
		bs := make([]byte, len(name))
		bs[0] = byte(name[0] + ASCII_GAP)
		copy(bs[1:], name[1:])
		return string(bs), nil
	}
	return name, nil
}

func (e *Encoder) existClassDef(clsName string) (int, bool) {
	for i := 0; i < len(e.clsDefList); i++ {
		if strings.Compare(clsName, e.clsDefList[i].FullClassName) == 0 {
			return i, true
		}
	}
	return 0, false
}
