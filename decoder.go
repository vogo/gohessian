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
	"bufio"
	"bytes"
	"io"
	"reflect"
	"time"
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
	reader     *bufio.Reader
	typMap     map[string]reflect.Type
	typList    []string
	refList    []reflect.Value
	clsDefList []ClassDef
}

//NewDecoder new
func NewDecoder(r *bufio.Reader, typ map[string]reflect.Type) *Decoder {
	if typ == nil {
		typ = make(map[string]reflect.Type, 17)
	}
	decode := &Decoder{
		typMap: typ,
	}
	decode.Reset(r)
	return decode
}

//Decode object
func (d *Decoder) Decode() (interface{}, error) {
	return EnsureInterface(d.ReadData())
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
func (d *Decoder) Reset(r *bufio.Reader) {
	d.reader = r
	d.typList = make([]string, 0, 17)
	d.clsDefList = make([]ClassDef, 0, 17)
	d.refList = make([]reflect.Value, 0, 17)
}

func (d *Decoder) readTag() (byte, error) {
	return readTag(d.reader)
}

func (d *Decoder) readBytes(size int) ([]byte, error) {
	return readBytes(d.reader, size)
}

//ReadObjectWithType name is option, if it is nil, use type.Name()
func (d *Decoder) ReadObjectWithType(typ reflect.Type, name string) (interface{}, error) {
	//register the type if it did exist
	if _, ok := d.typMap[name]; ok {
		hlog.Debugf("overwrite existing type: %s", name)
	}
	d.typMap[name] = typ
	return EnsureInterface(d.ReadData())
}

func (d *Decoder) readBoolean(flag int32) (bool, error) {
	return decodeBooleanValue(d.reader, flag)
}

func (d *Decoder) readBinary(flag int32) ([]byte, error) {
	return decodeBinaryValue(d.reader, flag)
}

func (d *Decoder) readInt(flag int32) (int32, error) {
	return decodeIntValue(d.reader, flag)
}

func (d *Decoder) readLong(flag int32) (int64, error) {
	return decodeLongValue(d.reader, flag)
}

func (d *Decoder) readDouble(flag int32) (float64, error) {
	return decodeDoubleValue(d.reader, flag)
}

func (d *Decoder) readString(flag int32) (string, error) {
	return decodeStringValue(d.reader, flag)
}

func (d *Decoder) readDate(flag int32) (time.Time, error) {
	return decodeDateValue(d.reader, flag)
}

func (d *Decoder) readStruct() (interface{}, error) {
	tag, err := d.readTag()
	if err != nil {
		hlog.Debugf("reading tag err:%v", err)
		return nil, nil //ignore
	}

	switch {
	case tag == _endFlag:
		return nil, io.EOF
	case tag == _nilTag:
		return nil, nil
	case dateTag(tag):
		return d.readDate(int32(tag))
	case tag == _objectDefTag:
		return d.readObjectDef()
	case objectLenTag(tag):
		return d.ReadLenTagObject(tag)
	case tag == _objectTag:
		return d.readTagObject()
	case refTag(tag):
		return d.readRef(tag)
	default:
		return nil, newCodecError("readStruct", "unknown tag: 0x%x", tag)
	}
}

//ReadData read object
func (d *Decoder) ReadData() (interface{}, error) {
	tag, err := d.readTag()
	if err != nil {
		hlog.Debugf("reading tag err:%v", err)
		return nil, nil //ignore
	}

	switch {
	case tag == _endFlag:
		return nil, io.EOF
	case tag == _nilTag:
		return nil, nil
	case tag == _boolTrueTag:
		return true, nil
	case tag == _boolFalseTag:
		return false, nil
	case intTag(tag):
		return d.readInt(int32(tag))
	case longTag(tag):
		return d.readLong(int32(tag))
	case doubleTag(tag):
		return d.readDouble(int32(tag))
	case stringTag(tag):
		return d.readString(int32(tag))
	case dateTag(tag):
		return d.readDate(int32(tag))
	case binaryTag(tag):
		return d.readBinary(int32(tag))
	case refTag(tag):
		return d.readRef(tag)
	case tag == _mapTypedTag:
		return d.readTypedMap()
	case tag == _mapUntypedTag:
		return d.readUntypedMap()
	case tag == _objectDefTag:
		return d.readObjectDef()
	case objectLenTag(tag):
		return d.ReadLenTagObject(tag)
	case tag == _objectTag:
		return d.readTagObject()
	case typedListTag(tag):
		return d.readTypedList(tag)
	case untypedListTag(tag):
		return d.readUntypedList(tag)
	default:
		return nil, newCodecError("readData", "unknown tag: 0x%x", tag)
	}
}
