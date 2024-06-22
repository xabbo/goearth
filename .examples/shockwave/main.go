package main

import (
	"fmt"

	g "xabbo.b7c.io/goearth"
	"xabbo.b7c.io/goearth/in"
	"xabbo.b7c.io/goearth/out"
)

var ext = g.NewExt(g.ExtInfo{
	Title:       "Go Earth",
	Description: "example: shockwave",
	Version:     "1.0",
	Author:      "b7",
})

func main() {
	ext.Initialized(func(e *g.InitArgs) { fmt.Printf("Initialized (connected=%t)\n", e.Connected) })
	ext.Connected(func(e *g.ConnectArgs) { fmt.Printf("Connected (%s)\n", e.Host) })
	ext.Intercept(in.Chat).With(handleChat)
	ext.Intercept(out.Chat).With(handleChat)
	ext.Connected(func(e *g.ConnectArgs) {
		fmt.Println(e.Client)
		pkt := ext.NewPacket(g.Identifier{g.Out, "CHAT"})
		pkt.WriteString(fmt.Sprintf("connected to %s", e.Client.Identifier))
		ext.SendPacket(pkt)
	})
	ext.Run()
}

func handleChat(e *g.InterceptArgs) {
	if e.Dir() == g.In {
		index := e.Packet.ReadInt()
		msg := e.Packet.ReadString()
		fmt.Println(index, msg)
	} else if e.Dir() == g.Out {
		msg := e.Packet.ReadString()
		fmt.Println(msg)
	}
}
