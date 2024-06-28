package main

import (
	"io"
	"log"

	g "xabbo.b7c.io/goearth"
	"xabbo.b7c.io/goearth/shockwave/inventory"
	"xabbo.b7c.io/goearth/shockwave/room"
	"xabbo.b7c.io/goearth/shockwave/trade"
)

// An extension for testing the shockwave game state managers.

var (
	ext = g.NewExt(g.ExtInfo{
		Title:       "Go Earth",
		Description: "example: shockwave mgrs",
		Author:      "b7",
		Version:     "1.0",
	})
	invMgr   = inventory.NewManager(ext)
	roomMgr  = room.NewManager(ext)
	tradeMgr = trade.NewManager(ext)

	l struct {
		inv, room, trade *log.Logger
	}
)

func init() {
	flags := log.Ltime | log.Ldate | log.Lmicroseconds | log.Lmsgprefix
	l.inv = log.New(io.Discard, "[inventory] ", flags)
	l.room = log.New(io.Discard, "     [room] ", flags)
	l.trade = log.New(io.Discard, "    [trade] ", flags)
}

func main() {
	// Inventory
	invMgr.Updated(onInventoryUpdated)
	invMgr.ItemRemoved(onInventoryItemRemoved)

	// Room
	roomMgr.Entered(onEnteredRoom)
	roomMgr.ObjectsLoaded(onObjectsLoaded)
	roomMgr.ObjectAdded(onObjectAdded)
	roomMgr.ObjectRemoved(onObjectRemoved)
	roomMgr.ItemsAdded(onItemsAdded)
	roomMgr.ItemRemoved(onItemRemoved)
	roomMgr.EntitiesAdded(onEntitiesEntered)
	roomMgr.EntityChat(onEntityChat)
	roomMgr.EntityLeft(onEntityLeft)
	roomMgr.Left(onLeftRoom)

	// Trade
	tradeMgr.Updated(onTradeUpdated)
	tradeMgr.Accepted(onTradeAccepted)
	tradeMgr.Completed(onTradeCompleted)
	tradeMgr.Closed(onTradeClosed)

	ext.Run()
}

// Room

func onEnteredRoom(e *room.Args) {
	if e.Info != nil {
		l.room.Printf("Entered room %q (id:%d) by %s",
			e.Info.Name, e.Info.Id, e.Info.Owner)
	} else {
		l.room.Printf("Entered room (id: %d)", e.Id)
	}
}

func onObjectsLoaded(e *room.ObjectsArgs) {
	l.room.Printf("Added %d floor item(s)", len(e.Objects))
}

func onObjectAdded(e *room.ObjectArgs) {
	l.room.Printf("Added floor item %s (id: %s)", e.Object.Class, e.Object.Id)
}

func onObjectRemoved(e *room.ObjectArgs) {
	l.room.Printf("Removed floor item %s (id: %s)", e.Object.Class, e.Object.Id)
}

func onItemsAdded(e *room.ItemsArgs) {
	l.room.Printf("Added %d wall item(s)", len(e.Items))
}

func onItemRemoved(e *room.ItemArgs) {
	l.room.Printf("Removed wall item %s (id: %d)", e.Item.Class, e.Item.Id)
}

func onEntitiesEntered(e *room.EntitiesArgs) {
	if len(e.Entities) > 0 {
		if e.Entered && len(e.Entities) == 1 {
			l.room.Printf("%s entered the room", e.Entities[0].Name)
		} else {
			l.room.Printf("Added %d entities", len(e.Entities))
		}
	}
}

func onEntityChat(e *room.EntityChatArgs) {
	symbol := ""
	switch e.Type {
	case room.Talk:
		symbol = "-"
	case room.Whisper:
		symbol = "*"
	case room.Shout:
		symbol = "!"
	}
	l.room.Printf("[%s] %s: %s", symbol, e.Entity.Name, e.Message)
}

func onEntityLeft(e *room.EntityArgs) {
	l.room.Printf("%s left the room", e.Entity.Name)
}

func onLeftRoom(e *room.Args) {
	l.room.Println("Left room")
}

// Inventory

func onInventoryUpdated() {
	l.inv.Printf("Inventory updated (items: %d)", len(invMgr.Items()))
}

func onInventoryItemRemoved(e *inventory.ItemArgs) {
	l.inv.Printf("Inventory item %s removed (id: %d)", e.Item.Class, e.Item.ItemId)
}

// Trade

func onTradeUpdated(e *trade.Args) {
	if e.Opened {
		l.trade.Printf("Trade opened (%s -> %s)",
			e.Offers.Trader().Name, e.Offers.Tradee().Name)
	} else {
		l.trade.Printf("Trade updated")
		for _, offer := range e.Offers {
			l.trade.Printf("  %s: %d item(s) (accepted: %t)", offer.Name, len(offer.Items), offer.Accepted)
		}
	}
}

func onTradeAccepted(e *trade.AcceptArgs) {
	if e.Accepted {
		l.trade.Printf("Trade accepted by %s", e.Name)
	} else {
		l.trade.Printf("Trade unaccepted by %s", e.Name)
	}
}

func onTradeCompleted(e *trade.Args) {
	l.trade.Printf("Trade completed (exchanged %d <-> %d item(s))",
		len(e.Offers[0].Items), len(e.Offers[1].Items))
}

func onTradeClosed(e *trade.Args) {
	l.trade.Printf("Trade closed")
}
