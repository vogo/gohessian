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

package tests

import (
	"encoding/base64"
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vogo/gohessian"
)

//TConfig define
type TConfig struct {
	Enable bool
	Msg    string
	Flag   int
}

//HessianCodecName java class name
func (TConfig) HessianCodecName() string {
	return "test.TConfig"
}

//TConfigMap define
type TConfigMap map[string]*TConfig

//HessianCodecName java class name
func (TConfigMap) HessianCodecName() string {
	return "java.util.concurrent.ConcurrentHashMap"
}

//decodeTConfigMap from hessian encode bytes
func decodeTConfigMap(t *testing.T, data []byte, typMap map[string]reflect.Type) (cfg TConfigMap, err error) {
	if data == nil || len(data) == 0 {
		return nil, errors.New("nil byte")
	}
	res, err := hessian.ToObject(data, typMap)
	if err != nil {
		t.Errorf("failed decode config map bytes: %v, %v\n", base64.StdEncoding.EncodeToString(data), err)
		return nil, err
	}

	if sn, ok := res.(map[interface{}]interface{}); ok && len(sn) == 0 {
		return TConfigMap{}, nil
	}

	t.Log("decoded: ", res)
	if sn, ok := res.(TConfigMap); ok {
		cfg = sn
		return
	}
	t.Errorf("unexpect decode config map result: %v, type:%v, base64:%v\n", res, reflect.TypeOf(res), base64.StdEncoding.EncodeToString(data))
	err = errors.New("failed to decode config map")
	return
}

//encodeTConfigMap to bytes
func encodeTConfigMap(t *testing.T, cfg TConfigMap, nameMap map[string]string) ([]byte, error) {
	return hessian.ToBytes(cfg, nameMap)
}

func TestUntypedMap(t *testing.T) {

	m := make(map[interface{}]interface{})
	m["test"] = "test"
	m["test2"] = 1
	m[1] = "test"
	m[2] = 2

	bytes, err := hessian.ToBytes(m, nil)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("base64: %s", string(bytes))

	result, err := hessian.ToObject(bytes, nil)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(result)
	rmap, ok := result.(map[interface{}]interface{})
	if !ok {
		t.Error("can't convert map")
		return
	}
	assert.Equal(t, len(m), len(rmap))
}

func TestEncodeDecodeMapType(t *testing.T) {
	typeMap, nameMap := hessian.ExtractTypeNameMap(TConfigMap{})

	t.Log(typeMap)
	t.Log(nameMap)

	tMap := make(TConfigMap)
	tMap["200101"] = &TConfig{Enable: true, Msg: "test1", Flag: 999}
	tMap["200102"] = &TConfig{Enable: false, Msg: "test2", Flag: -999}

	t.Log("config map type:", reflect.TypeOf(tMap))
	bytes, err := encodeTConfigMap(t, tMap, nameMap)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("base64: %s\n", string(bytes))

	cfg, err := decodeTConfigMap(t, bytes, typeMap)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(cfg)

	assert.Equal(t, 2, len(cfg))

	c1, ok := cfg["200101"]
	assert.True(t, ok)
	assert.Equal(t, true, c1.Enable)
	assert.Equal(t, "test1", c1.Msg)
	assert.Equal(t, 999, c1.Flag)

	c2, ok := cfg["200102"]
	assert.True(t, ok)
	assert.Equal(t, false, c2.Enable)
	assert.Equal(t, "test2", c2.Msg)
	assert.Equal(t, -999, c2.Flag)
}
