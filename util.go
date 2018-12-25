// Copyright 2012-2016 luckin coffee.
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

func strTag(tag byte) bool {
	return (tag >= BC_STRING_DIRECT && tag <= STRING_DIRECT_MAX) || (tag >= 0x30 && tag <= 0x34) || (tag == BC_STRING || tag == BC_STRING_CHUNK)
}
