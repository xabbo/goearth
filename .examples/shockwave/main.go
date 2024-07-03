package main

import (
	"log"
	"strconv"

	g "xabbo.b7c.io/goearth"
	"xabbo.b7c.io/goearth/shockwave/in"
)

// A simple chat log extension for the shockwave client.

var ext = g.NewExt(g.ExtInfo{
	Title:       "Go Earth",
	Description: "example: shockwave",
	Version:     "1.0",
	Author:      "b7",
})

var (
	users            = map[int]*User{}
	usersPacketCount = 0
)

type User struct {
	Index      int
	Name       string
	Figure     string
	Gender     string
	Custom     string
	X, Y       int
	Z          float64
	PoolFigure string
	BadgeCode  string
	Type       int
}

func main() {
	ext.Initialized(func(e g.InitArgs) { log.Printf("Extension initialized") })
	ext.Connected(func(e g.ConnectArgs) { log.Printf("Connected (%s)", e.Host) })
	ext.Disconnected(func() { log.Println("Disconnected") })
	ext.Intercept(in.OPC_OK).With(handleEnterRoom)
	ext.Intercept(in.USERS).With(handleUsers)
	ext.Intercept(in.LOGOUT).With(handleRemoveUser)
	ext.Intercept(in.CHAT, in.CHAT_2, in.CHAT_3).With(handleChat)
	ext.Run()
}

func handleEnterRoom(e *g.Intercept) {
	usersPacketCount = 0
	clear(users)
}

func handleUsers(e *g.Intercept) {
	// Observations:
	// The first USERS packet sent upon entering the room (after OPC_OK)
	// is the list of users that are already in the room.
	// The second USERS packet contains a single user, yourself.
	// The following USERS packets indicate someone entering the room.
	usersPacketCount++
	for range e.Packet.ReadInt() {
		var user User
		e.Packet.Read(&user)
		if user.Type == 1 {
			if usersPacketCount >= 3 {
				log.Printf("* %s entered the room", user.Name)
			}
			users[user.Index] = &user
		}
	}
}

func handleChat(e *g.Intercept) {
	index := e.Packet.ReadInt()
	msg := e.Packet.ReadString()
	if user, ok := users[index]; ok {
		log.Printf("%s: %s", user.Name, msg)
	}
}

func handleRemoveUser(e *g.Intercept) {
	s := e.Packet.ReadString()
	index, err := strconv.Atoi(s)
	if err != nil {
		return
	}
	if user, ok := users[index]; ok {
		log.Printf("* %s left the room.", user.Name)
		delete(users, index)
	}
}
