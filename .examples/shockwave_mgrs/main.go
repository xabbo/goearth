package main

import (
	g "xabbo.b7c.io/goearth"
	"xabbo.b7c.io/goearth/shockwave/inventory"
	"xabbo.b7c.io/goearth/shockwave/nav"
	"xabbo.b7c.io/goearth/shockwave/profile"
	"xabbo.b7c.io/goearth/shockwave/room"
	"xabbo.b7c.io/goearth/shockwave/trade"
)

// An extension for testing the shockwave game state managers.
// Run with `-tags goearth_debug` to see debug output.

var (
	ext = g.NewExt(g.ExtInfo{
		Title:       "Go Earth",
		Description: "example: shockwave managers",
		Author:      "b7",
		Version:     "1.0",
	})
)

func main() {
	inventory.NewManager(ext)
	room.NewManager(ext)
	trade.NewManager(ext)
	profile.NewManager(ext)
	nav.NewManager(ext)
	ext.Run()
}
