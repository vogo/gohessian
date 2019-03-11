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
	"io"
	"reflect"
)

//Serializer a api composite of encoder and decoder
type Serializer interface {
	WriteTo(w io.Writer, object interface{}) error
	Write(object interface{}) error
	ToBytes(interface{}) ([]byte, error)

	ReadFrom(reader ByteRuneReader) (interface{}, error)
	Read() (interface{}, error)
	ToObject([]byte) (interface{}, error)
}

// goHessian the serializer cache struct, a composite of encoder and decoder
type goHessian struct {
	encoder *Encoder
	decoder *Decoder
}

//NewSerializer init
func NewSerializer(typMap map[string]reflect.Type, nameMap map[string]string) Serializer {
	return &goHessian{
		encoder: NewEncoder(nil, nameMap),
		decoder: NewDecoder(nil, typMap),
	}
}

// WriteObject to writer
func (gh *goHessian) WriteTo(w io.Writer, object interface{}) error {
	return gh.encoder.WriteTo(w, object)
}

// Write object to writer continuously , it must be called after calling goHessian.WriteObject
func (gh *goHessian) Write(object interface{}) error {
	return gh.encoder.WriteObject(object)
}

// ToBytes convert object to bytes
func (gh *goHessian) ToBytes(object interface{}) ([]byte, error) {
	return gh.encoder.ToBytes(object)
}

// ReadObject from reader
func (gh *goHessian) ReadFrom(reader ByteRuneReader) (interface{}, error) {
	return gh.decoder.ReadFrom(reader)
}

// Read from reader continuously, it must be called after calling goHessian.ReadObject
func (gh *goHessian) Read() (interface{}, error) {
	return gh.decoder.ReadData()
}

// ToObject convert bytes to object
func (gh *goHessian) ToObject(bts []byte) (interface{}, error) {
	return gh.decoder.ToObject(bts)
}

// ---------------------------------------------

//ToBytes [NO-CACHE API] serialize object to bytes
func ToBytes(object interface{}, nameMap map[string]string) ([]byte, error) {
	e := NewEncoder(nil, nameMap)
	return e.ToBytes(object)
}

//ToObject [NO-CACHE API] deserialize bytes to object
func ToObject(ins []byte, typMap map[string]reflect.Type) (interface{}, error) {
	d := NewDecoder(nil, typMap)
	return d.ToObject(ins)
}
