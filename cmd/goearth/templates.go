package main

var templateBasic = `
package main

import (
	"log"
	"strings"

	g "xabbo.b7c.io/goearth"
	"{{.MsgPackage}}/in"
	"{{.MsgPackage}}/out"
)

var ext = g.NewExt(g.ExtInfo{
	Title: "{{.Title}}",
	Description: "{{.Description}}",
	Author: "{{.Author}}",
	Version: "{{.Version}}",
})

func main() {
	ext.Initialized(onInitialized)
	ext.Connected(onConnected)
	ext.Disconnected(onDisconnected)
	ext.Intercept({{.InChatIdentifiers}}).With(handleChat)
	ext.Run()
}

func onInitialized(e g.InitArgs) {
	log.Println("Extension initialized")
}

func onConnected(e g.ConnectArgs) {
	log.Printf("Game connected (%s)\n", e.Host)
}

func onDisconnected() {
	log.Println("Game disconnected")
}

func handleChat(e *g.Intercept) {
	e.Packet.ReadInt() // skip entity index
	msg := e.Packet.ReadString()
	if strings.Contains(msg, "hello") {
		log.Println("Received hello, sending wave")
		ext.Send({{.OutWaveArgs}})
	}
}
`

type TemplateBasicData struct {
	Module      string
	Package     string
	MsgPackage  string
	Title       string
	Description string
	Author      string
	Version     string

	InChatIdentifiers string
	OutWaveArgs       string
}
