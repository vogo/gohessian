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

const (
	mask         = byte(127)
	flag         = byte(128)
	TagRead      = int32(-1)
	AsciiGap     = 32
	BcDate       = byte(0x4a) // 64-bit millisecond UTC date
	BcDateMinute = byte(0x4b) // 32-bit minute UTC date
	EndFlag      = byte('Z')
	BcNull       = byte('N')
	BcRef        = byte(0x51)

	PPacketChunk    = byte(0x4f)
	PPacket         = byte('P')
	PPacketDirect   = byte(0x80)
	PacketDirectMax = byte(0x7f)
	PPacketShort    = byte(0x70)
	PacketShortMax  = 0xfff

	ArrayString = "[string"
	ArrayInt    = "[int"
	ArrayDouble = "[double"
	ArrayFloat  = "[float"
	ArrayBool   = "[boolean"
	ArrayLong   = "[long"
)
