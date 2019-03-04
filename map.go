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
//
// see: http://hessian.caucho.com/doc/hessian-serialization.html##map
//
// Map Grammar
//
// map        ::= M type (value value)* Z
//
// Represents serialized maps and can represent objects. The type element describes the type of the map.
// The type may be empty, i.e. a zero length.
// The parser is responsible for choosing a type if one is not specified. For objects, unrecognized keys will be ignored.
// Each map is added to the reference list. Any time the parser expects a map, it must also be able to support a null or a ref.
// The type is chosen by the service.
//
//
// -------------- Map examples
//
// A sparse array
//
// map = new HashMap();
// map.put(new Integer(1), "fee");
// map.put(new Integer(16), "fie");
// map.put(new Integer(256), "foe");
//
// ---
//
// H           # untyped map (HashMap for Java)
// x91       # 1
// x03 fee   # "fee"
//
// xa0       # 16
// x03 fie   # "fie"
//
// xc9 x00   # 256
// x03 foe   # "foe"
//
// Z
//
//  ----------------------------
// Map Representation of a Java Object
//
// public class Car implements Serializable {
// String color = "aquamarine";
// String model = "Beetle";
// int mileage = 65536;
// }
//
// ---
// M
// x13 com.caucho.test.Car  # type
//
// x05 color                # color field
// x0a aquamarine
//
// x05 model                # model field
// x06 Beetle
//
// x07 mileage              # mileage field
// I x00 x01 x00 x00
// Z

package hessian

import (
	"io"
	"reflect"
)

const (
	_mapTypedTag   = byte('M')
	_mapUntypedTag = byte('H')
)

func (e *Encoder) writeMap(data interface{}) (int, error) {
	// object data MUST not be unpacked
	vv := reflect.ValueOf(data)

	// check ref
	if n, ok := e.checkEncodeRefMap(vv); ok {
		return e.writeRef(n)
	}

	vv = UnpackPtrValue(vv)
	typ := vv.Type()

	mapName, ok := e.nameMap[typ.Name()]
	if ok {
		e.writeBT(_mapTypedTag)
		e.writeString(mapName)
	} else {
		e.writeBT(_mapUntypedTag)
	}

	count := 0

	if typ.Kind() == reflect.Map {
		// -------> untyped map
		keys := vv.MapKeys()
		count = len(keys)
		for i := 0; i < count; i++ {
			k := keys[i]
			_, err := e.WriteData(k.Interface())
			if err != nil {
				return 0, err
			}
			_, err = e.WriteData(vv.MapIndex(keys[i]).Interface())
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
			_, err := e.WriteData(f.Interface())
			if err != nil {
				return 0, err
			}
		}
	}

	e.writeBT(_endFlag)

	return count, nil
}

//readTypedMap read typed map
func (d *Decoder) readTypedMap() (interface{}, error) {
	typ, err := d.readType()
	if err != nil {
		return nil, newCodecError("ReadType", err)
	}
	mType, ok := d.typMap[typ]
	if !ok {
		return nil, newCodecError("ReadType", "no type map for %v", typ)
	}

	var mValue reflect.Value
	if mType.Kind() == reflect.Map {
		mValue = reflect.MakeMap(mType)
	} else {
		mValue = reflect.New(mType)
	}

	d.addDecoderRef(PackPtr(mValue))

	for {
		key, err := d.ReadData()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		value, err := d.ReadData()
		if err != nil {
			return nil, err
		}
		if mType.Kind() == reflect.Map {
			mValue.SetMapIndex(EnsureRawValue(key), EnsureRawValue(value))
		} else {
			fieldName, ok := key.(string)
			if !ok {
				return nil, newCodecError("readTypedMap", "the type of map key must be string, but get [%v]", key)
			}
			fieldValue := mValue.FieldByName(fieldName)
			if fieldValue.IsValid() {
				fieldValue.Set(EnsureRawValue(value))
			}
		}
	}

	m := mValue.Interface()
	return m, nil
}

//readUntypedMap read untyped map
func (d *Decoder) readUntypedMap() (interface{}, error) {
	m := make(map[interface{}]interface{})
	d.addDecoderRef(reflect.ValueOf(&m))

	//read key and value
	for {
		key, err := EnsureInterface(d.ReadData())
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err

		}
		value, err := EnsureInterface(d.ReadData())
		if err != nil {
			return nil, err
		}

		m[key] = value
	}
	return m, nil
}

func (d *Decoder) readMap(dest reflect.Value) error {
	tag, _ := d.readTag()

	// read ref value if ref
	if refTag(tag) {
		r, err := d.readRef(tag)
		if err != nil {
			return err
		}
		SetValue(dest, r)
		return nil
	}

	if tag == _mapTypedTag {
		d.readString(_tagRead)
	} else if tag == _mapUntypedTag {
		//do nothing
	} else {
		return newCodecError("readMap", "unknown map tag: %x", tag)
	}

	mapTyp := UnpackPtrType(dest.Type())
	m := reflect.MakeMap(mapTyp)
	mapValue := PackPtr(m)

	d.addDecoderRef(mapValue)

	//read key and value
	for {
		key, err := d.ReadData()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return newCodecError("readMap", err)
			}
		}
		vl, err := d.ReadData()
		if err != nil {
			return err
		}
		m.SetMapIndex(EnsureRawValue(key), EnsureRawValue(vl))
		//m.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(vl))
	}
	SetValue(dest, mapValue)
	return nil
}
