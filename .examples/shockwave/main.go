package main

import (
	"log"
	"strings"

	g "xabbo.b7c.io/goearth"
)

var ext = g.NewExt(g.ExtInfo{
	Title:       "Go Earth",
	Description: "example: shockwave",
	Version:     "1.0",
	Author:      "b7",
})

func main() {
	ext.Initialized(func(e *g.InitArgs) { log.Printf("Initialized (connected=%t)", e.Connected) })
	ext.Connected(func(e *g.ConnectArgs) { log.Printf("Connected (%s)", e.Host) })
	ext.Disconnected(func() { log.Println("Disconnected") })
	// intercept arbitrary message names with InterceptIn/Out
	ext.InterceptIn("Chat", "Chat_2", "Chat_3").With(handleChat)
	ext.Run()
}

// wave when someone says "hello"
func handleChat(e *g.InterceptArgs) {
	e.Packet.Skip(0) // skip an integer
	msg := e.Packet.ReadString()
	if strings.Contains(strings.ToLower(msg), "hello") {
		// send packets with arbitrary message names,
		// not defined in `in`/`out` packages
		ext.Send(g.Out.Id("Wave"))
	}
}
