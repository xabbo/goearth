package nav

import (
	"os"
	"path/filepath"
	"testing"

	g "xabbo.b7c.io/goearth"
)

func TestNavNodeInfo(t *testing.T) {
	dir := ".testdata/navnodeinfo"
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	for _, entry := range entries {
		t.Run(entry.Name(), func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
			if err != nil {
				t.Fatalf("err: %s", err)
			}

			p := &g.Packet{
				Client: g.Shockwave,
				Header: g.Header{Dir: g.In},
				Data:   data,
			}

			var navNodeInfo NodeInfo
			navNodeInfo.Parse(p, &p.Pos)

			if p.Pos < len(p.Data) {
				t.Fatal("parser failed to read entire packet")
			}
		})
	}
}

func TestFlatResults(t *testing.T) {
	dir := ".testdata/flat_results"
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	for _, entry := range entries {
		t.Run(entry.Name(), func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
			if err != nil {
				t.Fatalf("err: %s", err)
			}

			p := &g.Packet{
				Client: g.Shockwave,
				Header: g.Header{Dir: g.In},
				Data:   data,
			}

			var results Rooms
			results.Parse(p, &p.Pos)

			if p.Pos < len(p.Data) {
				t.Fatalf("parser failed to read entire packet")
			}
		})
	}
}
