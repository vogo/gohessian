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

// Decoder implement hessian 2 protocol, It follows java hessian package standard.
// It assume that you using the java name convention
// baisca difference between java and go
// fully qualify java class name is composed of package + class name
// Go assume upper case of field name is exportable and java did not have that constrain
// but in general java using camo camlecase. So it did conversion of field name from
// the first letter of from upper to lower case
// typMap{string]reflect.Type contain full java package+class name and go relfect.Type
// must provide in order to correctly decode to galang interface
//

package hessian

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"reflect"
	"strings"
)

var _ = bytes.MinRead
var _ = reflect.Value{}

// ClassDef class def
type ClassDef struct {
	FullClassName string
	FieldName     []string
}

//Decoder type
type Decoder struct {
	reader     io.Reader
	typMap     map[string]reflect.Type
	typList    []string
	refList    []interface{}
	clsDefList []ClassDef
}

//NewDecoder new
func NewDecoder(r io.Reader, typ map[string]reflect.Type) *Decoder {
	if typ == nil {
		typ = make(map[string]reflect.Type, 17)
	}
	decode := &Decoder{r, typ, make([]string, 0, 17), make([]interface{}, 0, 17), make([]ClassDef, 0, 17)}
	return decode
}

//RegisterType register key/value type
func (d *Decoder) RegisterType(key string, value reflect.Type) {
	d.typMap[key] = value
}

//RegisterTypeMap register map
func (d *Decoder) RegisterTypeMap(mp map[string]reflect.Type) {
	d.typMap = mp
}

//RegisterVal register from value
func (d *Decoder) RegisterVal(key string, val interface{}) {
	d.typMap[key] = reflect.TypeOf(val)
}

//Reset reset
func (d *Decoder) Reset() {
	d.typMap = make(map[string]reflect.Type, 17)
	d.clsDefList = make([]ClassDef, 0, 17)
	d.refList = make([]interface{}, 17)
}

func (d *Decoder) readBufByte() (byte, error) {
	buf := make([]byte, 1)
	_, err := io.ReadFull(d.reader, buf)
	if err != nil {
		return 0, newCodecError("readBufByte", err)
	}
	return buf[0], nil
}

func (d *Decoder) readBuf(s int) ([]byte, error) {
	buf := make([]byte, s)
	_, err := io.ReadFull(d.reader, buf)
	if err != nil {
		return nil, newCodecError("readBuf", err)
	}
	return buf, nil
}

//ReadObjectWithType name is option, if it is nil, use type.Name()
func (d *Decoder) ReadObjectWithType(typ reflect.Type, name string) (interface{}, error) {
	//register the type if it did exist
	if _, ok := d.typMap[name]; ok {
		hlog.Debug("over write existing type")
	}
	d.typMap[name] = typ
	return d.ReadObject()
}

func (d *Decoder) readBinary(flag int32) ([]byte, error) {
	return decodeBinaryTag(d.reader, flag)
}

func (d *Decoder) readInt(flag int32) (int32, error) {
	return decodeIntTag(d.reader, flag)
}

func (d *Decoder) readLong(flag int32) (int64, error) {
	return decodeLongTag(d.reader, flag)
}

func (d *Decoder) readDouble(flag int32) (interface{}, error) {
	var tag byte
	if flag != TagRead {
		tag = byte(flag)
	} else {
		tag, _ = d.readBufByte()
	}
	switch tag {
	case BcLongInt:
		return d.readInt(TagRead)
	case BcDoubleZero:
		return float64(0), nil
	case BcDoubleOne:
		return float64(1), nil
	case BcDoubleByte:
		bt, _ := d.readBufByte()
		return float64(bt), nil
	case BcDoubleShort:
		bf, _ := d.readBuf(2)
		return float64(int(bf[0])*256 + int(bf[1])), nil
	case BcDoubleMill:
		i, _ := d.readInt(TagRead)
		return float64(i), nil
	case BcDouble:
		buf, _ := d.readBuf(8)
		bits := binary.BigEndian.Uint64(buf)
		datum := math.Float64frombits(bits)
		return datum, nil
	}
	return nil, newCodecError("parse double wrong tag " + string(tag))
}

func (d *Decoder) readString(flag int32) (string, error) {
	var tag byte
	if flag != TagRead {
		tag = byte(flag)
	} else {
		tag, _ = d.readBufByte()
	}
	last := true

	if tag == BcNull || !strTag(tag) {
		// null string will not match
		return "", nil
	}

	if tag == BcStringChunk {
		last = false
	} else {
		last = true
	}
	l, err := d.getStrLen(tag)
	if err != nil {
		return "", newCodecError("getStrLen", err)
	}

	var length int32
	length = l
	data := make([]byte, length)
	for i := 0; ; {
		if int32(i) == length {
			if last {
				break
			}

			buf := make([]byte, 1)
			_, err := io.ReadFull(d.reader, buf)

			if err != nil {
				return "", newCodecError("byte1 integer", err)
			}
			b := buf[0]
			switch {
			case b == BcStringChunk || b == BcString:
				if b == BcStringChunk {
					last = false
				} else {
					last = true
				}
				l, err := d.getStrLen(b)
				if err != nil {
					return "", newCodecError("getStrLen", err)
				}
				length += l
				bs := make([]byte, 0, length)
				copy(bs, data)
				data = bs
			default:
				return "", newCodecError("tag error ", err)
			}
		} else {
			buf := make([]byte, 1)
			_, err := io.ReadFull(d.reader, buf)
			if err != nil {
				return "", newCodecError("byte2 integer", err)
			}
			data[i] = buf[0]
			i++
		}
	}

	return string(data), nil
}

func (d *Decoder) getStrLen(tag byte) (int32, error) {
	switch {
	case tag >= BcStringDirect && tag <= StringDirectMax:
		return int32(tag - 0x00), nil
	case tag >= 0x30 && tag <= 0x34:
		buf := make([]byte, 1)
		_, err := io.ReadFull(d.reader, buf)
		if err != nil {
			return -1, newCodecError("byte4 integer", err)
		}
		len := int32(tag-0x30)<<8 + int32(buf[0])
		return len, nil

	case tag == BcStringChunk || tag == BcString:
		buf := make([]byte, 1)
		_, err := io.ReadFull(d.reader, buf)
		if err != nil {
			return -1, newCodecError("byte5 integer", err)
		}
		len := int32(tag)<<8 + int32(buf[0])
		return len, nil
	default:
		return -1, newCodecError("getStrLen")
	}
}

func (d *Decoder) readInstance(typ reflect.Type, cls ClassDef) (interface{}, error) {
	if typ.Kind() != reflect.Struct {
		return nil, newCodecError("wrong type expect struct but get " + typ.String())
	}
	vv := reflect.New(typ)
	st := reflect.ValueOf(vv.Interface()).Elem()
	for i := 0; i < len(cls.FieldName); i++ {
		fldName := cls.FieldName[i]
		index, err := findField(fldName, typ)
		if err != nil {
			hlog.Debugf("%s is not found, will skip type ->p %v", fldName, typ)
			continue
		}
		fldValue := st.Field(index)
		if !fldValue.CanSet() {
			return nil, newCodecError("CanSet false for " + fldName)
		}
		kind := fldValue.Kind()
		switch kind {
		case reflect.String:
			str, err := d.readString(TagRead)
			if err != nil {
				return nil, newCodecError("ReadString "+fldName, err)
			}
			if str != "" {
				fldValue.SetString(str)
			}
		case reflect.Int32, reflect.Int, reflect.Int16, reflect.Int8:
			i, err := d.readInt(TagRead)
			if err != nil {
				return nil, newCodecError("decode int error "+fldName, err)
			}
			v := int64(i)
			fldValue.SetInt(v)
		case reflect.Uint8, reflect.Uint16:
			i, err := d.readInt(TagRead)
			if err != nil {
				return nil, newCodecError("decode int error "+fldName, err)
			}
			v := uint64(i)
			fldValue.SetUint(v)
		case reflect.Int64:
			i, err := d.readLong(TagRead)
			if err != nil {
				return nil, newCodecError("decode long error "+fldName, err)
			}
			fldValue.SetInt(i)
		case reflect.Uint64, reflect.Uint, reflect.Uint32:
			i, err := d.readLong(TagRead)
			if err != nil {
				return nil, newCodecError("decode long error "+fldName, err)
			}
			fldValue.SetUint(uint64(i))
		case reflect.Bool:
			b, err := d.ReadObject()
			if err != nil {
				return nil, newCodecError("decode bool error "+fldName, err)
			}
			fldValue.SetBool(b.(bool))
		case reflect.Float32, reflect.Float64:
			d, err := d.readDouble(TagRead)
			if err != nil {
				return nil, newCodecError("decode float error "+fldName, err)
			}
			fldValue.SetFloat(d.(float64))
		case reflect.Struct:
			s, err := d.ReadObject()
			if err != nil {

				return nil, newCodecError("decode struct error "+fldName, err)
			}
			fldValue.Set(reflect.Indirect(reflect.ValueOf(s)))
		case reflect.Map:
			d.readMap(fldValue)
		case reflect.Slice, reflect.Array:
			m, err := d.ReadObject()
			if err != nil {
				if err == io.EOF {
					break // ignore nil slice
				}
				return nil, newCodecError("decode error "+fldName, err)
			}
			err = SetSlice(fldValue, m)
			if err != nil {
				return nil, err
			}
		}
	}
	return vv, nil
}

// http://hessian.caucho.com/doc/hessian-serialization.html#anchor27
func (d *Decoder) readMap(value reflect.Value) error {
	tag, _ := d.readBufByte()
	if tag == BcMap {
		d.readString(TagRead)
	} else if tag == BcMapUntyped {
		//do nothing
	} else {
		return newCodecError("wrong header BC_MAP_UNTYPED")
	}
	m := reflect.MakeMap(value.Type())

	//read key and value
	for {
		key, err := d.ReadObject()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return newCodecError("ReadType", err)
			}
		}
		vl, err := d.ReadObject()
		m.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(vl))
	}
	value.Set(m)
	return nil
}

func (d *Decoder) readSlice(value reflect.Value) (interface{}, error) {
	tag, _ := d.readBufByte()
	var i int
	if tag >= BcListDirectUntyped && tag <= 0x7f {
		i = int(tag - BcListDirectUntyped)
	} else {
		ii, err := d.readInt(TagRead)
		if err != nil {
			return nil, newCodecError("ReadType", err)
		}
		i = int(ii)
	}
	ary := reflect.MakeSlice(value.Type(), i, i)
	for j := 0; j < i; j++ {
		it, err := d.ReadObject()
		if err != nil {
			return nil, newCodecError("ReadList", err)
		}
		ary.Index(j).Set(reflect.ValueOf(it))
	}
	d.readBufByte()
	value.Set(ary)
	return ary, nil
}

func capitalizeName(name string) string {
	if name[0] >= 'A' && name[0] <= 'Z' {
		return name
	}
	if name[0] >= 'a' && name[0] <= 'z' {
		bs := make([]byte, len(name))
		bs[0] = byte(name[0] - AsciiGap)
		copy(bs[1:], name[1:])
		return string(bs)
	}
	return name

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
	return 0, newCodecError("findField")
}

func (d *Decoder) readType() (string, error) {
	buf := make([]byte, 1)
	_, err := io.ReadFull(d.reader, buf)
	if err != nil {
		return "", newCodecError("reading tag", err)
	}
	tag := buf[0]
	if strTag(tag) {
		t, err := d.readString(int32(tag))
		if err != nil {
			return "", newCodecError("reading tag", err)
		}
		d.typList = append(d.typList, t)
		return t, nil
	}
	i, err := d.readInt(TagRead)
	if err != nil {
		return "", newCodecError("reading tag", err)
	}
	index := int(i)
	return d.typList[index], nil

}

//ReadTypedMap read typed map
// see: http://hessian.caucho.com/doc/hessian-serialization.html#anchor27
func (d *Decoder) ReadTypedMap() (interface{}, error) {
	typ, err := d.readType()
	if err != nil {
		return nil, newCodecError("ReadType", err)
	}
	mType, ok := d.typMap[typ]
	if !ok {
		return nil, newCodecError("ReadType", "no type map for ", typ)
	}
	var mValue reflect.Value
	if mType.Kind() == reflect.Map {
		mValue = reflect.MakeMap(mType)
	} else {
		mValue = reflect.New(mType)
	}

	for {
		key, err := d.ReadObject()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		value, err := d.ReadObject()
		if err != nil {
			return nil, err
		}
		if mType.Kind() == reflect.Map {
			mValue.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))
		} else {
			fieldName, _ := key.(string)
			fieldValue := mValue.FieldByName(fieldName)
			if fieldValue.IsValid() {
				fieldValue.Set(reflect.ValueOf(value))
			}
		}
	}

	m := mValue.Interface()
	return m, nil
}

//ReadMapUntyped read untyped map
// see: http://hessian.caucho.com/doc/hessian-serialization.html#anchor27
func (d *Decoder) ReadUntypedMap() (interface{}, error) {
	m := make(map[interface{}]interface{})
	//read key and value
	for {
		key, err := d.ReadObject()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err

		}
		value, err := d.ReadObject()
		if err != nil {
			return nil, err
		}
		m[key] = value
	}
	return m, nil
}

//ReadObject read object
func (d *Decoder) ReadObject() (interface{}, error) {
	tag, err := d.readBufByte()
	if err != nil {
		hlog.Debugf("reading tag err:%v", err)
		return nil, nil //ignore
	}
	switch {
	case tag == BcEnd:
		return nil, io.EOF
	case tag == BcNull:
		return nil, nil
	case tag == BcTrue:
		return true, nil
	case tag == BcFalse:
		return false, nil
	case tag == BcMap:
		return d.ReadTypedMap()
	case tag == BcMapUntyped:
		return d.ReadUntypedMap()
	case tag == BcObjectDef:
		clsDef, err := d.readClassDef()
		if err != nil {
			return nil, newCodecError("byte double", err)
		}
		clsD, _ := clsDef.(ClassDef)
		//add to slice
		d.clsDefList = append(d.clsDefList, clsD)
		//read from refList of ClassDef
		return d.ReadObject()
	case tag == BcObject:
		i, _ := d.readInt(TagRead)
		idx := int(i)
		clsD := d.clsDefList[idx]
		typ, ok := d.typMap[clsD.FullClassName]
		if !ok {
			return nil, newCodecError("undefined type for "+clsD.FullClassName, err)
		}
		return UnpackValue(d.readInstance(typ, clsD))
	case (tag >= 0x80 && tag <= 0xbf) || (tag >= 0xc0 && tag <= 0xcf) ||
		(tag >= 0xd0 && tag <= 0xd7) || (tag == BcInt):
		return d.readInt(int32(tag))

	case (tag >= 0xd8 && tag <= 0xef) || (tag >= 0xf4 && tag <= 0xff) ||
		(tag >= 0x38 && tag <= 0x3f) || (tag == BcLongInt) ||
		(tag == BcLong):
		return d.readLong(int32(tag))
	case tag == BcDoubleZero:
		return float64(0), nil
	case tag == BcDoubleOne:
		return float64(1), nil
	case tag == BcDoubleByte:
		bf1 := make([]byte, 1)
		if _, err := io.ReadFull(d.reader, bf1); err != nil {
			return nil, newCodecError("byte double", err)
		}
		i := float64(bf1[0])
		return i, nil
	case tag == BcDoubleShort:
		bf1 := make([]byte, 2)
		if _, err := io.ReadFull(d.reader, bf1); err != nil {
			return nil, newCodecError("short long", err)
		}
		i := float64(bf1[0])*256 + float64(bf1[0])
		return i, nil
	case tag == BcDoubleMill:
		t, err := d.readInt(int32(tag))
		if err == nil {
			return t, err
		}
		return nil, newCodecError("double mill", err)
	case tag == BcDouble:
		return d.readDouble(int32(tag))
	case tag == BcDate:
		_, err := d.readLong(int32(tag))
		if err != nil {
			return nil, newCodecError("date", err)
		}
		return nil, newCodecError("date decode not yet implemented")
	case tag == BcDateMinute:
		return nil, newCodecError("date minute decode not yet implemented")
	case strTag(tag):
		return d.readString(int32(tag))
	case tag >= 0x60 && tag <= 0x6f:
		i := int(tag - 0x60)
		clsD := d.clsDefList[i]
		typ, ok := d.typMap[clsD.FullClassName]
		if !ok {
			return nil, newCodecError("undefined type for "+clsD.FullClassName, err)
		}
		return UnpackValue(d.readInstance(typ, clsD))
	case binaryTag(tag):
		return d.readBinary(int32(tag))
	case (tag >= BcListDirect && tag <= 0x77) || (tag == BcListFixed || tag == BcListVariable):
		styp, err := d.readType()
		if err != nil {
			return nil, newCodecError("ReadType", err)
		}
		var i int
		if tag >= BcListDirect && tag <= 0x77 {
			i = int(tag - BcListDirect)
		} else {
			ii, err := d.readInt(TagRead)
			if err != nil {
				return nil, newCodecError("ReadType", err)
			}
			i = int(ii)
		}
		isBuildInType(styp)

		// read first array item
		it, err := d.ReadObject()
		if err != nil {
			return nil, newCodecError("ReadList", err)
		}

		// return when no element
		if i <= 0 || it == nil {
			return []interface{}{}, nil
		}

		aryType := reflect.SliceOf(reflect.TypeOf(it))
		aryValue := reflect.MakeSlice(aryType, i, i)
		aryValue.Index(0).Set(reflect.ValueOf(it))
		for j := 1; j < i; j++ {
			it, err := d.ReadObject()
			if err != nil {
				return nil, newCodecError("ReadList", err)
			}
			aryValue.Index(j).Set(reflect.ValueOf(it))
		}

		return aryValue.Interface(), nil
	case (tag >= BcListDirectUntyped && tag <= 0x7f) || (tag == BcListFixedUntyped || tag == BcListVariableUntyped):
		var i int
		if tag >= BcListDirectUntyped && tag <= 0x7f {
			i = int(tag - BcListDirectUntyped)
		} else {
			ii, err := d.readInt(TagRead)
			if err != nil {
				return nil, newCodecError("ReadType", err)
			}
			i = int(ii)
		}
		ary := make([]interface{}, i)
		for j := 0; j < i; j++ {
			it, err := d.ReadObject()
			if err != nil {
				if err == io.EOF {
					continue
				}
				return nil, newCodecError("ReadList", err)
			}
			ary[j] = it
		}

		if tag == BcListVariableUntyped {
			// read list end tag 'Z'
			d.readBufByte()
		}
		return ary, nil
	default:
		return nil, newCodecError("unkonw tag")
	}
}

func (d *Decoder) readClassDef() (interface{}, error) {
	clsName, err := d.readString(TagRead)
	if err != nil {
		return nil, newCodecError("ReadClassDef", err)
	}
	count, err := d.readInt(TagRead)
	if err != nil {
		return nil, newCodecError("ReadClassDef", err)
	}

	fields := make([]string, count)
	for i := 0; i < int(count); i++ {
		s, err := d.readString(TagRead)
		if err != nil {
			return nil, newCodecError("ReadClassDef", err)
		}
		fields[i] = s
	}
	cls := ClassDef{clsName, fields}
	return cls, nil
}
