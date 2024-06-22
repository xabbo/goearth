package encoding

import (
	"math/bits"
)

// Shockwave encoding

// VL64Len returns the number of bytes required to represent a variable-length base64-encoded integer.
func VL64Len(v int) int {
	if v < 0 {
		v *= -1
	}
	return (bits.Len32(uint32(v)) + 9) / 6
}

func VL64EncodedLen(b byte) int {
	return int(b >> 3 & 7)
}

// VL64Encode encodes an integer to variable-length base64 into the specified byte slice.
func VL64Encode(b []byte, v int) {
	abs := v
	if abs < 0 {
		abs *= -1
	}
	n := VL64Len(v)

	b[0] = 64 | (byte(n)&7)<<3 | byte(abs&3)
	if v < 0 {
		b[0] |= 4
	}
	for i := 1; i < n; i++ {
		b[i] = 64 | byte((abs>>(2+6*(i-1)))&0x3f)
	}
}

// VL64Decode decodes a variable-length base64-encoded integer from the specified byte slice.
func VL64Decode(b []byte) int {
	value := int(b[0] & 3)

	n := VL64EncodedLen(b[0])
	for i := 1; i < n; i++ {
		value |= int(b[i]&0x3f) << (2 + 6*(i-1))
	}

	if b[0]&4 != 0 {
		value *= -1
	}
	return value
}

// B64Encode encodes an integer to base64 into the specified byte slice.
func B64Encode(b []byte, v int) {
	for i := 0; i < len(b); i++ {
		b[i] = 64 | byte(v>>((len(b)-i-1)*6)&0x3f)
	}
}

// B64Decode decodes a base64-encoded integer from the specified byte slice.
func B64Decode(b []byte) int {
	v := 0
	for i := 0; i < len(b); i++ {
		v |= int(b[i]&0x3f) << ((len(b) - i - 1) * 6)
	}
	return v
}
