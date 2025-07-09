package trade

import (
	g "xabbo.b7c.io/goearth"
)

// Offers is an array that holds the offers of the trader and tradee, respectively.
type Offers [2]Offer

// Trader returns the offer of the user who initiated the trade.
func (offers Offers) Trader() Offer {
	return offers[0]
}

// Tradee returns the offer of the user who received the trade request.
func (offers Offers) Tradee() Offer {
	return offers[1]
}

// Offer represents a user's offer in a trade.
type Offer struct {
	Accepted  bool
	UserId    int
	ItemCount int
	Items     []Item
}

type Item struct {
	ItemId     int
	Type       string
	Id         int
	ClassId    int
	Class      string
	Colors     string
	DimX, DimY int
	Category   string
	Groupable  int
	Data       string
	Day        int
	Month      int
	Year       int
	SongId     int
	SongName   string
	SongDesc   string
	PosterName string
}

func (item *Item) Parse(p *g.Packet, pos *int) {
	*item = Item{}
	item.ItemId = p.ReadIntPtr(pos)
	item.Type = p.ReadStringPtr(pos)
	item.Id = p.ReadIntPtr(pos)
	item.ClassId = p.ReadIntPtr(pos)
	item.Class = p.ReadStringPtr(pos)
	if item.Type == "s" {
		item.Colors = p.ReadStringPtr(pos)
		item.DimX = p.ReadIntPtr(pos)
		item.DimY = p.ReadIntPtr(pos)
	} else {
		item.DimX = 1
		item.DimY = 1
	}
	item.Category = p.ReadStringPtr(pos)
	item.Groupable = p.ReadIntPtr(pos)
	item.Data = p.ReadStringPtr(pos)
	item.Day = p.ReadIntPtr(pos)
	item.Month = p.ReadIntPtr(pos)
	item.Year = p.ReadIntPtr(pos)
	if item.Type == "s" {
		item.SongId = p.ReadIntPtr(pos)
		item.SongName = "furni_" + item.Class + "_name"
		item.SongDesc = "furni_" + item.Class + "_desc"
	} else {
		if item.Class == "poster" {
			item.PosterName = "poster_" + item.Data + "_name"
		} else {
			item.PosterName = "wallitem_" + item.Class + "_name"
		}
	}
}
