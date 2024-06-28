package inventory

import (
	g "xabbo.b7c.io/goearth"
	"xabbo.b7c.io/goearth/internal/debug"
	"xabbo.b7c.io/goearth/shockwave/in"
	"xabbo.b7c.io/goearth/shockwave/out"
)

var dbg = debug.NewLogger("[inventory]")

// Manager tracks the state of the inventory.
type Manager struct {
	ext         *g.Ext
	updated     g.VoidEvent
	itemRemoved g.Event[ItemArgs]
	items       map[int]Item
}

// NewManager creates a new inventory Manager using the provided extension.
func NewManager(ext *g.Ext) *Manager {
	mgr := &Manager{
		ext:   ext,
		items: map[int]Item{},
	}
	ext.Intercept(in.STRIPINFO_2).With(mgr.handleStripInfo2)
	ext.Intercept(in.REMOVESTRIPITEM).With(mgr.handleRemoveStripItem)
	return mgr
}

// Items returns the inventory items.
func (mgr *Manager) Items() map[int]Item {
	return mgr.items
}

// Update sends a request to retrieve the user's inventory.
func (mgr *Manager) Update() {
	mgr.ext.Send(out.GETSTRIP)
}

func (mgr *Manager) handleStripInfo2(e *g.Intercept) {
	var inv Inventory
	e.Packet.Read(&inv)

	clear(mgr.items)
	for _, item := range inv.Items {
		if debug.Enabled {
			if _, exists := mgr.items[item.Id]; exists {
				dbg.Printf("WARNING: duplicate item (ID: %d)", item.Id)
			}
		}
		mgr.items[item.ItemId] = item
	}
	mgr.updated.Dispatch()

	dbg.Printf("added %d items", len(inv.Items))
}

func (mgr *Manager) handleRemoveStripItem(e *g.Intercept) {
	itemId := e.Packet.ReadInt()
	if item, ok := mgr.items[itemId]; ok {
		delete(mgr.items, itemId)
		mgr.itemRemoved.Dispatch(&ItemArgs{item})
		dbg.Printf("removed item (ID: %d)", itemId)
	} else {
		dbg.Printf("failed to remove item (ID: %d)", itemId)
	}
}
