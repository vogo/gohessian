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
// Each list item is added to the reference list to handle shared and circular elements. See the ref element.
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
	"fmt"
	"io"
	"reflect"
)

const (
	ListVariableTypedTag   = byte(0x55)
	ListVariableUntypedTag = byte(0x57)

	ListFixedTypedStartTag    = byte('V')
	ListFixedUntypedTag       = byte(0x58)
	ListFixedTypedLenTagMin   = byte(0x70)
	ListFixedTypedLenTagMax   = byte(0x77)
	ListFixedUntypedLenTagMin = byte(0x78)
	ListFixedUntypedLenTagMax = byte(0x7f)
)

func listFixedTypedLenTag(tag byte) bool {
	return tag >= ListFixedTypedLenTagMin && tag <= ListFixedTypedLenTagMax
}

// Include 3 formats:
// list ::= x55 type value* 'Z'   # variable-length list
//      ::= 'V' type int value*   # fixed-length list
//      ::= [x70-77] type value*  # fixed-length typed list
func typedListTag(tag byte) bool {
	return tag == ListFixedTypedStartTag || tag == ListVariableTypedTag || listFixedTypedLenTag(tag)
}

func listFixedUntypedLenTag(tag byte) bool {
	return tag >= ListFixedUntypedLenTagMin && tag <= ListFixedUntypedLenTagMax
}

// Include 3 formats:
//      ::= x57 value* 'Z'        # variable-length untyped list
//      ::= x58 int value*        # fixed-length untyped list
//      ::= [x78-7f] value*       # fixed-length untyped list
func untypedListTag(tag byte) bool {
	return tag == ListFixedUntypedTag || tag == ListVariableUntypedTag || listFixedUntypedLenTag(tag)
}

// only write as fixed-length untyped list
func (e *Encoder) writeList(data interface{}) (int, error) {
	if bt, ok := data.([]byte); ok {
		return e.writeBinary(bt)
	}

	vv := reflect.ValueOf(data)
	e.writeBT(ListFixedUntypedTag)
	e.writeInt(int32(vv.Len()))
	for i := 0; i < vv.Len(); i++ {
		e.WriteData(vv.Index(i).Interface())
	}
	return vv.Len(), nil
}

//ReadList read list
func (d *Decoder) ReadList() (interface{}, error) {
	tag, err := d.readTag()
	if err != nil {
		hlog.Debugf("reading tag err:%v", err)
		return nil, nil //ignore
	}

	if binaryTag(tag) {
		return d.readBinary(int32(tag))
	}

	switch {
	case typedListTag(tag):
		return d.ReadTypedList(tag)
	case untypedListTag(tag):
		return d.ReadUntypedList(tag)
	default:
		return nil, fmt.Errorf("expect list tag but get %x", tag)
	}
}

// ReadTypedList read typed list
// Include 3 formats:
// list ::= x55 type value* 'Z'   # variable-length list
//      ::= 'V' type int value*   # fixed-length list
//      ::= [x70-77] type value*  # fixed-length typed list
func (d *Decoder) ReadTypedList(tag byte) (interface{}, error) {
	listTyp, err := d.readType()
	if err != nil {
		return nil, newCodecError("ReadType", err)
	}

	isVariableArr := tag == ListVariableTypedTag

	var length int
	if listFixedTypedLenTag(tag) {
		length = int(tag - ListFixedTypedLenTagMin)
	} else if tag == ListFixedTypedStartTag {
		ii, err := d.readInt(TagRead)
		if err != nil {
			return nil, newCodecError("ReadType", err)
		}
		length = int(ii)
	} else if isVariableArr {
		length = 1
	} else {
		return nil, fmt.Errorf("expect typed list tag, but get %x", tag)
	}

	isBuildInType(listTyp)

	// read first array item
	it, err := d.ReadData()
	if err != nil {
		return nil, newCodecError("ReadList", err)
	}

	// return when no element
	if length <= 0 || it == nil {
		return nil, nil
	}

	aryType := reflect.SliceOf(reflect.TypeOf(it))
	aryValue := reflect.MakeSlice(aryType, length, length)
	aryValue.Index(0).Set(reflect.ValueOf(it))

	for j := 1; j < length || isVariableArr; j++ {
		it, err := d.ReadData()
		if err != nil {
			if err == io.EOF && isVariableArr {
				break
			}
			return nil, newCodecError("ReadList", err)
		}

		v := reflect.ValueOf(it)
		if isVariableArr {
			aryValue = reflect.Append(aryValue, v)
		} else {
			aryValue.Index(j).Set(v)
		}
	}

	return aryValue.Interface(), nil
}

//ReadUntypedList read untyped list
// Include 3 formats:
//      ::= x57 value* 'Z'        # variable-length untyped list
//      ::= x58 int value*        # fixed-length untyped list
//      ::= [x78-7f] value*       # fixed-length untyped list
func (d *Decoder) ReadUntypedList(tag byte) (interface{}, error) {
	isVariableArr := tag == ListVariableUntypedTag

	var length int
	if listFixedUntypedLenTag(tag) {
		length = int(tag - ListFixedUntypedLenTagMin)
	} else if tag == ListFixedUntypedTag {
		ii, err := d.readInt(TagRead)
		if err != nil {
			return nil, newCodecError("ReadType", err)
		}
		length = int(ii)
	} else if isVariableArr {
		length = 0
	} else {
		return nil, fmt.Errorf("expect untyped list tag, but get %x", tag)
	}

	ary := make([]interface{}, length)
	for j := 0; j < length || isVariableArr; j++ {
		it, err := d.ReadData()
		if err != nil {
			if err == io.EOF && isVariableArr {
				continue
			}
			return nil, newCodecError("ReadList", err)
		}

		if isVariableArr {
			ary = append(ary, it)
		} else {
			ary[j] = it
		}
	}

	return ary, nil
}
