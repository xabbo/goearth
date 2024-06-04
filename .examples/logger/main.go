package main

import (
	"flag"
	"fmt"
	"log"

	g "xabbo.b7c.io/goe"
)

const (
	RED   = "\x1b[31m"
	CYAN  = "\x1b[36m"
	RESET = "\x1b[0m"
)

var ext = g.NewExt(g.ExtInfo{
	Title:       "Go Earth",
	Description: "example: logger",
	Version:     "1.0",
	Author:      "b7",
})

var logBytes bool

func init() {
	log.SetFlags(log.Ltime | log.Lmicroseconds)

	g.InitFlags()
	flag.BoolVar(&logBytes, "b", false, "Print out packet bytes.")
	flag.Parse()
}

func main() {
	ext.Initialized(func(e *g.InitArgs) {
		log.Printf("Extension initialized (connected=%t)", e.Connected)
	})
	ext.Connected(func(e *g.ConnectArgs) {
		log.Printf("Game connection established (%s:%d)", e.Host, e.Port)
		log.Printf("Client %s (%s)", e.Client.Identifier, e.Client.Version)
	})
	ext.InterceptAll(handleIntercept)
	ext.Disconnected(func() { log.Printf("Game connection lost") })
	ext.Run()
}

func handleIntercept(e *g.InterceptArgs) {
	var indicator, color string
	bytes := fmt.Sprintf("(%d bytes)", e.Packet.Length())
	if logBytes {
		bytes = fmt.Sprintf("[% x]", e.Packet.Data)
	}
	switch e.Dir() {
	case g.In:
		indicator = "<<"
		color = RED
	case g.Out:
		indicator = ">>"
		color = CYAN
	}
	log.Printf("%s%s %6d %s %s%s\n", color, indicator, e.Sequence(), e.Packet.Header.Name, bytes, RESET)
}
