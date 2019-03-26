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

package hessian

import (
	"encoding/base64"
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// serverApiT server api
type serverApiT struct {
	ApiName string `json:"apiName"`
	ApiDesc string `json:"apiDesc"`
	AppRoot string `json:"appRoot"`
}

//serverNodeT server serverNode
type serverNodeT struct {
	Name     string       `json:"name"`
	Version  string       `json:"version"`
	Desc     string       `json:"desc"`
	Address  string       `json:"address"`
	Channels []string     `json:"channels"`
	ApiList  []serverApiT `json:"apiList"`
}

//HessianCodecName for serverApiT
func (serverApiT) HessianCodecName() string {
	return "test.serverApiT"
}

//HessianCodecName for serverNodeT
func (serverNodeT) HessianCodecName() string {
	return "test.serverNodeT"
}

//testDecodeServerNode from hessian encode bytes
func testDecodeServerNode(t *testing.T, data []byte, typMap map[string]reflect.Type) (node *serverNodeT, err error) {
	if data == nil || len(data) == 0 {
		return nil, errors.New("nil byte")
	}
	res, err := ToObject(data, typMap)
	if err != nil {
		t.Log("failed decode bytes:", base64.StdEncoding.EncodeToString(data))
		return nil, err
	}
	if sn, ok := res.(*serverNodeT); ok {
		node = sn
		return
	}
	err = errors.New("failed to decode serverNodeT")
	return
}

//testEncodeServerNode to bytes
func testEncodeServerNode(t *testing.T, node *serverNodeT, nameMap map[string]string) ([]byte, error) {
	return ToBytes(*node, nameMap)
}

func TestHessianEncodeDecode(t *testing.T) {
	typeMap, nameMap := ExtractTypeNameMap(serverNodeT{})
	t.Log(typeMap)
	t.Log(nameMap)

	node := &serverNodeT{
		Version: "v1",
		Name:    "api",
		//Desc:     "dd",
		Address:  "127.0.0.1",
		Channels: []string{"c1", "c2"},
		ApiList: []serverApiT{
			{AppRoot: "/user", ApiName: "user", ApiDesc: "user"},
			{AppRoot: "/list", ApiName: "list", ApiDesc: "list"},
		},
	}

	bt, err := testEncodeServerNode(t, node, nameMap)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	t.Log(base64.StdEncoding.EncodeToString(bt))

	decodeNode, err := testDecodeServerNode(t, bt, typeMap)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	assert.NotNil(t, decodeNode)
	t.Log(decodeNode)
	assert.Equal(t, node.Version, decodeNode.Version)
	assert.Equal(t, node.Name, decodeNode.Name)
	assert.Equal(t, node.Address, decodeNode.Address)
	assert.Equal(t, 2, len(decodeNode.Channels))
	assert.Equal(t, node.Channels[0], decodeNode.Channels[0])
	assert.Equal(t, node.Channels[1], decodeNode.Channels[1])
	assert.Equal(t, len(node.ApiList), len(decodeNode.ApiList))
	assert.Equal(t, node.ApiList[0].AppRoot, decodeNode.ApiList[0].AppRoot)
	assert.Equal(t, node.ApiList[0].ApiDesc, decodeNode.ApiList[0].ApiDesc)
	assert.Equal(t, node.ApiList[0].ApiName, decodeNode.ApiList[0].ApiName)
}
