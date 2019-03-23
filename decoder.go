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

type ByteRuneReader interface {
	io.Reader
	io.RuneReader
}

// ClassDef class def
type ClassDef struct {
	FullClassName string
	FieldName     []string
}

//Decoder type
type Decoder struct {
	reader     ByteRuneReader
	typMap     map[string]reflect.Type
	typList    []string
	refList    []reflect.Value
	clsDefList []ClassDef
}

//NewDecoder new
func NewDecoder(r ByteRuneReader, typ map[string]reflect.Type) *Decoder {
	if typ == nil {
		typ = make(map[string]reflect.Type, 11)
	}
	decode := &Decoder{
		typMap: typ,
	}
	if r != nil {
		decode.Reset(r)
	}
	return decode
}

//Reset reset
func (d *Decoder) Reset(r ByteRuneReader) {
	d.reader = r
	d.typList = make([]string, 0, 11)
	d.clsDefList = make([]ClassDef, 0, 11)
	d.refList = make([]reflect.Value, 0, 11)
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

func (d *Decoder) readTag() (byte, error) {
	return readTag(d.reader)
}

func (d *Decoder) readBytes(size int) ([]byte, error) {
	return readBytes(d.reader, size)
}

//Decode decode bytes to object
func (d *Decoder) Decode(bts []byte) (interface{}, error) {
	buf := bufio.NewReader(bytes.NewReader(bts))
	return d.ReadFrom(buf)
}

//ReadObject read new object from reader
func (d *Decoder) ReadObject() (interface{}, error) {
	return EnsureInterface(d.ReadData())
}

//ReadFrom read object from target reader
func (d *Decoder) ReadFrom(reader ByteRuneReader) (interface{}, error) {
	d.Reset(reader)
	return d.ReadObject()
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
	case typedListTag(tag) || untypedListTag(tag):
		return d.ReadList(int32(tag))
	default:
		return nil, newCodecError("readData", "unknown tag: 0x%x", tag)
	}
}
