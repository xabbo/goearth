package room

import g "xabbo.b7c.io/goearth"

// Args hold the arguments for room events.
type Args struct {
	Id   int
	Info *Info
}

// ObjectArgs holds the arguments for floor item events involving a single item.
type ObjectArgs struct {
	Object Object
}

// ObjectArgs holds the arguments for floor item events involving a list of items.
type ObjectsArgs struct {
	Objects []Object
}

// ItemArgs holds the arguments for wall item events involing a single item.
type ItemArgs struct {
	Item Item
}

// ItemsArgs holds the arguments for wall item events involing a list of items.
type ItemsArgs struct {
	Items []Item
}

// EntityArgs holds the arguments for events involving a single entity.
type EntityArgs struct {
	Entity Entity
}

// EntitiesArgs holds the arguments for events involving a list of entities.
type EntitiesArgs struct {
	Entered  bool
	Entities []Entity
}

// EntityChat holds the arguments for chat events.
type EntityChatArgs struct {
	EntityArgs
	Type    ChatType
	Message string
}

// Entered registers an event handler that is invoked when the user enters a room.
func (mgr *Manager) Entered(handler g.EventHandler[Args]) {
	mgr.entered.Register(handler)
}

// ObjectsLoaded registers an event handler that is invoked when floor items are loaded.
func (mgr *Manager) ObjectsLoaded(handler g.EventHandler[ObjectsArgs]) {
	mgr.objectsLoaded.Register(handler)
}

// ObjectAdded registers an event handler that is invoked when a floor item is added to the room.
func (mgr *Manager) ObjectAdded(handler g.EventHandler[ObjectArgs]) {
	mgr.objectAdded.Register(handler)
}

// ObjectRemoved registers an event handler that is invoked when a floor item is removed from the room.
func (mgr *Manager) ObjectRemoved(handler g.EventHandler[ObjectArgs]) {
	mgr.objectRemoved.Register(handler)
}

// ItemsAdded registers an event handler that is invoked when wall items are loaded or added to the room.
func (mgr *Manager) ItemsAdded(handler g.EventHandler[ItemsArgs]) {
	mgr.itemsAdded.Register(handler)
}

// ItemRemoved registers an event handler that is invoked when an item is removed from the room.
func (mgr *Manager) ItemRemoved(handler g.EventHandler[ItemArgs]) {
	mgr.itemRemoved.Register(handler)
}

// EntitiesAdded registers an event handler that is invoked when entities are loaded or enter the room.
// The Entered flag on the EntitiesArgs indicates whether the entity entered the room.
// If not, the entities were already in the room and are being loaded.
func (mgr *Manager) EntitiesAdded(handler g.EventHandler[EntitiesArgs]) {
	mgr.entitiesAdded.Register(handler)
}

// EntityChat registers an event handler that is invoked when an entity sends a chat message.
func (mgr *Manager) EntityChat(handler g.EventHandler[EntityChatArgs]) {
	mgr.entityChat.Register(handler)
}

// EntityLeft registers an event handler that is invoked when an entity leaves the room.
func (mgr *Manager) EntityLeft(handler g.EventHandler[EntityArgs]) {
	mgr.entityLeft.Register(handler)
}

// Left registers an event handler that is invoked when the user leaves the room.
func (mgr *Manager) Left(handler g.EventHandler[Args]) {
	mgr.left.Register(handler)
}
