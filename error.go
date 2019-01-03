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
	"fmt"
	"path/filepath"
	"runtime"
)

// ErrDecoder is returned when the encoder encounters an error.
type ErrDecoder struct {
	Message string
	Err     error
}

func (e ErrDecoder) Error() string {
	if e.Err == nil {
		return "cannot decode " + e.Message
	}
	return "cannot decode " + e.Message + ": " + e.Err.Error()
}

func newCodecError(dataType string, a ...interface{}) *ErrDecoder {
	var err error
	var format, message string
	var ok bool

	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "???"
		line = 0
	}
	file = filepath.Base(file)
	caller := fmt.Sprintf("(%s:%d)", file, line)

	if len(a) == 0 {
		return &ErrDecoder{dataType + ": no reason given" + caller, nil}
	}
	// if last item is error: save it
	if err, ok = a[len(a)-1].(error); ok {
		a = a[:len(a)-1] // pop it
	}
	// if items left, first ought to be format string
	if len(a) > 0 {
		if format, ok = a[0].(string); ok {
			a = a[1:] // unshift
			message = fmt.Sprintf(format, a...)
		}
	}
	if message != "" {
		message = ": " + message
	}
	return &ErrDecoder{dataType + message + caller, err}
}
