package room

import (
	"fmt"
	"strconv"
	"strings"

	g "xabbo.b7c.io/goearth"
)

// Info contains information about a room.
type Info struct {
	CanOthersMoveFurni  bool
	Door                int
	Id                  int
	Owner               string
	Marker              string
	Name                string
	Description         string
	ShowOwnerName       bool
	Trading             int
	Alert               int
	MaxVisitors         int
	AbsoluteMaxVisitors int
}

// Object represents a floor item in a room.
type Object struct {
	Id            string
	Class         string
	X, Y          int
	Width, Height int
	Direction     int
	Z             float64
	Colors        string
	RuntimeData   string
	Extra         int
	StuffData     string
}

// Item represents a wall item in a room.
type Item struct {
	Id       int
	Class    string
	Owner    string
	Location string
	Type     string
}

type Items []Item

func (items *Items) Parse(p *g.Packet, pos *int) {
	*items = []Item{}
	for _, line := range strings.Split(p.ReadString(), "\r") {
		if line == "" {
			continue
		}
		var item Item
		item.Parse(line)
		*items = append(*items, item)
	}
}

func (item *Item) Parse(s string) {
	fields := strings.Split(s, "\t")
	if len(fields) != 5 {
		panic(fmt.Errorf("Item field length != 5: %d", len(fields)))
	}
	id, err := strconv.Atoi(fields[0])
	if err != nil {
		panic("failed to parse Item ID: " + fields[0])
	}
	item.Id = id
	item.Class = fields[1]
	item.Owner = fields[2]
	item.Location = fields[3]
	item.Type = fields[4]
}

type EntityType int

const (
	User EntityType = iota + 1
	Pet
	PublicBot
	PrivateBot
)

func (entityType *EntityType) Parse(p *g.Packet, pos *int) {
	*entityType = EntityType(p.ReadIntPtr(pos))
}

// Entity represents a user, pet or bot in a room.
type Entity struct {
	Index      int
	Name       string
	Figure     string
	Gender     string
	Custom     string
	X, Y       int
	Z          float64
	PoolFigure string
	BadgeCode  string
	Type       EntityType
}

type ChatType int

const (
	Talk ChatType = iota + 1
	Whisper
	Shout
)
