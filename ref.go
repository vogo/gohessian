// Copyright 2019 vogo.
// Author: wongoo
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// -----------> Ref Grammar
//
// ref ::= x51 int
//
// An integer referring to a previous list, map, or object instance.
// As each list, map or object is read from the input stream, it is assigned the integer position in the stream,
// i.e. the first list or map is '0', the next is '1', etc.
// A later ref can then use the previous object.
// Writers MAY generate refs. Parsers MUST be able to recognize them.
//
// ref can refer to incompletely-read items.
// For example, a circularT linked-list will refer to the first link before the entire list has been read.
//
// A possible implementation would add each map, list, and object to an array as it is read.
// The ref will return the corresponding value from the array.
// To support circularT structures, the implementation would store the map, list or object immediately, before filling in the contents.
//
// Each map or list is stored into an array as it is parsed. ref selects one of the stored objects. The first object is numbered '0'.
//
//
// -----------> Ref Examples
//
// Circular list
//
// list = new LinkedList();
// list.data = 1;
// list.tail = list;
//
// ---
// C
//   x0a LinkedList
//   x92
//   x04 head
//   x04 tail
//
// o x90      # object stores ref #0
//   x91      # data = 1
//   x51 x90  # next field refers to itself, i.e. ref #0
//
// ref only refers to list, map and objects elements.
// Strings and binary data, in particular, will only share references if they're wrapped in a list or map.

package hessian

import (
	"reflect"
)

const (
	refStartTag = 0x51
)

func refTag(tag byte) bool {
	return tag == refStartTag
}

func (e *Encoder) writeRef(index int) (int, error) {
	e.writeBT(refStartTag)
	return e.writer.Write(encodeInt(int32(index)))
}

// return the order number of ref object if found ,
// otherwise, add the object into the encode ref map
func (e *Encoder) checkEncodeRefMap(v reflect.Value) (int, bool) {
	if v.Kind() == reflect.Ptr {
		for v.Elem().Kind() == reflect.Ptr {
			v = v.Elem()
		}
	} else {
		// pack the raw value with a pointer value in order to get the pointer address
		v = PackPtr(v)
	}

	// check whether to ref other object
	addr := v.Pointer()
	if n, ok := e.refMap[addr]; ok {
		return n, ok
	}

	n := len(e.refMap)
	e.refMap[v.Pointer()] = n
	return 0, false
}

func (d *Decoder) addDecoderRef(v reflect.Value) {
	// fmt.Printf("addDecoderRef: %v, %v, %p\n", v.Type(), v.Interface(), v.Interface())
	d.refList = append(d.refList, v)
}

func (d *Decoder) readRef(tag byte) (interface{}, error) {
	if tag != refStartTag {
		return nil, newCodecError("readRef", "should be ref tag: %x, but got %x", tag, refStartTag)
	}
	index, err := d.readInt(TagRead)
	if err != nil {
		return nil, err
	}
	idx := int(index)
	if len(d.refList) <= idx {
		return nil, newCodecError("readRef", "ref index out of bound, max %d, but got %d", len(d.refList), index)
	}
	return d.refList[idx].Interface(), nil
}
