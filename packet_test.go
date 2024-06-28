package goearth

import "testing"

var testClients = []ClientType{Flash, Shockwave, Unity}
var testDirections = []Direction{In, Out}

func runTestClientsDirections(test func(clientType ClientType, dir Direction)) {
	for _, clientType := range testClients {
		for _, direction := range testDirections {
			test(clientType, direction)
		}
	}
}

func TestStringReplacements(t *testing.T) {
	var stringReplaceTests = [][2]string{
		{"hello", "world"},                     // equal length
		{"hello", "goodbye"},                   // extend
		{"test", "this is a very long string"}, // long extend
		{"hello", "bye"},                       // shrink
		{"this is a very long string", "test"}, // long shrink
	}

	runTestClientsDirections(func(clientType ClientType, dir Direction) {
		t.Logf("testing client %s (%s)", clientType, dir)
		for _, testData := range stringReplaceTests {
			t.Logf("replacing string %q -> %q", testData[0], testData[1])
			expectedDiff := len([]byte(testData[1])) - len([]byte(testData[0]))

			pkt := &Packet{Client: clientType, Header: &NamedHeader{Header: Header{Dir: dir}}}
			pkt.WriteInt(31337)
			pkt.WriteString(testData[0])
			pkt.WriteInt(31337)
			preLen := pkt.Length()

			pkt.Pos = 4
			t.Logf("length before replacement: %d", pkt.Length())
			pkt.ReplaceString(testData[1])
			t.Logf("length after replacement: %d", pkt.Length())
			postLen := pkt.Length()

			actualDiff := postLen - preLen
			if actualDiff != expectedDiff {
				t.Fatalf("incorrect difference, expected: %d, actual: %d (client: %s)", expectedDiff, actualDiff, clientType)
			}

			pkt.Pos = 0
			if pkt.ReadInt() != 31337 {
				t.Fatalf("corrupted head after string replacement (client: %s)", clientType)
			}

			actualValue := pkt.ReadString()
			expectedValue := testData[1]
			if actualValue != expectedValue {
				t.Fatalf("replaced string is incorrect, expected: %q, actual: %q (client: %s)", expectedValue, actualValue, clientType)
			}

			if pkt.ReadInt() != 31337 {
				t.Fatalf("corrupted tail after string replacement (client: %s)", clientType)
			}
		}
	})
}
