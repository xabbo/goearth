package encoding

import (
	"testing"
)

// value -> expected length
var vl64len_tests = map[int]int{
	0:        1,
	1:        1,
	-1:       1,
	2:        1,
	-2:       1,
	3:        1,
	-3:       1,
	4:        2,
	-4:       2,
	128:      2,
	-128:     2,
	255:      2,
	-255:     2,
	256:      3,
	-256:     3,
	8192:     3,
	-8192:    3,
	16383:    3,
	-16383:   3,
	16384:    4,
	-16384:   4,
	1048575:  4,
	-1048575: 4,
	1048576:  5,
	-1048576: 5,
}

// value -> encoded
var vl64_tests = map[int]string{
	-1:       "M",
	0:        "H",
	1:        "I",
	2:        "J",
	3:        "K",
	4:        "PA",
	64:       "PP",
	250:      "R~",
	256:      "X@A",
	-256:     "\\@A",
	1024:     "X@D",
	16384:    "`@@A",
	-16384:   "d@@A",
	1048576:  "h@@@A",
	-1048576: "l@@@A",
}

// value -> encoded
var b64_tests = map[int]string{
	0:    "@@",
	1:    "@A",
	16:   "@P",
	256:  "D@",
	1337: "Ty",
	2048: "`@",
	4000: "~`",
}

func TestVL64len(t *testing.T) {
	for value, expected := range vl64len_tests {
		actual := VL64EncodeLen(value)
		if actual != expected {
			t.Fatalf("vl64len(%d) was %d, expected %d", value, actual, expected)
		}
	}
}

func TestVL64encode(t *testing.T) {
	for value, expected := range vl64_tests {
		length := VL64EncodeLen(value)
		buf := make([]byte, length)
		VL64Encode(buf, value)
		actual := string(buf)
		if actual != expected {
			t.Fatalf("vl64encode(%d) was %q, expected %q", value, actual, expected)
		}
	}
}

func TestVL64decode(t *testing.T) {
	for expected, encoded := range vl64_tests {
		buf := []byte(encoded)
		actual := VL64Decode(buf)
		if actual != expected {
			t.Fatalf("vl64decode(%q) was %d, expected %d", encoded, actual, expected)
		}
	}
}

func TestB64encode(t *testing.T) {
	for value, expected := range b64_tests {
		buf := make([]byte, 2)
		B64Encode(buf, value)
		actual := string(buf)
		if actual != expected {
			t.Fatalf("b64encode(%d) was %q, expected %q", value, actual, expected)
		}
	}
}

func TestB64decode(t *testing.T) {
	for expected, encoded := range b64_tests {
		actual := B64Decode([]byte(encoded))
		if actual != expected {
			t.Fatalf("b64decode(%q) was %d, expected %d", encoded, actual, expected)
		}
	}
}
