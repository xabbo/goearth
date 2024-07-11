package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"

	g "xabbo.b7c.io/goearth"
	"xabbo.b7c.io/goearth/shockwave/out"
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

var opts struct {
	filter string
	block  string
}

var filterRegex, blockRegex *regexp.Regexp

func init() {
	log.SetFlags(log.Ltime | log.Lmicroseconds)
	log.SetOutput(os.Stdout)

	g.InitFlags()
	flag.StringVar(&opts.filter, "filter", "", "Regex to filter packet names.")
	flag.StringVar(&opts.block, "block", "", "Regex to block packets.")
	flag.Parse()

	if opts.filter != "" {
		filterRegex = regexp.MustCompile("(?i)" + opts.filter)
	}
	if opts.block != "" {
		blockRegex = regexp.MustCompile("(?i)" + opts.block)
	}
}

func main() {
	ext.Initialized(func(e g.InitArgs) {
		log.Printf("Extension initialized (connected=%t)", e.Connected)
	})
	ext.Connected(func(e g.ConnectArgs) {
		log.Printf("Game connection established (%s:%d)", e.Host, e.Port)
		log.Printf("Client %s (%s)", e.Client.Identifier, e.Client.Version)
	})
	ext.InterceptAll(handleIntercept)
	ext.Disconnected(func() { log.Printf("Game connection lost") })
	ext.Run()
}

func handleIntercept(e *g.Intercept) {
	name := e.Name()
	if blockRegex != nil && blockRegex.MatchString(name) {
		e.Block()
		log.Printf("blocked %s", name)
	}
	if filterRegex != nil && !filterRegex.MatchString(name) {
		return
	}
	if e.Is(out.TRY_LOGIN) {
		return
	}

	var indicator, color string
	bytes := fmt.Sprintf("(%d bytes)", e.Packet.Length())
	switch e.Dir() {
	case g.In:
		indicator = "<<"
		color = RED
	case g.Out:
		indicator = ">>"
		color = CYAN
	}
	log.Printf("%s%s %4d %s %s%s\n%s",
		color, indicator, e.Packet.Header.Value, name, bytes, RESET, hex.Dump(e.Packet.Data))
}
