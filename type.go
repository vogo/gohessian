// Copyright 2018 luckin coffee.
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

func (d *Decoder) readType() (string, error) {
	tag, err := readTag(d.reader)
	if err != nil {
		return "", newCodecError("reading tag", err)
	}
	if stringTag(tag) {
		t, err := d.readString(int32(tag))
		if err != nil {
			return "", newCodecError("reading tag", err)
		}
		d.typList = append(d.typList, t)
		return t, nil
	}
	i, err := d.readInt(TagRead)
	if err != nil {
		return "", newCodecError("reading tag", err)
	}
	index := int(i)
	return d.typList[index], nil
}
