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

func (item *Item) Parse(p *g.Packet, pos *int) {
	item.ParseString(p.ReadString())
}

type Items []Item

func (items *Items) Parse(p *g.Packet, pos *int) {
	*items = []Item{}
	for p.Pos < p.Length() {
		line := strings.TrimSuffix(p.ReadString(), "\r")
		var item Item
		item.ParseString(line)
		*items = append(*items, item)
	}
}

func (item *Item) ParseString(s string) {
	fields := strings.Split(s, "\t")
	if len(fields) != 5 {
		panic(fmt.Errorf("Item field length != 5: %d", len(fields)))
	}

	id, err := strconv.Atoi(fields[0])
	if err != nil {
		panic("failed to parse Item ID: " + fields[0])
	}

	*item = Item{
		Id:       id,
		Class:    fields[1],
		Owner:    fields[2],
		Location: fields[3],
		Type:     fields[4],
	}
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

// Tile represents 3-dimensional coordinates in a room.
type Tile struct {
	X, Y int
	Z    float64
}

type EntityBase struct {
	Index  int
	Name   string
	Figure string
	Gender string
	Custom string
	Tile
	PoolFigure string
	BadgeCode  string
	Type       EntityType
}

// Entity represents a user, pet or bot in a room.
type Entity struct {
	EntityBase
	Dir     int
	HeadDir int
	Action  string
}

func (ent *Entity) Parse(p *g.Packet, pos *int) {
	*ent = Entity{}
	p.ReadPtr(pos, &ent.EntityBase)
}

func (ent *Entity) Compose(p *g.Packet, pos *int) {
	p.WritePtr(pos, ent.EntityBase)
}

// EntityStatus represents a status update of an entity in a room.
type EntityStatus struct {
	Index int
	Tile
	HeadDir, BodyDir int
	Action           string
}

type ChatType int

const (
	Talk ChatType = iota + 1
	Whisper
	Shout
)
