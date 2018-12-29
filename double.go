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

import "math"

const (
	BcDouble      = byte('D') // IEEE 64-bit double
	BcDoubleZero  = byte(0x5b)
	BcDoubleOne   = byte(0x5c)
	BcDoubleByte  = byte(0x5d)
	BcDoubleShort = byte(0x5e)
	BcDoubleMill  = byte(0x5f)
)

// see: http://hessian.caucho.com/doc/hessian-serialization.html##double
func encodeDouble(value float64) ([]byte, error) {
	v := float64(int64(value))
	if v == value {
		iv := int64(value)
		if iv == 0 {
			return []byte{BcDoubleZero}, nil
		}
		if iv == 1 {
			return []byte{BcDoubleOne}, nil
		}

		if iv >= -0x80 && iv < 0x80 {
			return []byte{BcDoubleByte, byte(iv)}, nil
		}

		if iv >= -0x8000 && iv < 0x8000 {
			return []byte{BcDoubleByte, byte(iv >> 8), byte(iv)}, nil
		}
		return nil, newCodecError("encodeDouble", "unsupported double range ", iv)
	}

	bits := uint64(math.Float64bits(value))
	return []byte{BcDouble,
		byte(bits >> 56),
		byte(bits >> 48),
		byte(bits >> 40),
		byte(bits >> 32),
		byte(bits >> 24),
		byte(bits >> 16),
		byte(bits >> 8),
		byte(bits)}, nil
}
