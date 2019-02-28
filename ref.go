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
// For example, a circular linked-list will refer to the first link before the entire list has been read.
//
// A possible implementation would add each map, list, and object to an array as it is read.
// The ref will return the corresponding value from the array.
// To support circular structures, the implementation would store the map, list or object immediately, before filling in the contents.
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
