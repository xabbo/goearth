//go:build !goearth_debug

package main

import (
	"fmt"
	"os"
)

func init() {
	fmt.Fprintln(os.Stderr, "Run this extension with the `-tags goearth_debug` build tag to see debug output.")
	os.Exit(1)
}
