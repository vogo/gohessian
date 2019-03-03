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

// see: http://hessian.caucho.com/doc/hessian-serialization.html##map
package hessian

import (
	"io"
	"reflect"
)

const (
	MapTypedTag   = byte('M')
	MapUntypedTag = byte('H')
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
		e.writeBT(MapTypedTag)
		e.writeString(mapName)
	} else {
		e.writeBT(MapUntypedTag)
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

	e.writeBT(EndFlag)

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

//readUntypedMap read untyped map
func (d *Decoder) readUntypedMap() (interface{}, error) {
	m := make(map[interface{}]interface{})
	d.addDecoderRef(reflect.ValueOf(&m))

	//read key and value
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

	if tag == MapTypedTag {
		d.readString(TagRead)
	} else if tag == MapUntypedTag {
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
		m.SetMapIndex(EnsurePackValue(key), EnsurePackValue(vl))
	}
	SetValue(dest, mapValue)
	return nil
}
