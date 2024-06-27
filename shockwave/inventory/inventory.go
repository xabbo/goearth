package inventory

import (
	g "xabbo.b7c.io/goearth"
)

// Inventory represents a list of inventory items.
type Inventory struct {
	Items []Item
}

func (inv *Inventory) Parse(p *g.Packet, pos *int) {
	*inv = Inventory{}
	p.Read(&inv.Items)
}

type ItemType string

const (
	Floor ItemType = "S"
	Wall  ItemType = "I"
)

func (itemType *ItemType) Parse(p *g.Packet, pos *int) {
	*itemType = ItemType(p.ReadStringPtr(pos))
}

// Item represents an inventory item.
type Item struct {
	ItemId int
	Pos    int
	// Type represents the type of the item.
	// May be "S" for "stuff" (floor item), or "I" for "item" (wall item).
	Type       ItemType
	Id         int
	Class      string
	DimX, DimY int
	Colors     string
	Props      string
}

func (item *Item) Parse(p *g.Packet, pos *int) {
	*item = Item{}
	p.ReadPtr(pos, &item.ItemId, &item.Pos, &item.Type, &item.Id, &item.Class)
	switch item.Type {
	case "S":
		p.Read(&item.DimX, &item.DimY, &item.Colors)
	case "I":
		p.Read(&item.Props)
	}
}
