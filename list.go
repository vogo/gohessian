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

// see: http://hessian.caucho.com/doc/hessian-serialization.html##list
//
// List Grammar
//
// list ::= x55 type value* 'Z'   # variable-length list
//      ::= 'V' type int value*   # fixed-length list
//      ::= x57 value* 'Z'        # variable-length untyped list
//      ::= x58 int value*        # fixed-length untyped list
//      ::= [x70-77] type value*  # fixed-length typed list
//      ::= [x78-7f] value*       # fixed-length untyped list
//
//
// An ordered list, like an array. The two list productions are a fixed-length list and a variable length list.
// Both lists have a type. The type string may be an arbitrary UTF-8 string understood by the service.
//
// Each list item is added to the reference list to handle shared and circularT elements. See the ref element.
//
// Any parser expecting a list must also accept a null or a shared ref.
//
// The valid values of type are not specified in this document and may depend on the specific application.
// For example, a server implemented in a language with static typing which exposes an Hessian interface can use the type information to instantiate the specific array type.
// On the other hand, a server written in a dynamicly-typed language would likely ignore the contents of type entirely and create a generic array.
//
// ---------------------------
// fixed length list
// Hessian 2.0 allows a compact form of the list for successive lists of the same type where the length is known beforehand.
// The type and length are encoded by integers, where the type is a reference to an earlier specified type.
//
//
// ---------------------------
// List examples
//
// ----------> Serialization of a typed int array: int[] = {0, 1}
//
// V                    # fixed length, typed list
//   x04 [int           # encoding of int[] type
//   x92                # length = 2
//   x90                # integer 0
//   x91                # integer 1
//
// ----------> untyped variable-length list = {0, 1}
//
// x57                  # variable-length, untyped
//   x90                # integer 0
//   x91                # integer 1
//   Z
//
// ----------> fixed-length type
//
// x72                # typed list length=2
//   x04 [int         # type for int[] (save as type #0)
//   x90              # integer 0
//   x91              # integer 1
//
// x73                # typed list length = 3
//   x90              # type reference to int[] (integer #0)
//   x92              # integer 2
//   x93              # integer 3
//   x94              # integer 4

package hessian

import (
	"io"
	"reflect"
)

const (
	_listVariableTypedTag   = byte(0x55)
	_listVariableUntypedTag = byte(0x57)

	_listFixedTypedStartTag    = byte('V')
	_listFixedUntypedTag       = byte(0x58)
	_listFixedTypedLenTagMin   = byte(0x70)
	_listFixedTypedLenTagMax   = byte(0x77)
	_listFixedTypedLenMax      = _listFixedTypedLenTagMax - _listFixedTypedLenTagMin
	_listFixedUntypedLenTagMin = byte(0x78)
	_listFixedUntypedLenTagMax = byte(0x7f)
	_listFixedUntypedLenMax    = _listFixedUntypedLenTagMax - _listFixedUntypedLenTagMin
)

func listFixedTypedLenTag(tag byte) bool {
	return tag >= _listFixedTypedLenTagMin && tag <= _listFixedTypedLenTagMax
}

// Include 3 formats:
// list ::= x55 type value* 'Z'   # variable-length list
//      ::= 'V' type int value*   # fixed-length list
//      ::= [x70-77] type value*  # fixed-length typed list
func typedListTag(tag byte) bool {
	return tag == _listFixedTypedStartTag || tag == _listVariableTypedTag || listFixedTypedLenTag(tag)
}

func listFixedUntypedLenTag(tag byte) bool {
	return tag >= _listFixedUntypedLenTagMin && tag <= _listFixedUntypedLenTagMax
}

// Include 3 formats:
//      ::= x57 value* 'Z'        # variable-length untyped list
//      ::= x58 int value*        # fixed-length untyped list
//      ::= [x78-7f] value*       # fixed-length untyped list
func untypedListTag(tag byte) bool {
	return tag == _listFixedUntypedTag || tag == _listVariableUntypedTag || listFixedUntypedLenTag(tag)
}

// write as fixed-length list
func (e *Encoder) writeList(data interface{}) (int, error) {
	if bt, ok := data.([]byte); ok {
		return e.writeBinary(bt)
	}

	// object data MUST not be unpacked
	vv := reflect.ValueOf(data)

	// check ref
	if n, ok := e.checkEncodeRefMap(vv); ok {
		return e.writeRef(n)
	}

	// unpack to parser values
	vv = UnpackPtrValue(vv)

	typ := UnpackPtrType(vv.Type())
	arrayTypeName := TypeName(typ)
	listTypeName, ok := e.nameMap[arrayTypeName]

	if !ok || _interfaceTypeName == arrayRootElemName(arrayTypeName) {
		// fixed-length untyped list
		e.writeBT(_listFixedUntypedTag)
		e.writeInt(int32(vv.Len()))
	} else if byte(vv.Len()) <= _listFixedTypedLenMax {
		// fixed-length typed list
		e.writeBT(_listFixedTypedLenTagMin + byte(vv.Len()))
		e.writeString(listTypeName)
	} else {
		// fixed-length
		e.writeBT(_listFixedTypedStartTag)
		e.writeString(listTypeName)
		e.writeInt(int32(vv.Len()))
	}

	for i := 0; i < vv.Len(); i++ {
		e.WriteData(vv.Index(i).Interface())
	}
	return vv.Len(), nil
}

//ReadList read list
func (d *Decoder) ReadList(flag int32) (interface{}, error) {
	tag, err := getTag(d.reader, flag)
	if err != nil {
		hlog.Debugf("reading tag err:%v", err)
		return nil, nil //ignore
	}

	if binaryTag(tag) {
		return d.readBinary(int32(tag))
	}

	switch {
	case tag == _nilTag:
		return nil, nil
	case refTag(tag):
		return d.readRef(tag)
	case typedListTag(tag):
		return d.readTypedList(tag)
	case untypedListTag(tag):
		return d.readUntypedList(tag)
	default:
		return nil, newCodecError("ReadList", "error list tag: 0x%x", tag)
	}
}

// readTypedList read typed list
// Include 3 formats:
// list ::= x55 type value* 'Z'   # variable-length list
//      ::= 'V' type int value*   # fixed-length list
//      ::= [x70-77] type value*  # fixed-length typed list
func (d *Decoder) readTypedList(tag byte) (interface{}, error) {
	listTyp, err := d.readType()
	if err != nil {
		return nil, newCodecError("readTypedList", "read list type: %s", listTyp, err)
	}

	isVariableArr := tag == _listVariableTypedTag

	length := -1
	switch {
	case isVariableArr:
		length = 0
	case listFixedTypedLenTag(tag):
		length = int(tag - _listFixedTypedLenTagMin)
	case tag == _listFixedTypedStartTag:
		ii, err := d.readInt(_tagRead)
		if err != nil {
			return nil, newCodecError("readTypedList", err)
		}
		length = int(ii)
	default:
		return nil, newCodecError("readTypedList", "error typed list tag: 0x%x", tag)
	}

	// return when no element
	if length < 0 {
		return nil, nil
	}

	aryType, ok := d.typMap[listTyp]
	if !ok {
		return nil, newCodecError("readTypedList", "can't find list type %s", listTyp)
	}

	aryValue := reflect.MakeSlice(aryType, length, length)
	holder := d.addDecoderRef(aryValue)

	for j := 0; j < length || isVariableArr; j++ {
		item, err := d.ReadData()
		if err != nil {
			if err == io.EOF && isVariableArr {
				break
			}
			return nil, newCodecError("readTypedList", err)
		}

		if item == nil {
			break
		}

		v := EnsureRawValue(item)
		if isVariableArr {
			aryValue = reflect.Append(aryValue, v)
			holder.change(aryValue)
		} else {
			SetValue(aryValue.Index(j), v)
		}
	}

	return holder, nil
}

//readUntypedList read untyped list
// Include 3 formats:
//      ::= x57 value* 'Z'        # variable-length untyped list
//      ::= x58 int value*        # fixed-length untyped list
//      ::= [x78-7f] value*       # fixed-length untyped list
func (d *Decoder) readUntypedList(tag byte) (interface{}, error) {
	isVariableArr := tag == _listVariableUntypedTag

	length := -1

	switch {
	case isVariableArr:
		length = 0
	case listFixedUntypedLenTag(tag):
		length = int(tag - _listFixedUntypedLenTagMin)
	case tag == _listFixedUntypedTag:
		ii, err := d.readInt(_tagRead)
		if err != nil {
			return nil, newCodecError("readUntypedList", err)
		}
		length = int(ii)
	default:
		return nil, newCodecError("readUntypedList", "error untyped list tag: %x", tag)
	}

	// return when no element
	if length < 0 {
		return nil, nil
	}

	ary := make([]interface{}, length)
	aryValue := reflect.ValueOf(ary)
	holder := d.addDecoderRef(aryValue)

	for j := 0; j < length || isVariableArr; j++ {
		it, err := d.ReadData()
		if err != nil {
			if err == io.EOF && isVariableArr {
				continue
			}
			return nil, newCodecError("readUntypedList", err)
		}

		if isVariableArr {
			aryValue = reflect.Append(aryValue, EnsureRawValue(it))
			holder.change(aryValue)
		} else {
			ary[j] = it
		}
	}

	return holder, nil
}
