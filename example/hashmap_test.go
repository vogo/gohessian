// authors: wangoo
// created: 2018-12-13
// test java hashmap decode

package hessian

import (
	"encoding/base64"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	hessian "github.com/luckincoffee/gohessian"
)

type mcidConfig struct {
	Enable bool
	Msg    string
}

func (mcidConfig) JavaClassName() string {
	return "test.McidConfig"
}

type mcidConfigMap map[string]*mcidConfig

func (mcidConfigMap) JavaClassName() string {
	return "java.util.concurrent.ConcurrentHashMap"
}

func TestHashMap(t *testing.T) {
	doTestDecodeHashMap(t, 0, "SFo=")
	doTestDecodeHashMap(t, 1, "TTAmamF2YS51dGlsLmNvbmN1cnJlbnQuQ29uY3VycmVudEhhc2hNYXADMTIzQw90ZXN0Lk1jaWRDb25maWeSBmVuYWJsZQNtc2dgVAR0ZXN0Wg==")
	doTestDecodeHashMap(t, 2, "TTAmamF2YS51dGlsLmNvbmN1cnJlbnQuQ29uY3VycmVudEhhc2hNYXADMTIzQw90ZXN0Lk1jaWRDb25maWeSBmVuYWJsZQNtc2dgVAV0ZXN0MQM0NTZgRgV0ZXN0Mlo=")
}

func doTestDecodeHashMap(t *testing.T, size int, base64Str string) {
	data, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		t.Error(err)
	}

	t.Logf("data: %v", string(data))

	var mcidHessianTypeMap map[string]reflect.Type
	var mcidHessianNameMap map[string]string
	var mcidCfgType reflect.Type
	var mcidMapType reflect.Type

	cfg := mcidConfig{}
	group := mcidConfigMap{}
	mcidCfgType = reflect.TypeOf(cfg)
	mcidMapType = reflect.TypeOf(group)

	mcidHessianTypeMap = hessian.TypeMapOf(mcidCfgType)
	mcidHessianTypeMap[cfg.JavaClassName()] = mcidCfgType
	mcidHessianTypeMap[group.JavaClassName()] = mcidMapType

	mcidHessianNameMap = make(map[string]string)
	mcidHessianNameMap[mcidCfgType.Name()] = cfg.JavaClassName()
	mcidHessianNameMap[mcidMapType.Name()] = group.JavaClassName()

	res, err := hessian.ToObject(data, mcidHessianTypeMap)
	t.Logf("res: %v, %v", reflect.TypeOf(res), res)

	if err != nil {
		t.Errorf("failed decode monitor cid config group bytes: %v", base64.StdEncoding.EncodeToString(data))
	}
	if untypedMap, ok := res.(map[interface{}]interface{}); ok {
		t.Logf("untyped map: %v", untypedMap)
		assert.Equal(t, size, len(untypedMap))
		return
	}
	if mcidMap, ok := res.(map[string]*mcidConfig); ok {
		t.Logf("mcid map: %v", mcidMap)
		assert.Equal(t, size, len(mcidMap))
		return
	}

	t.Error("unknown map type")
}
