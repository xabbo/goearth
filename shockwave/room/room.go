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
	Id            int
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

func (obj Object) String() string {
	return obj.Class + "(" + strconv.Itoa(obj.Id) + ")"
}

func (obj *Object) Parse(p *g.Packet, pos *int) {
	strId := p.ReadStringPtr(pos)
	id, err := strconv.Atoi(strId)
	if err != nil {
		panic(fmt.Errorf("invalid object ID: %q", strId))
	}

	*obj = Object{Id: id}
	p.ReadPtr(pos, &obj.Class,
		&obj.X, &obj.Y, &obj.Width, &obj.Height,
		&obj.Direction, &obj.Z,
		&obj.Colors, &obj.RuntimeData,
		&obj.Extra, &obj.StuffData)
}

// Item represents a wall item in a room.
type Item struct {
	Id       int
	Class    string
	Owner    string
	Location string
	Type     string
}

func (item Item) String() string {
	return item.Class + "(" + strconv.Itoa(item.Id) + ")"
}

func (item *Item) Parse(p *g.Packet, pos *int) {
	item.ParseString(p.ReadStringPtr(pos))
}

type Items []Item

func (items *Items) Parse(p *g.Packet, pos *int) {
	*items = []Item{}
	for p.Pos < p.Length() {
		line := strings.TrimSuffix(p.ReadStringPtr(pos), "\r")
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

type SlideObjectBundle struct {
	From, To      Point
	Objects       []SlideObject
	RollerId      int
	SlideMoveType SlideMoveType
	Entity        SlideObject
}

func (bundle *SlideObjectBundle) Parse(p *g.Packet, pos *int) {
	*bundle = SlideObjectBundle{}
	p.ReadPtr(pos, &bundle.From, &bundle.To, &bundle.Objects, &bundle.RollerId)
	if p.Pos < p.Length() {
		p.ReadPtr(pos, &bundle.SlideMoveType)
		if bundle.SlideMoveType == SlideMoveTypeMove || bundle.SlideMoveType == SlideMoveTypeSlide {
			p.ReadPtr(pos, &bundle.Entity)
		}
	}
}

type SlideObject struct {
	Id         int
	FromZ, ToZ float64
}

type SlideMoveType int

const (
	SlideMoveTypeNone SlideMoveType = iota
	SlideMoveTypeMove
	SlideMoveTypeSlide
)

func (slideType *SlideMoveType) Parse(p *g.Packet, pos *int) {
	*slideType = SlideMoveType(p.ReadIntPtr(pos))
}

type EntityType int

const (
	User EntityType = iota + 1
	Pet
	PublicBot
	PrivateBot
)

func (entityType EntityType) String() string {
	switch entityType {
	case User:
		return "user"
	case Pet:
		return "pet"
	case PublicBot:
		return "public bot"
	case PrivateBot:
		return "private bot"
	default:
		return strconv.Itoa(int(entityType))
	}
}

func (entityType *EntityType) Parse(p *g.Packet, pos *int) {
	*entityType = EntityType(p.ReadIntPtr(pos))
}

func (entityType EntityType) Compose(p *g.Packet, pos *int) {
	p.WriteIntPtr(pos, int(entityType))
}

// Point represents 2-dimensional coordinates in a room.
type Point struct {
	X, Y int
}

func (pt Point) String() string {
	return strconv.Itoa(pt.X) + ", " + strconv.Itoa(pt.Y)
}

// ToTile converts the Point to a Tile with Z = 0.
func (pt Point) ToTile() Tile {
	return Tile{pt.X, pt.Y, 0}
}

// Tile represents 3-dimensional coordinates in a room.
type Tile struct {
	X, Y int
	Z    float64
}

func (tile Tile) String() string {
	return strconv.Itoa(tile.X) + ", " + strconv.Itoa(tile.Y) + ", " + strconv.FormatFloat(tile.Z, 'f', 2, 64)
}

// ToPoint converts the Tile to a Point.
func (tile Tile) ToPoint() Point {
	return Point{tile.X, tile.Y}
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

func (ent Entity) String() string {
	return ent.Name
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
