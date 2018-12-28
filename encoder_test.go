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
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/stretchr/testify/assert"
	"reflect"
	"strings"
	"testing"
)

type P struct {
	X, Y, Z int
	Name    string
}

type Box struct {
	Width  int
	Height int
	Color  string
	Open   bool
}

type BB struct {
	Name   string
	List   []string
	Mp     map[int32]string
	Pst    P
	Number int32
}
type IntS struct {
	V   int
	V8  int8
	V16 int16
	V32 int32
	V64 int64
	U   uint
	U8  uint8
	U16 uint16
	U32 uint32
	U64 uint64
}

func TestIntConvert(t *testing.T) {
	var V int = 1
	var V8 int8 = 8
	var V16 int16 = 16
	var V32 int32 = 32
	var V64 int64 = 64
	var U uint = 128
	var U8 uint8 = 9
	var U16 uint16 = 17
	var U32 uint32 = 33
	var U64 uint64 = 256

	assert.False(t, conv32(V))
	assert.False(t, conv32(V8))
	assert.False(t, conv32(V16))
	assert.True(t, conv32(V32))
	assert.True(t, conv64(V64))
	assert.False(t, conv64(U))
	assert.False(t, conv64(U8))
	assert.False(t, conv64(U16))
	assert.False(t, conv64(U32))
	assert.False(t, conv64(U64))
}

func conv32(i interface{}) bool {
	x, ok := i.(int32)
	fmt.Print(x)
	return ok
}
func conv64(i interface{}) bool {
	x, ok := i.(int64)
	fmt.Print(x)
	return ok
}

func TestSerializer(t *testing.T) {
	ts4 := []string{"t1", "t2", "t3"}
	mp1 := map[int32]string{1: "test1", 2: "test2", 3: "test3"}
	p := P{1, 2, 3, "ABC"}
	bb := BB{"AB", ts4, mp1, p, 4}
	gh := NewGoHessian(nil, nil)
	bt, err := gh.ToBytes(bb)
	if err != nil {
		t.Error("serializer error:", err)
		return
	}
	t.Log("bt", string(bt))

}

func TestEncodeDecodeInt(t *testing.T) {
	s := IntS{
		V:   1,
		V8:  2,
		V16: 3,
		V32: 4,
		V64: 5,
		U:   6,
		U8:  7,
		U16: 8,
		U32: 9,
		U64: 10,
	}

	bytes, err := Encode(s)
	if err != nil {
		t.Error("encode int struct error:", err)
		return
	}
	t.Log("int struct:", string(bytes))
	t.Log("int struct HEX:", hex.EncodeToString(bytes))

	s1, err := Decode(bytes, reflect.TypeOf(s))
	if err != nil {
		t.Error("decode int struct error:", err)
		return
	}
	assert.True(t, reflect.DeepEqual(&s, s1))
}

func TestDecoder_Instance(t *testing.T) {
	ts4 := []string{"t1", "t2", "t3"}
	mp1 := map[int32]string{1: "test1", 2: "test2", 3: "test3"}
	p := P{1, 2, 3, "ABC"}
	bb := BB{"AB", ts4, mp1, p, 4}
	br := bytes.NewBuffer(nil)
	e := NewEncoder(br, nil)
	e.WriteObject(bb)
	t.Log("bytes:", br)

	bt := bytes.NewReader(br.Bytes())
	d := NewDecoder(bt, nil)
	d.RegisterType("BB", reflect.TypeOf(BB{}))
	d.RegisterType("P", reflect.TypeOf(P{}))
	it, err := d.ReadObject()
	if err != nil {
		t.Error(err)
	}
	t.Log("decode t", it, "bt len", len(br.Bytes()))
}

func TestEncoder_WriteObject(t *testing.T) {
	mp2 := make(map[int]string)
	mp2[1] = "test1"
	mp2[2] = "test2"
	mp2[3] = "test3"
	br := bytes.NewBuffer(nil)
	e7 := NewEncoder(br, nil)
	e7.WriteObject(mp2)
	t.Log("encode map buf->", string(br.Bytes()), len(br.Bytes()), br.Bytes())
	bt2 := br.Bytes()
	br2 := bytes.NewReader(br.Bytes())
	d7 := NewDecoder(br2, nil)
	t7, err := d7.ReadObject()
	if err != nil {
		t.Error("read object error:", err)
		return
	}
	t.Log("decode map", t7, "bt2", len(bt2))

}

func TestEncoder_WriteList(t *testing.T) {
	ts3 := [3]string{"t1", "t2", "t3"}
	br := bytes.NewBuffer(nil)
	e7 := NewEncoder(br, nil)
	_, err := e7.WriteObject(ts3)
	if err != nil {
		t.Error("write object error:", err)
		return
	}
	t.Log("encode array buf->", string(br.Bytes()), len(br.Bytes()), br.Bytes())
	bt2 := br.Bytes()
	br2 := bytes.NewReader(br.Bytes())
	d7 := NewDecoder(br2, nil)
	t7, err := d7.ReadObject()
	if err != nil {
		t.Error("read object error:", err)
		return
	}
	t.Log("decode array", t7, "bt2", len(bt2))

}

func TestEncoder_WriteString(t *testing.T) {
	str := "HessianSerializer"
	br := bytes.NewBuffer(nil)
	e7 := NewEncoder(br, nil)
	_, err := e7.WriteObject(str)
	if err != nil {
		t.Error("write object error:", err)
		return
	}
	bt2 := br.Bytes()
	br2 := bytes.NewReader(br.Bytes())
	d7 := NewDecoder(br2, nil)
	t7, err := d7.ReadObject()
	if err != nil {
		t.Error("read object error:", err)
		return
	}
	t.Log("decode string", t7, "bt2", len(bt2))
	if strings.Compare(str, t7.(string)) == 0 {
		t.Log("succes for ", str)
	}
}
