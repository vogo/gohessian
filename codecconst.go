/*
 *
 *  * Copyright 2012-2016 Viant.
 *  *
 *  * Licensed under the Apache License, Version 2.0 (the "License"); you may not
 *  * use this file except in compliance with the License. You may obtain a copy of
 *  * the License at
 *  *
 *  * http://www.apache.org/licenses/LICENSE-2.0
 *  *
 *  * Unless required by applicable law or agreed to in writing, software
 *  * distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
 *  * WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
 *  * License for the specific language governing permissions and limitations under
 *  * the License.
 *
 */

package hessian

import "reflect"

const (
	mask        = byte(127)
	flag        = byte(128)
	_tagRead    = int32(-1)
	_asciiGap   = 32
	BcDate      = byte(0x4a) // 64-bit millisecond UTC date
	_dateMinute = byte(0x4b) // 32-bit minute UTC date
	_endFlag    = byte('Z')
	_nilTag     = byte('N')
	BcRef       = byte(0x51)

	PPacketChunk    = byte(0x4f)
	PPacket         = byte('P')
	PPacketDirect   = byte(0x80)
	PacketDirectMax = byte(0x7f)
	PPacketShort    = byte(0x70)
	PacketShortMax  = 0xfff

	_interfaceTypeName = "interface {}"
)

var (
	_buildInTypeNameMap = make(map[string]string)
)

func addBuildInNameType(i interface{}, convertName string) {
	typ := reflect.TypeOf(i)
	name := typ.Name()
	if name == "" {
		panic("type name is nil for type " + typ.String())
	}
	if convertName == "" {
		convertName = name
	}
	_buildInTypeNameMap[name] = convertName
	_buildInTypeNameMap[convertName] = convertName
}

func init() {
	addBuildInNameType(" ", "string")

	addBuildInNameType(int(1), "int")
	addBuildInNameType(int8(1), "int")
	addBuildInNameType(int16(1), "int")
	addBuildInNameType(int32(1), "int")

	// java: long
	addBuildInNameType(int64(1), "long")

	// java: float
	addBuildInNameType(float32(1.0), "float")

	// java: double
	addBuildInNameType(float64(1.0), "double")

	// java: boolean
	addBuildInNameType(true, "boolean")
}
