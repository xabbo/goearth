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

func (item *Item) Parse(r g.PacketReader) {
	*item = Item{}
	item.ItemId = r.ReadInt()
	item.Type = r.ReadString()
	item.Id = r.ReadInt()
	item.ClassId = r.ReadInt()
	item.Class = r.ReadString()
	if item.Type == "s" {
		item.Colors = r.ReadString()
		item.DimX = r.ReadInt()
		item.DimY = r.ReadInt()
	} else {
		item.DimX = 1
		item.DimY = 1
	}
	item.Category = r.ReadString()
	item.Groupable = r.ReadInt()
	item.Data = r.ReadString()
	item.Day = r.ReadInt()
	item.Month = r.ReadInt()
	item.Year = r.ReadInt()
	if item.Type == "s" {
		item.SongId = r.ReadInt()
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
