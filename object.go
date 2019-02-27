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
	"fmt"
	"io"
	"reflect"
	"strings"
)

const (
	ObjectTag       = byte('O')
	ObjectDefTag    = byte('C')
	ObjectLenTagMin = byte(0x60)
	ObjectLenTagMax = byte(0x6f)
	ObjectTagMaxLen = ObjectLenTagMax - ObjectLenTagMin
)

func objectLenTag(tag byte) bool {
	return tag >= ObjectLenTagMin && tag <= ObjectLenTagMax
}

//see: http://hessian.caucho.com/doc/hessian-serialization.html##object
func (e *Encoder) writeObject(data interface{}) (int, error) {
	typ := reflect.TypeOf(data)
	vv := reflect.ValueOf(data)

	clsName, ok := e.nameMap[typ.Name()]
	if !ok {
		clsName = typ.Name()
		e.nameMap[clsName] = clsName
	}
	length, ok := e.existClassDef(clsName)
	if !ok {
		length, _ = e.writeClsDef(typ, clsName)
	}
	if byte(length) <= ObjectTagMaxLen && length != 2 {
		// 20181231 NOTE: when length=2, length+ObjectLenTagMin='b', the same as the binary chunk start at.
		e.writeBT(byte(length) + ObjectLenTagMin)
	} else {
		e.writeBT(ObjectTag)
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
	e.writeBT(ObjectDefTag)
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

//ReadTagObject read tag object
func (d *Decoder) ReadTagObject() (interface{}, error) {
	i, _ := d.readInt(TagRead)
	idx := int(i)
	clsD := d.clsDefList[idx]
	typ, ok := d.typMap[clsD.FullClassName]
	if !ok {
		return nil, newCodecError("ReadTagObject", "undefined type for "+clsD.FullClassName)
	}
	return UnpackValue(d.readObject(typ, clsD))
}

//ReadLenTagObject read length tag object
func (d *Decoder) ReadLenTagObject(tag byte) (interface{}, error) {
	i := int(tag - ObjectLenTagMin)
	clsD := d.clsDefList[i]
	typ, ok := d.typMap[clsD.FullClassName]
	if !ok {
		return nil, newCodecError("ReadLenTagObject", "undefined type for "+clsD.FullClassName)
	}
	return UnpackValue(d.readObject(typ, clsD))
}

//ReadObjectDef read object def
func (d *Decoder) ReadObjectDef() (interface{}, error) {
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

	if tag == ObjectTag {
		return d.ReadTagObject()
	}
	return nil, newCodecError("unknown tag after class def ", tag)
}

func (d *Decoder) readObject(typ reflect.Type, cls ClassDef) (interface{}, error) {
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

		err = d.readField(fldName, fldValue)
		if err != nil {
			return nil, newCodecError("failed to decode field "+fldName, err)
		}
	}
	return vv, nil
}

func (d *Decoder) readField(fldName string, fldValue reflect.Value) error {
	kind := fldValue.Kind()
	switch kind {
	case reflect.String:
		str, err := d.readString(TagRead)
		if err != nil {
			return err
		}
		if str != "" {
			fldValue.SetString(str)
		}
	case reflect.Int32, reflect.Int, reflect.Int16, reflect.Int8:
		i, err := d.readInt(TagRead)
		if err != nil {
			return err
		}
		v := int64(i)
		fldValue.SetInt(v)
	case reflect.Uint8, reflect.Uint16:
		i, err := d.readInt(TagRead)
		if err != nil {
			return err
		}
		v := uint64(i)
		fldValue.SetUint(v)
	case reflect.Int64:
		i, err := d.readLong(TagRead)
		if err != nil {
			return err
		}
		fldValue.SetInt(i)
	case reflect.Uint64, reflect.Uint, reflect.Uint32:
		i, err := d.readLong(TagRead)
		if err != nil {
			return err
		}
		fldValue.SetUint(uint64(i))
	case reflect.Bool:
		b, err := d.ReadData()
		if err != nil {
			return err
		}
		fldValue.SetBool(b.(bool))
	case reflect.Float32, reflect.Float64:
		f, err := d.readDouble(TagRead)
		if err != nil {
			return err
		}
		fldValue.SetFloat(f)
	case reflect.Struct:
		s, err := d.ReadTypeData(reflect.Struct)
		if err != nil {
			return err
		}
		fldValue.Set(reflect.Indirect(reflect.ValueOf(s)))
	case reflect.Map:
		d.readMap(fldValue)
	case reflect.Slice, reflect.Array:
		m, err := d.ReadList()
		if err != nil {
			if err == io.EOF {
				break // ignore nil slice
			}
			return err
		}
		err = SetSlice(fldValue, m)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported field: %s, type: %v", fldName, fldValue.Type())
	}

	return nil
}
