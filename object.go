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

// see http://hessian.caucho.com/doc/hessian-serialization.html##object
//
// Object Grammar
//
// class-def  ::= 'C' string int string*
//
// object     ::= 'O' int value*
//            ::= [x60-x6f] value*
//
//
// -------------------------
// class definition
// Hessian 2.0 has a compact object form where the field names are only serialized once.
// Following objects only need to serialize their values.
//
// The object definition includes a mandatory type string, the number of fields, and the field names.
// The object definition is stored in the object // definition map and will be referenced by object instances with an integer reference.
//
//
// -------------------------
// object instantiation
// Hessian 2.0 has a compact object form where the field names are only serialized once.
// Following objects only need to serialize their values.
//
// The object instantiation creates a new object based on a previous definition.
// The integer value refers to the object definition.
//
//
// -------------------------
// Object examples
//
// Object serialization
//
// class Car {
//   String color;
//   String model;
// }
//
// out.writeObject(new Car("red", "corvette"));
// out.writeObject(new Car("green", "civic"));
//
// ---
//
// C                        # object definition (#0)
//   x0b example.Car        # type is example.Car
//   x92                    # two fields
//   x05 color              # color field name
//   x05 model              # model field name
//
// O                        # object def (long form)
//   x90                    # object definition #0
//   x03 red                # color field value
//   x08 corvette           # model field value
//
// x60                      # object def #0 (short form)
//   x05 green              # color field value
//   x05 civic              # model field value
// ------------------------------------------------------
//  enum Color {
//   RED,
//   GREEN,
//   BLUE,
// }
//
// out.writeObject(Color.RED);
// out.writeObject(Color.GREEN);
// out.writeObject(Color.BLUE);
// out.writeObject(Color.GREEN);
//
// ---
//
// C                         # class definition #0
//   x0b example.Color       # type is example.Color
//   x91                     # one field
//   x04 name                # enumeration field is "name"
//
// x60                       # object #0 (class def #0)
//   x03 RED                 # RED value
//
// x60                       # object #1 (class def #0)
//   x90                     # object definition ref #0
//   x05 GREEN               # GREEN value
//
// x60                       # object #2 (class def #0)
//   x04 BLUE                # BLUE value
//
// x51 x91                   # object ref #1, i.e. Color.GREEN

package hessian

import (
	"io"
	"reflect"
	"strings"
	"time"
)

const (
	_objectTag       = byte('O')
	_objectDefTag    = byte('C')
	_objectLenTagMin = byte(0x60)
	_objectLenTagMax = byte(0x6f)
	_objectTagMaxLen = _objectLenTagMax - _objectLenTagMin
)

func objectLenTag(tag byte) bool {
	return tag >= _objectLenTagMin && tag <= _objectLenTagMax
}

//see: http://hessian.caucho.com/doc/hessian-serialization.html##object
func (e *Encoder) writeObject(data interface{}) (int, error) {
	// object data MUST not be unpacked
	vv := reflect.ValueOf(data)

	// check ref
	if n, ok := e.checkEncodeRefMap(vv); ok {
		return e.writeRef(n)
	}

	vv = UnpackPtrValue(vv)

	// check date type for date is a struct
	if date, ok := vv.Interface().(time.Time); ok {
		return e.writeBytes(encodeDate(date))
	}

	typ := vv.Type()
	clsName, ok := e.nameMap[typ.Name()]
	if !ok {
		clsName = typ.Name()
		e.nameMap[clsName] = clsName
	}
	length, ok := e.existClassDef(clsName)
	if !ok {
		length, _ = e.writeClsDef(typ, clsName)
	}
	if byte(length) <= _objectTagMaxLen {
		// NOTE: when length=2, length+_objectLenTagMin='b', the same as the binary chunk start with,
		// which will be special processed in decoder
		e.writeBT(byte(length) + _objectLenTagMin)
	} else {
		e.writeBT(_objectTag)
		e.writeInt(int32(length))
	}
	for i := 0; i < vv.NumField(); i++ {
		_, err := e.WriteData(vv.Field(i).Interface())
		if err != nil {
			return 0, err
		}
	}
	return vv.NumField(), nil
}

func (e *Encoder) writeClsDef(typ reflect.Type, clsName string) (int, error) {
	e.writeBT(_objectDefTag)
	e.writeString(clsName)
	fldList := make([]string, typ.NumField())
	e.writeInt(int32(len(fldList)))
	for i := 0; i < len(fldList); i++ {
		str, _ := lowerName(typ.Field(i).Name)
		fldList[i] = str
		e.writeString(fldList[i])
	}
	clsDef := ClassDef{clsName, fldList}
	length := len(e.clsDefList)
	e.clsDefList = append(e.clsDefList, clsDef)
	return length, nil
}

func (e *Encoder) existClassDef(clsName string) (int, bool) {
	for i := 0; i < len(e.clsDefList); i++ {
		if strings.Compare(clsName, e.clsDefList[i].FullClassName) == 0 {
			return i, true
		}
	}
	return 0, false
}

func (d *Decoder) readClassDef() (interface{}, error) {
	clsName, err := d.readString(_tagRead)
	if err != nil {
		return nil, newCodecError("ReadClassDef", err)
	}
	count, err := d.readInt(_tagRead)
	if err != nil {
		return nil, newCodecError("ReadClassDef", err)
	}

	fields := make([]string, count)
	for i := 0; i < int(count); i++ {
		s, err := d.readString(_tagRead)
		if err != nil {
			return nil, newCodecError("ReadClassDef", err)
		}
		fields[i] = s
	}
	cls := ClassDef{clsName, fields}
	return cls, nil
}

//readTagObject read tag object
func (d *Decoder) readTagObject() (interface{}, error) {
	i, _ := d.readInt(_tagRead)
	idx := int(i)
	clsD := d.clsDefList[idx]
	typ, ok := d.typMap[clsD.FullClassName]
	if !ok {
		return nil, newCodecError("readTagObject", "undefined type: %s", clsD.FullClassName)
	}
	return EnsureInterface(d.readObject(typ, clsD))
}

//ReadLenTagObject read length tag object
func (d *Decoder) ReadLenTagObject(tag byte) (interface{}, error) {
	i := int(tag - _objectLenTagMin)
	if i >= len(d.clsDefList) {
		return nil, newCodecError("ReadLenTagObject", "cls def ref index %d over max %d", i, len(d.clsDefList))
	}
	clsD := d.clsDefList[i]
	typ, ok := d.typMap[clsD.FullClassName]
	if !ok {
		return nil, newCodecError("ReadLenTagObject", "undefined type: %s", clsD.FullClassName)
	}
	return EnsureInterface(d.readObject(typ, clsD))
}

//readObjectDef read object def
func (d *Decoder) readObjectDef() (interface{}, error) {
	clsDef, err := d.readClassDef()
	if err != nil {
		return nil, err
	}
	clsD, _ := clsDef.(ClassDef)
	//add to slice
	d.clsDefList = append(d.clsDefList, clsD)

	tag, err := d.readTag()
	if err != nil {
		hlog.Debugf("reading tag err:%v", err)
		return nil, nil //ignore
	}

	if objectLenTag(tag) {
		return d.ReadLenTagObject(tag)
	}

	if tag == _objectTag {
		return d.readTagObject()
	}
	return nil, newCodecError("readObjectDef", "unknown tag after class def: 0x%x", tag)
}

// var readObjectIndex = 0

func (d *Decoder) readObject(typ reflect.Type, cls ClassDef) (interface{}, error) {
	if typ.Kind() != reflect.Struct {
		return nil, newCodecError("readObject", "expect type struct but get %v", typ)
	}
	vv := reflect.New(typ)
	d.addDecoderRef(vv)

	// readObjectIndex++
	// readObjectIndexCurr := readObjectIndex

	st := vv.Elem()
	for i := 0; i < len(cls.FieldName); i++ {
		fldName := cls.FieldName[i]
		index, err := findField(fldName, typ)

		// fmt.Printf("[%d]  >>>> start read field %s: %v, %v, %p\n", readObjectIndexCurr, fldName, vv.Type(), vv.Interface(), vv.Interface())
		if err != nil {
			hlog.Debugf("%s is not found, will skip type ->p %v", fldName, typ)
			continue
		}
		fldValue := st.Field(index)
		if !fldValue.CanSet() {
			return nil, newCodecError("readObject", "field %s can set", fldName)
		}

		err = d.readField(fldName, fldValue)
		if err != nil {
			return nil, newCodecError("readObject", "failed to decode field '%s'", fldName, err)
		}

		// fmt.Printf("[%d]  <<<<<< end read field %s: %v, %v, %p\n", readObjectIndexCurr, fldName, vv.Type(), vv.Interface(), vv.Interface())
	}
	return vv, nil
}

func (d *Decoder) readField(fldName string, fldValue reflect.Value) error {
	sourceValue := fldValue
	typ := UnpackPtrType(fldValue.Type())
	fldValue = UnpackPtrValue(fldValue)
	switch typ.Kind() {
	case reflect.String:
		str, err := d.readString(_tagRead)
		if err != nil {
			return err
		}
		if str != "" {
			fldValue.SetString(str)
		}
	case reflect.Int32, reflect.Int, reflect.Int16, reflect.Int8:
		i, err := d.readInt(_tagRead)
		if err != nil {
			return err
		}
		v := int64(i)
		fldValue.SetInt(v)
	case reflect.Uint8, reflect.Uint16:
		i, err := d.readInt(_tagRead)
		if err != nil {
			return err
		}
		v := uint64(i)
		fldValue.SetUint(v)
	case reflect.Int64:
		i, err := d.readLong(_tagRead)
		if err != nil {
			return err
		}
		fldValue.SetInt(i)
	case reflect.Uint64, reflect.Uint, reflect.Uint32:
		i, err := d.readLong(_tagRead)
		if err != nil {
			return err
		}
		fldValue.SetUint(uint64(i))
	case reflect.Bool:
		b, err := d.readBoolean(_tagRead)
		if err != nil {
			return err
		}
		fldValue.SetBool(b)
	case reflect.Float32, reflect.Float64:
		f, err := d.readDouble(_tagRead)
		if err != nil {
			return err
		}
		fldValue.SetFloat(f)
	case reflect.Struct:
		s, err := d.readStruct()
		if err != nil {
			return err
		}
		SetValue(sourceValue, EnsureRawValue(s))
	case reflect.Map:
		return d.readMap(sourceValue)
	case reflect.Slice, reflect.Array:
		m, err := d.ReadList(_tagRead)
		if err != nil {
			if err == io.EOF {
				break // ignore nil slice
			}
			return err
		}
		err = SetSlice(sourceValue, m)
		if err != nil {
			return err
		}
	default:
		return newCodecError("readField", "unsupported field: %s, type: %v, kind: %v", fldName, sourceValue.Type(), typ.Kind())
	}

	return nil
}
