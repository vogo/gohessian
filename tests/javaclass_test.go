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
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vogo/gohessian"
)

// ServerAPI server api
type ServerApi struct {
	ApiName string `json:"apiName"`
	ApiDesc string `json:"apiDesc"`
	AppRoot string `json:"appRoot"`
}

//ServerNode server serverNode
type ServerNode struct {
	Name     string      `json:"name"`
	Version  string      `json:"version"`
	Desc     string      `json:"desc"`
	Address  string      `json:"address"`
	Channels []string    `json:"channels"`
	ApiList  []ServerApi `json:"apiList"`
}

//HessianCodecName for ServerApi
func (ServerApi) HessianCodecName() string {
	return "test.ServerApi"
}

//HessianCodecName for ServerNode
func (ServerNode) HessianCodecName() string {
	return "test.ServerNode"
}

//DecodeServerNode from hessian encode bytes
func DecodeServerNode(data []byte, typMap map[string]reflect.Type) (node *ServerNode, err error) {
	if data == nil || len(data) == 0 {
		return nil, errors.New("nil byte")
	}
	res, err := hessian.ToObject(data, typMap)
	if err != nil {
		fmt.Println("failed decode bytes:", base64.StdEncoding.EncodeToString(data))
		return nil, err
	}
	if sn, ok := res.(*ServerNode); ok {
		node = sn
		return
	}
	err = errors.New("failed to decode ServerNode")
	return
}

//EncodeServerNode to bytes
func EncodeServerNode(node *ServerNode, nameMap map[string]string) ([]byte, error) {
	return hessian.ToBytes(*node, nameMap)
}

func TestHessianEncodeDecode(t *testing.T) {
	typeMap, nameMap := hessian.ExtractTypeNameMap(ServerNode{})
	fmt.Println(typeMap)
	fmt.Println(nameMap)

	node := &ServerNode{
		Version: "v1",
		Name:    "api",
		//Desc:     "dd",
		Address:  "127.0.0.1",
		Channels: []string{"c1", "c2"},
		ApiList: []ServerApi{
			{AppRoot: "/user", ApiName: "user", ApiDesc: "user"},
			{AppRoot: "/list", ApiName: "list", ApiDesc: "list"},
		},
	}

	bt, err := EncodeServerNode(node, nameMap)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	t.Log(base64.StdEncoding.EncodeToString(bt))

	decodeNode, err := DecodeServerNode(bt, typeMap)
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
