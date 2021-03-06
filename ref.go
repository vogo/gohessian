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
// For example, a circular linked-list will refer to the first link before the entire value has been read.
//
// A possible implementation would add each map, list, and object to an array as it is read.
// The ref will return the corresponding value from the array.
// To support circular structures, the implementation would store the map, list or object immediately, before filling in the contents.
//
// Each map or value is stored into an array as it is parsed. ref selects one of the stored objects. The first object is numbered '0'.
//
//
// -----------> Ref Examples
//
// Circular list
//
// list = new LinkedList();
// list.data = 1;
// list.tail = value;
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
	"unsafe"
)

const (
	_refStartTag = 0x51
)

// used to ref object,list,map
type _refElem struct {
	// record the kind of target, objects are the same only if the address and kind are the same
	kind reflect.Kind

	// ref index
	index int
}

func refTag(tag byte) bool {
	return tag == _refStartTag
}

func (e *Encoder) writeRef(index int) (int, error) {
	e.writeBT(_refStartTag)
	return e.writer.Write(encodeInt(int32(index)))
}

// return the order number of ref object if found ,
// otherwise, add the object into the encode ref map
func (e *Encoder) checkEncodeRefMap(v reflect.Value) (int, bool) {
	var (
		kind reflect.Kind
		addr unsafe.Pointer
	)

	if v.Kind() == reflect.Ptr {
		for v.Elem().Kind() == reflect.Ptr {
			v = v.Elem()
		}
		kind = v.Elem().Kind()
		if kind == reflect.Slice || kind == reflect.Map {
			addr = unsafe.Pointer(v.Elem().Pointer())
		} else {
			addr = unsafe.Pointer(v.Pointer())
		}
	} else {
		kind = v.Kind()
		switch kind {
		case reflect.Slice, reflect.Map:
			addr = unsafe.Pointer(v.Pointer())
		default:
			addr = unsafe.Pointer(PackPtr(v).Pointer())
		}
	}

	if elem, ok := e.refMap[addr]; ok {
		// the array addr is equal to the first elem, which must ignore
		if elem.kind == kind {
			// fmt.Printf("-----> find ref: %d, %p, %v, %v\n", elem.index, addr, kind, v)
			return elem.index, ok
		}
		return 0, false
	}

	n := len(e.refMap)
	e.refMap[addr] = _refElem{kind, n}
	// fmt.Printf("---> add ref: %d, %p, %v, %v\n", n, addr, kind, v)
	return 0, false
}

type _refHolder struct {
	// destinations
	destinations []reflect.Value

	value reflect.Value
}

var _refHolderType = reflect.TypeOf(_refHolder{})

// change ref value
func (h *_refHolder) change(v reflect.Value) {
	if h.value.CanAddr() && v.CanAddr() && h.value.Pointer() == v.Pointer() {
		return
	}
	h.value = v
}

// notice all destinations ref to the value
func (h *_refHolder) notify() {
	for _, dest := range h.destinations {
		SetValue(dest, h.value)
	}
}

// add destination
func (h *_refHolder) add(dest reflect.Value) {
	h.destinations = append(h.destinations, dest)
	// SetValue(dest, h.value)
}

func (d *Decoder) addDecoderRef(v reflect.Value) *_refHolder {
	// fmt.Printf("--> addDecoderRef: %d, %p, %v, %v\n", len(d.refList), v.Interface(), v.Type(), v.Interface())
	var holder *_refHolder
	// only slice and array need ref holder , for its address changes when decoding
	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		holder = &_refHolder{
			value: v,
		}
		v = reflect.ValueOf(holder)
	}

	d.refList = append(d.refList, v)
	return holder
}

// read the ref reflect.Value , which may be one of type _refHolder
func (d *Decoder) readRef(tag byte) (reflect.Value, error) {
	if tag != _refStartTag {
		return _zeroValue, newCodecError("readRef", "error ref tag: 0x%x", tag)
	}
	index, err := d.readInt(_tagRead)
	if err != nil {
		return _zeroValue, err
	}
	idx := int(index)
	if len(d.refList) <= idx {
		return _zeroValue, newCodecError("readRef", "ref index out of bound, max %d, but got %d", len(d.refList), index)
	}

	ref := d.refList[idx]

	// fmt.Printf("----> readRef: %d, %p, %v, %v\n", idx, unsafe.Pointer(ref.Pointer()), ref.Elem().Kind(), ref.Interface())
	return ref, nil
}
