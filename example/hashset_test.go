package hessian_test

import (
	"encoding/base64"
	"reflect"
	"testing"

	hessian "github.com/luckincoffee/gohessian"
	"github.com/stretchr/testify/assert"
)

func TestHashSet(t *testing.T) {
	var hashsetJavaClassName = "java.util.HashSet"
	var hessianHashsetBase64 = "chFqYXZhLnV0aWwuSGFzaFNldAZjY2NkZGQGYWFhYmJi"

	data, err := base64.StdEncoding.DecodeString(hessianHashsetBase64)
	if err != nil {
		t.Error(err)
	}

	var hashsetType reflect.Type
	var hashsetHessianTypeMap map[string]reflect.Type
	var hashsetHessianNameMap map[string]string

	hashset := []string{}
	hashsetType = reflect.TypeOf(hashset)

	hashsetHessianTypeMap = hessian.TypeMapOf(hashsetType)
	hashsetHessianTypeMap[hashsetJavaClassName] = hashsetType

	hashsetHessianNameMap = make(map[string]string)
	hashsetHessianNameMap[hashsetType.Name()] = hashsetJavaClassName

	obj, err := hessian.ToObject(data, hashsetHessianTypeMap)
	if err != nil {
		t.Error(err)
	}

	t.Log(obj)

	arr, ok := obj.([]interface{})
	if !ok {
		t.Error("result not []interface{}")
	}

	t.Logf("arr length:%d", len(arr))
	t.Logf("arr:%v", arr)
	assert.Equal(t, 2, len(arr))
	assert.Equal(t, "aaabbb", arr[1])
	assert.Equal(t, "cccddd", arr[0])

	sarr := make([]string, len(arr))
	for i, s := range arr {
		sarr[i] = s.(string)
	}
	t.Logf("sarr:%v", sarr)
}
