/*

 go-float16 - IEEE 754 binary16 half precision format
 Written in 2013 by h2so5 <mail@h2so5.net>

 To the extent possible under law, the author(s) have dedicated all copyright and
 related and neighboring rights to this software to the public domain worldwide.
 This software is distributed without any warranty.
 You should have received a copy of the CC0 Public Domain Dedication along with this software.
 If not, see <http://creativecommons.org/publicdomain/zero/1.0/>.

*/

package half

import (
	"math"
	"testing"
)

func getFloatTable() map[Float16]float32 {
	table := map[Float16]float32{
		0x3c00: 1,
		0x4000: 2,
		0xc000: -2,
		0x7bfe: 65472,
		0x7bff: 65504,
		0xfbff: -65504,
		0x0000: 0,
		0x8000: float32(math.Copysign(0, -1)),
		0x7c00: float32(math.Inf(1)),
		0xfc00: float32(math.Inf(-1)),
		0x5b8f: 241.875,
		0x48c8: 9.5625,
	}
	return table
}

func TestFloat32(t *testing.T) {
	for k, v := range getFloatTable() {
		f := k.Float32()
		if f != v {
			t.Errorf("ToFloat32(%d) = %f, want %f.", k, f, v)
		}
	}
}

func TestNewFloat16(t *testing.T) {
	for k, v := range getFloatTable() {
		i := NewFloat16(v)
		if i != k {
			t.Errorf("FromFloat32(%f) = %d, want %d.", v, i, k)
		}
	}
}
