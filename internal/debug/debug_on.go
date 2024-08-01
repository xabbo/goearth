//go:build goearth_debug

package debug

import "log"

const Enabled = true

func init() {
	log.Printf("goearth_debug enabled")
}
