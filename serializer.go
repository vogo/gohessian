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
	"bufio"
	"bytes"
	"io"
	"reflect"
)

//Serializer serializer
type Serializer interface {
	WriteObject(w io.Writer, object interface{}) error
	Write(object interface{}) error
	ToBytes(interface{}) ([]byte, error)
	ReadObject(r *bufio.Reader) (interface{}, error)
	Read() (interface{}, error)
	ToObject([]byte) (interface{}, error)
}

// goHessian the serializer cache struct, which will cache the type map, name map, and the encoder and decoder
type goHessian struct {
	typMap  map[string]reflect.Type
	nameMap map[string]string
	encoder *Encoder
	decoder *Decoder
}

//NewGoHessian init
func NewGoHessian(typMap map[string]reflect.Type, nmeMap map[string]string) Serializer {
	if typMap == nil {
		typMap = make(map[string]reflect.Type, 17)
	}
	if nmeMap == nil {
		nmeMap = make(map[string]string, 17)
	}
	return &goHessian{typMap: typMap, nameMap: nmeMap,}
}

// WriteObject to writer
func (gh *goHessian) WriteObject(w io.Writer, object interface{}) error {
	if gh.encoder == nil {
		gh.encoder = NewEncoder(w, gh.nameMap)
	} else {
		gh.encoder.Reset(w)
	}
	_, err := gh.encoder.WriteData(object)
	return err
}

// Write object to writer continuously , it must be called after calling goHessian.WriteObject
func (gh *goHessian) Write(object interface{}) error {
	if gh.encoder == nil {
		return newCodecError("Write", "encoder is nil")
	}
	_, err := gh.encoder.WriteData(object)
	return err
}

// ToBytes convert object to bytes
func (gh *goHessian) ToBytes(object interface{}) ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	err := gh.WriteObject(buffer, object)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// ReadObject from reader
func (gh *goHessian) ReadObject(r *bufio.Reader) (interface{}, error) {
	if gh.decoder == nil {
		gh.decoder = &Decoder{reader: r, typMap: gh.typMap}
	} else {
		gh.decoder.Reset(r)
	}
	return EnsureInterface(gh.decoder.ReadData())
}

// Read from reader continuously, it must be called after calling goHessian.ReadObject
func (gh *goHessian) Read() (interface{}, error) {
	if gh.decoder == nil {
		return nil, newCodecError("Read", "decoder is nil")
	}
	return EnsureInterface(gh.decoder.ReadData())
}

// ToObject convert bytes to object
func (gh *goHessian) ToObject(ins []byte) (interface{}, error) {
	return gh.ReadObject(bufio.NewReader(bytes.NewReader(ins)))
}

// ---------------------------------------------

//ToBytes [NO-CACHE API] serialize object to bytes
func ToBytes(object interface{}, nameMap map[string]string) ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	e := NewEncoder(buffer, nameMap)
	_, err := e.WriteData(object)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

//ToObject [NO-CACHE API] deserialize bytes to object
func ToObject(ins []byte, typMap map[string]reflect.Type) (interface{}, error) {
	ioBuf := bufio.NewReader(bytes.NewReader(ins))
	d := NewDecoder(ioBuf, typMap)
	return EnsureInterface(d.ReadData())
}
