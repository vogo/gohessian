package htests

import (
	"encoding/base64"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/luckincoffee/gohessian"
)

var hessianTypeMap map[string]reflect.Type
var hessianNameMap map[string]string

// ServerAPI server api
type ServerApi struct {
	ApiName string `json:"apiName"`
	ApiDesc string `json:"apiDesc"`
	AppRoot string `json:"appRoot"`
}

//ServerNode server node
type ServerNode struct {
	Name     string      `json:"name"`
	Version  string      `json:"version"`
	Desc     string      `json:"desc"`
	Address  string      `json:"address"`
	Channels []string    `json:"channels"`
	ApiList  []ServerApi `json:"apiList"`
}

//JavaClassName for ServerApi
func (ServerApi) JavaClassName() string {
	return "test.ServerApi"
}

//JavaClassName for ServerNode
func (ServerNode) JavaClassName() string {
	return "test.ServerNode"
}

func init() {
	node := ServerNode{}
	api := ServerApi{}
	serverNodeType := reflect.TypeOf(node)
	serverApiType := reflect.TypeOf(api)

	hessianTypeMap = hessian.TypeMapOf(serverNodeType)
	hessianTypeMap[node.JavaClassName()] = serverNodeType
	hessianTypeMap[api.JavaClassName()] = serverApiType

	hessianNameMap = make(map[string]string)
	hessianNameMap[serverNodeType.Name()] = node.JavaClassName()
	hessianNameMap[serverApiType.Name()] = api.JavaClassName()
}

//DecodeServerNode from hessian encode bytes
func DecodeServerNode(data []byte) (node *ServerNode, err error) {
	if data == nil || len(data) == 0 {
		return nil, errors.New("nil byte")
	}
	res, err := hessian.ToObject(data, hessianTypeMap)
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
func EncodeServerNode(node *ServerNode) ([]byte, error) {
	return hessian.ToBytes(*node, hessianNameMap)
}

func TestHessianEncodeDecode(t *testing.T) {
	node := &ServerNode{
		Version: "v1",
		Name:    "api",
		//Desc:     "dd",
		Address:  "127.0.0.1",
		Channels: []string{"c1", "c2"},
		ApiList: []ServerApi{
			ServerApi{AppRoot: "/user", ApiName: "user", ApiDesc: "user"},
			ServerApi{AppRoot: "/list", ApiName: "list", ApiDesc: "list"},
		},
	}

	bt, err := EncodeServerNode(node)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	t.Log(base64.StdEncoding.EncodeToString(bt))

	sn, err := DecodeServerNode(bt)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	t.Log(sn)
}
