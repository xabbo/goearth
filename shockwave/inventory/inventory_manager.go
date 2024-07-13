package inventory

import (
	"sync"

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

	mtxItems *sync.RWMutex
	items    map[int]Item
}

// NewManager creates a new inventory Manager using the provided extension.
func NewManager(ext *g.Ext) *Manager {
	mgr := &Manager{
		ext:      ext,
		mtxItems: &sync.RWMutex{},
		items:    map[int]Item{},
	}
	ext.Intercept(in.STRIPINFO_2).With(mgr.handleStripInfo2)
	ext.Intercept(in.REMOVESTRIPITEM).With(mgr.handleRemoveStripItem)
	return mgr
}

// Item gets the item with the specified ID.
func (mgr *Manager) Item(id int) *Item {
	mgr.mtxItems.RLock()
	defer mgr.mtxItems.RUnlock()
	if item, ok := mgr.items[id]; ok {
		return &item
	} else {
		return nil
	}
}

// Items iterates over all inventory items.
func (mgr *Manager) Items(yield func(item Item) bool) {
	mgr.mtxItems.RLock()
	for _, item := range mgr.items {
		mgr.mtxItems.RUnlock()
		if !yield(item) {
			return
		}
		mgr.mtxItems.RLock()
	}
	mgr.mtxItems.RUnlock()
}

// ItemCount returns the number of items in the inventory.
func (mgr *Manager) ItemCount() int {
	return len(mgr.items)
}

// RequestItems sends a request to retrieve the user's inventory.
func (mgr *Manager) RequestItems() {
	mgr.ext.Send(out.GETSTRIP)
}

func (mgr *Manager) loadItems(items []Item) {
	mgr.mtxItems.Lock()
	defer mgr.mtxItems.Unlock()

	clear(mgr.items)
	mgr.items = map[int]Item{}
	for _, item := range items {
		if _, exists := mgr.items[item.Id]; exists {
			dbg.Printf("WARNING: duplicate item (ID: %d)", item.Id)
		}
		mgr.items[item.ItemId] = item
	}

	dbg.Printf("loaded %d items", len(items))
}

func (mgr *Manager) removeItem(id int) (item Item, ok bool) {
	mgr.mtxItems.Lock()
	defer mgr.mtxItems.Unlock()

	if item, ok = mgr.items[id]; ok {
		delete(mgr.items, id)
		dbg.Printf("removed item (ID: %d)", id)
	} else {
		dbg.Printf("failed to find item to remove (ID: %d)", id)
	}

	return
}

// handlers

func (mgr *Manager) handleStripInfo2(e *g.Intercept) {
	var inv Inventory
	e.Packet.Read(&inv)

	mgr.loadItems(inv.Items)
	mgr.updated.Dispatch()
}

func (mgr *Manager) handleRemoveStripItem(e *g.Intercept) {
	itemId := e.Packet.ReadInt()
	if item, ok := mgr.removeItem(itemId); ok {
		mgr.itemRemoved.Dispatch(ItemArgs{item})
	}
}
