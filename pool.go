// Copyright 2018-2019 vogo.
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

import "reflect"

type Pool interface {
	Get() interface{}
	Return(interface{})
}

type objectFactory func() interface{}

type objectPool struct {
	factory objectFactory
	cached  chan interface{}
}

func newPool(size int, f objectFactory) Pool {
	return &objectPool{
		factory: f,
		cached:  make(chan interface{}, size),
	}
}

func (p *objectPool) Get() interface{} {
	var o interface{}
	select {
	case o = <-p.cached:
		return o
	default:
		return p.factory()
	}
}

func (p *objectPool) Return(o interface{}) {
	select {
	case p.cached <- o: // Try to put back into the pool
	default:
		// Pool is full, will be garbage collected
	}
}

//NewEncoderPool new pool for encoder
func NewEncoderPool(size int, nameMap map[string]string) Pool {
	return newPool(size, func() interface{} {
		return NewEncoder(nil, nameMap)
	})
}

//NewDecoderPool new pool for decoder
func NewDecoderPool(size int, typeMap map[string]reflect.Type) Pool {
	return newPool(size, func() interface{} {
		return NewDecoder(nil, typeMap)
	})
}

//NewSerializerPool new pool for serializer
func NewSerializerPool(size int, typeMap map[string]reflect.Type, nameMap map[string]string) Pool {
	return newPool(size, func() interface{} {
		return NewSerializer(typeMap, nameMap)
	})
}
