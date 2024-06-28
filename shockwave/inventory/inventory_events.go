package inventory

import g "xabbo.b7c.io/goearth"

// ItemArgs holds the arguments for inventory events involving a single item.
type ItemArgs struct {
	Item Item
}

// Updated registers an event handler that is invoked when the inventory is updated.
func (mgr *Manager) Updated(handler g.VoidHandler) {
	mgr.updated.Register(handler)
}

// ItemRemoved registers an event handler that is invoked when an item is removed from the inventory.
func (mgr *Manager) ItemRemoved(handler g.EventHandler[ItemArgs]) {
	mgr.itemRemoved.Register(handler)
}
