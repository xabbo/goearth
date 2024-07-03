package main

import (
	"log"
	"os"

	g "xabbo.b7c.io/goearth"
	"xabbo.b7c.io/goearth/in"
)

var ext = g.NewExt(g.ExtInfo{
	Title:       "Go Earth",
	Description: "example: chatlog",
	Version:     "1.0",
	Author:      "b7",
})

var entities = map[int]Entity{}
var roomCache = map[g.Id]RoomData{}
var currentRoomId g.Id
var currentRoomData RoomData
var info *log.Logger

func init() {
	flags := log.Ldate | log.Ltime
	log.SetFlags(flags)
	info = log.New(os.Stderr, "", flags)
}

func main() {
	ext.Initialized(func(_ g.InitArgs) { info.Printf("Extension initialized") })
	ext.Connected(func(_ g.ConnectArgs) { info.Printf("Game connection established") })
	ext.Intercept(in.GetGuestRoomResult).With(handleRoomData)
	ext.Intercept(in.OpenConnection).With(handleInitRoom)
	ext.Intercept(in.Users).With(handleUsers)
	ext.Intercept(in.Chat, in.Shout, in.Whisper).With(handleChat)
	ext.Disconnected(func() { info.Printf("Game connection lost") })
	ext.Run()
}

func handleRoomData(e *g.Intercept) {
	roomData := RoomData{}
	e.Packet.Read(&roomData)
	roomCache[roomData.Id] = roomData
	if currentRoomId == roomData.Id {
		if currentRoomData.Name != roomData.Name {
			info.Printf("Room name changed to %q", roomData.Name)
		}
		currentRoomData = roomData
	}
}

func handleInitRoom(e *g.Intercept) {
	for k := range entities {
		delete(entities, k)
	}
	e.Packet.Read(&currentRoomId)
	if room, ok := roomCache[currentRoomId]; ok {
		currentRoomData = room
		info.Printf("Entered room %q by %s (id:%d)", room.Name, room.OwnerName, room.Id)
	}
}

func handleUsers(e *g.Intercept) {
	ents := []Entity{}
	e.Packet.Read(&ents)
	for _, ent := range ents {
		entities[ent.Index] = ent
	}
}

func handleChat(e *g.Intercept) {
	idx, msg := 0, ""
	e.Packet.Read(&idx, &msg)
	if ent, ok := entities[idx]; ok && ent.Type == USER {
		log.Printf("%s: %s", ent.Name, msg)
	}
}
