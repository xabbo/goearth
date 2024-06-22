package goearth

import (
	"math/bits"
)

// Shockwave encoding

func vl64len(v int) int {
	if v < 0 {
		v *= -1
	}
	return (bits.Len32(uint32(v)) + 9) / 6
}

func vl64lenEncoded(b byte) int {
	return int(b >> 3 & 7)
}

func vl64encode(b []byte, v int) {
	abs := v
	if abs < 0 {
		abs *= -1
	}
	n := vl64len(v)

	b[0] = 64 | (byte(n)&7)<<3 | byte(abs&3)
	if v < 0 {
		b[0] |= 4
	}
	for i := 1; i < n; i++ {
		b[i] = 64 | byte((abs>>(2+6*(i-1)))&0x3f)
	}
}

func vl64decode(b []byte) int {
	value := int(b[0] & 3)

	n := vl64lenEncoded(b[0])
	for i := 1; i < n; i++ {
		value |= int(b[i]&0x3f) << (2 + 6*(i-1))
	}

	if b[0]&4 != 0 {
		value *= -1
	}
	return value
}

func b64encode(b []byte, v int) {
	for i := 0; i < len(b); i++ {
		b[i] = 64 | byte(v>>((len(b)-i-1)*6)&0x3f)
	}
}

func b64decode(b []byte) int {
	v := 0
	for i := 0; i < len(b); i++ {
		v |= int(b[i]&0x3f) << ((len(b) - i - 1) * 6)
	}
	return v
}
