package main

import (
	"fmt"
	"log"
	"strings"

	g "xabbo.b7c.io/goearth"
	"xabbo.b7c.io/goearth/in"
	"xabbo.b7c.io/goearth/out"
)

var ext = g.NewExt(g.ExtInfo{
	Title:       "Go Earth",
	Description: "example: basic",
	Version:     "1.0",
	Author:      "b7",
})

func init() {
	log.SetFlags(log.Ltime | log.Lmicroseconds)
}

func main() {
	// Handling extension initialization
	ext.Initialized(func(e g.InitArgs) {
		log.Printf("Extension initialized (connected=%t)", e.Connected)
	})

	// Handling extension activation
	// (when the green "play" button is clicked in G-Earth)
	ext.Activated(func() {
		log.Printf("Extension activated")
		if ext.IsConnected() {
			// Launch a goroutine to retrieve the user's info
			// ensuring the main extension loop can continue
			go getUserInfo()
		} else {
			log.Printf("Game is not connected")
		}
	})
	// Handling game connection start
	ext.Connected(func(e g.ConnectArgs) {
		log.Printf("Game connection established (%s:%d)", e.Host, e.Port)
		log.Printf("Client %s (%s)", e.Client.Identifier, e.Client.Version)
		log.Printf("Received %d message info", len(e.Messages))
	})
	// Intercepting all packets
	ext.InterceptAll(func(e *g.Intercept) {
		if ext.Headers().Is(e.Packet.Header, in.Ping) {
			log.Printf("Received ping")
		}
	})
	// Intercepting specific packets
	ext.Intercept(out.Chat, out.Shout, out.Whisper).With(handleChat)
	// Handling game disconnection
	ext.Disconnected(func() { log.Printf("Game connection lost") })
	ext.Run()
}

func handleChat(e *g.Intercept) {
	// Reading data from packets
	msg := e.Packet.ReadString()
	action := strings.ToLower(e.Name())
	if action == "chat" {
		action = "said"
	} else {
		action += "ed"
	}
	log.Printf("You %s %q", action, msg)
	if strings.Contains(msg, "block") {
		// Blocking packets
		e.Block()
		log.Printf("Blocking message!")
	} else if strings.Contains(msg, "apple") {
		// Modifying packets
		msg = strings.ReplaceAll(msg, "apple", "orange")
		e.Packet.ReplaceStringAt(0, msg)
		log.Printf("Replacing message: %q", msg)
	}
}

func getUserInfo() {
	log.Printf("Retrieving user info...")
	// Sending packets
	ext.Send(out.InfoRetrieve)
	// Receiving packets inline
	/*
		Block() prevents the packet from reaching the client.
		You can also chain Timeout(), TimeoutSec() or TimeoutMs()
		to adjust the timeout. The default timeout is 1 minute.
		If() can be chained to pass in a function that returns true
		if the packet should be intercepted.
	*/
	if pkt := ext.Recv(in.UserObject).Block().Wait(); pkt != nil {
		var id g.Id
		var name string
		pkt.Read(&id, &name)
		msg := fmt.Sprintf("Got user info. (id:%d, name:%q)", id, name)
		log.Println(msg)
		// Sending client-side packets
		ext.Send(in.Chat, 0, msg, 0, 34, 0, 0)
	} else {
		log.Printf("Timed out.")
	}
}
