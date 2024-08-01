package trade

import (
	"xabbo.b7c.io/goearth/shockwave/inventory"
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
	Name     string
	Accepted bool
	Items    []Item
}

// Item is a type alias for [inventory.Item].
type Item = inventory.Item
