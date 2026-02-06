package inventory

import (
	"strconv"

	g "xabbo.b7c.io/goearth"
)

// Inventory represents a list of inventory items.
type Inventory struct {
	Items []Item
}

func (inv *Inventory) Parse(r g.PacketReader) {
	*inv = Inventory{}
	r.Read(&inv.Items)
}

type ItemType string

const (
	Floor ItemType = "S"
	Wall  ItemType = "I"
)

func (itemType *ItemType) Parse(r g.PacketReader) {
	*itemType = ItemType(r.ReadString())
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

func (item Item) String() string {
	return item.Class + "(" + strconv.Itoa(item.ItemId) + ")"
}

func (item *Item) Parse(r g.PacketReader) {
	*item = Item{}
	r.Read(&item.ItemId, &item.Pos, &item.Type, &item.Id, &item.Class)
	switch item.Type {
	case "S":
		r.Read(&item.DimX, &item.DimY, &item.Colors)
	case "I":
		r.Read(&item.Props)
	}
}
