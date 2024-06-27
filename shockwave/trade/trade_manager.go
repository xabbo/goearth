package trade

import (
	"strconv"
	"strings"

	g "xabbo.b7c.io/goearth"
	"xabbo.b7c.io/goearth/internal/debug"
	"xabbo.b7c.io/goearth/shockwave/in"
	"xabbo.b7c.io/goearth/shockwave/inventory"
	"xabbo.b7c.io/goearth/shockwave/out"
)

var dbg = debug.NewLogger("[trade]")

// Manager tracks the state of trades.
type Manager struct {
	ext       *g.Ext
	updated   g.Event[Args]
	accepted  g.Event[AcceptArgs]
	completed g.Event[Args]
	closed    g.Event[Args]

	Trading bool
	Offers  Offers
}

// NewManager creates a new trade Manager using the provided extension.
func NewManager(ext *g.Ext) *Manager {
	mgr := &Manager{ext: ext}
	ext.Intercept(in.TRADE_ITEMS).With(mgr.handleTradeItems)
	ext.Intercept(in.TRADE_ACCEPT).With(mgr.handleTradeAccept)
	ext.Intercept(in.TRADE_CLOSE).With(mgr.handleTradeClose)
	ext.Intercept(in.TRADE_COMPLETED_2).With(mgr.handleTradeCompleted2)
	return mgr
}

// Offer offers an item with the specified ID in the current trade.
func (mgr *Manager) Offer(itemId int) {
	// The item ID is a raw string with no length header.
	mgr.ext.Send(out.TRADE_ADDITEM, []byte(strconv.Itoa(itemId)))
}

// OfferItem offers the specified inventory item in the current trade.
func (mgr *Manager) OfferItem(item inventory.Item) {
	mgr.Offer(item.ItemId)
}

// Accept accepts the trade.
func (mgr *Manager) Accept() {
	mgr.ext.Send(out.TRADE_ACCEPT)
}

// Unaccept unaccepts the trade.
func (mgr *Manager) Unaccept() {
	mgr.ext.Send(out.TRADE_UNACCEPT)
}

func (mgr *Manager) handleTradeItems(e *g.InterceptArgs) {
	var offers Offers
	e.Packet.Read(&offers)

	args := &Args{Offers: offers}

	/*
		There is no trade open packet, and we want to detect whether a trade was opened.
		To detect this, we keep track of whether a trade is open or not ourselves.

		Currently, when you close a trade, the server resends a TRADE_ITEMS packet,
		opening a new trade where you cannot offer any items.

		Closing that trade does not cause the server to respond with a TRADE_CLOSE packet.

		This means that if we then trade someone else, we will receive
		consecutive TRADE_ITEMS packets with different traders.

		To detect if a trade was opened with someone else, we are checking
		whether the trader's names have changed.
	*/

	if mgr.Trading {
		for i := range 2 {
			if mgr.Offers[i].Name != offers[i].Name {
				args.Opened = true
				dbg.Printf("detected trade open due to trader name change (%s -> %s)",
					mgr.Offers[i].Name, offers[i].Name)
				break
			}
		}
	} else {
		mgr.Trading = true
		args.Opened = true
	}

	mgr.Offers = offers
	mgr.updated.Dispatch(args)

	dbg.Println("trade updated")
	// TODO: check if this loop gets optimized away when !debug.Enabled
	for _, offer := range offers {
		dbg.Printf("%s: %d item(s) (accepted: %t)", offer.Name, len(offer.Items), offer.Accepted)
	}
}

func (mgr *Manager) handleTradeAccept(e *g.InterceptArgs) {
	if !mgr.Trading {
		return
	}

	s := e.Packet.ReadString()
	fields := strings.SplitN(s, "/", 2)
	if len(fields) != 2 {
		dbg.Printf("WARNING: fields length != 2: %q (%v)", s, fields)
		return
	}

	name := fields[0]
	accepted := fields[1] == "true"

	var offer *Offer
	for i := range 2 {
		if mgr.Offers[i].Name == name {
			offer = &mgr.Offers[i]
			break
		}
	}

	if offer != nil {
		offer.Accepted = accepted
		mgr.accepted.Dispatch(&AcceptArgs{name, accepted})
	} else {
		dbg.Printf("WARNING: failed to find offer for %q", name)
	}
}

func (mgr *Manager) handleTradeCompleted2(e *g.InterceptArgs) {
	if !mgr.Trading {
		return
	}

	mgr.completed.Dispatch(&Args{Offers: mgr.Offers})
	dbg.Printf("trade completed")
}

func (mgr *Manager) handleTradeClose(e *g.InterceptArgs) {
	if !mgr.Trading {
		return
	}

	offers := mgr.Offers
	mgr.Trading = false
	mgr.Offers = Offers{}
	mgr.closed.Dispatch(&Args{Offers: offers})
	dbg.Printf("trade closed")
}
