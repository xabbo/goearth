package main

import g "xabbo.b7c.io/goearth"

type EntityType int

func (t *EntityType) Parse(p *g.Packet, pos *int) {
	*t = EntityType(p.ReadIntPtr(pos))
}

const (
	USER EntityType = iota + 1
	PET
	PUBLIC_BOT
	PRIVATE_BOT
)

type Tile struct {
	X int
	Y int
	Z float64
}

type EntityBase struct {
	Id     g.Id
	Name   string
	Motto  string
	Figure string
	Index  int
	Tile   Tile
	Dir    int
	Type   EntityType
}

type User struct {
	Gender           string
	GroupId          g.Id
	GroupStatus      int
	GroupName        string
	FigureExtra      string
	AchievementScore int
	IsModerator      bool
}

type Pet struct {
	Breed                int
	OwnerId              g.Id
	OwnerName            string
	RarityLevel          int
	HasSaddle            bool
	IsRiding             bool
	CanBreed             bool
	CanHarvest           bool
	CanRevive            bool
	HasBreedingPermssion bool
	Level                int
	Posture              string
}

type PrivateBot struct {
	Gender    string
	OwnerId   g.Id
	OwnerName string
	Skills    []int16
}

type Entity struct {
	EntityBase
	Extra any
}

func (e *Entity) Parse(p *g.Packet, pos *int) {
	*e = Entity{}
	p.ReadPtr(pos, &e.EntityBase)
	switch e.Type {
	case USER:
		e.Extra = &User{}
	case PET:
		e.Extra = &Pet{}
	case PRIVATE_BOT:
		e.Extra = &PrivateBot{}
	}
	if e.Extra != nil {
		p.ReadPtr(pos, e.Extra)
	}
}
