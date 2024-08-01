package room

import (
	"strconv"
	"strings"
	"sync"

	g "xabbo.b7c.io/goearth"
	"xabbo.b7c.io/goearth/internal/debug"
	"xabbo.b7c.io/goearth/shockwave/in"
)

var dbg = debug.NewLogger("[room]")

type Manager struct {
	ext *g.Ext

	entered       g.Event[Args]
	rightsUpdated g.VoidEvent
	objectsLoaded g.Event[ObjectsArgs]
	objectAdded   g.Event[ObjectArgs]
	objectUpdated g.Event[ObjectUpdateArgs]
	objectRemoved g.Event[ObjectArgs]
	slide         g.Event[SlideArgs]
	itemsLoaded   g.Event[ItemsArgs]
	itemAdded     g.Event[ItemArgs]
	itemUpdated   g.Event[ItemUpdateArgs]
	itemRemoved   g.Event[ItemArgs]
	entitiesAdded g.Event[EntitiesArgs]
	entityUpdated g.Event[EntityUpdateArgs]
	entityChat    g.Event[EntityChatArgs]
	entityLeft    g.Event[EntityArgs]
	left          g.Event[Args]

	mtxCache  *sync.RWMutex
	infoCache map[int]Info

	usersPacketCount int

	mtxRoom   *sync.RWMutex
	isInRoom  bool
	roomId    int
	roomModel string
	roomInfo  *Info
	isOwner   bool // IsOwner indicates whether the user is the owner of the current room.
	hasRights bool // HasRights indicates whether the user has rights in the current room.
	heightmap []string

	mtxObjs  *sync.RWMutex
	objects  map[int]Object
	mtxItems *sync.RWMutex
	items    map[int]Item
	mtxEnts  *sync.RWMutex
	entities map[int]Entity
}

func NewManager(ext *g.Ext) *Manager {
	mgr := &Manager{
		ext:       ext,
		mtxRoom:   &sync.RWMutex{},
		mtxCache:  &sync.RWMutex{},
		infoCache: map[int]Info{},
		mtxObjs:   &sync.RWMutex{},
		objects:   map[int]Object{},
		mtxItems:  &sync.RWMutex{},
		items:     map[int]Item{},
		mtxEnts:   &sync.RWMutex{},
		entities:  map[int]Entity{},
	}
	ext.Intercept(in.FLATINFO).With(mgr.handleFlatInfo)
	ext.Intercept(in.OPC_OK).With(mgr.handleOpcOk)
	ext.Intercept(in.ROOM_READY).With(mgr.handleRoomReady)
	ext.Intercept(in.ROOM_RIGHTS, in.ROOM_RIGHTS_2, in.ROOM_RIGHTS_3).With(mgr.handleRoomRights)
	ext.Intercept(in.HEIGHTMAP).With(mgr.handleHeightmap)
	ext.Intercept(in.ACTIVEOBJECTS).With(mgr.handleActiveObjects)
	ext.Intercept(in.ACTIVEOBJECT_ADD).With(mgr.handleActiveObjectAdd)
	ext.Intercept(in.ACTIVEOBJECT_UPDATE).With(mgr.handleActiveObjectUpdate)
	ext.Intercept(in.ACTIVEOBJECT_REMOVE).With(mgr.handleActiveObjectRemove)
	ext.Intercept(in.SLIDEOBJECTBUNDLE).With(mgr.handleSlideObjectBundle)
	ext.Intercept(in.ITEMS).With(mgr.handleItems)
	ext.Intercept(in.ITEMS_2, in.UPDATEITEM).With(mgr.handleAddOrUpdateItem)
	ext.Intercept(in.REMOVEITEM).With(mgr.handleRemoveItem)
	ext.Intercept(in.USERS).With(mgr.handleUsers)
	ext.Intercept(in.STATUS).With(mgr.handleStatus)
	ext.Intercept(in.CHAT, in.CHAT_2, in.CHAT_3).With(mgr.handleChat)
	ext.Intercept(in.LOGOUT).With(mgr.handleLogout)
	ext.Intercept(in.CLC).With(mgr.handleClc)
	return mgr
}

func (mgr *Manager) IsInRoom() bool {
	return mgr.isInRoom
}

func (mgr *Manager) Id() int {
	return mgr.roomId
}

func (mgr *Manager) Model() string {
	return mgr.roomModel
}

func (mgr *Manager) Info() *Info {
	return mgr.roomInfo
}

func (mgr *Manager) IsOwner() bool {
	return mgr.isOwner
}

func (mgr *Manager) HasRights() bool {
	return mgr.hasRights
}

func (mgr *Manager) Heightmap() []string {
	return mgr.heightmap
}

// Object gets a floor item in the room by its ID.
// It returns nil if the object was not found.
func (mgr *Manager) Object(id int) *Object {
	mgr.mtxObjs.RLock()
	defer mgr.mtxObjs.RUnlock()

	if obj, ok := mgr.objects[id]; ok {
		return &obj
	} else {
		return nil
	}
}

// Objects iterates over all floor items currently in the room.
func (mgr *Manager) Objects(yield func(obj Object) bool) {
	mgr.mtxObjs.RLock()
	for _, obj := range mgr.objects {
		mgr.mtxObjs.RUnlock()
		if !yield(obj) {
			return
		}
		mgr.mtxObjs.RLock()
	}
	mgr.mtxObjs.RUnlock()
}

// ObjectCount returns the number of objects in the room.
func (mgr *Manager) ObjectCount() int {
	return len(mgr.objects)
}

// Item gets a wall item in the room by its ID.
// It returns nil if the item was not found.
func (mgr *Manager) Item(id int) *Item {
	mgr.mtxItems.RLock()
	defer mgr.mtxItems.RUnlock()

	if item, ok := mgr.items[id]; ok {
		return &item
	} else {
		return nil
	}
}

// Items iterates over all wall items currently in the room.
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

// ItemCount returns the number of items in the room.
func (mgr *Manager) ItemCount() int {
	return len(mgr.items)
}

// Entity gets an entity in the room by its index.
// It returns nil if the entity was not found.
func (mgr *Manager) Entity(id int) *Entity {
	mgr.mtxEnts.RLock()
	defer mgr.mtxEnts.RUnlock()

	if ent, ok := mgr.entities[id]; ok {
		return &ent
	} else {
		return nil
	}
}

// EntityByName gets the entity with the specified name.
// Names are case-insensitive.
// Returns nil if it does not exist.
func (mgr *Manager) EntityByName(name string) *Entity {
	// TODO add name -> entity map
	mgr.mtxEnts.RLock()
	defer mgr.mtxEnts.RUnlock()

	for _, ent := range mgr.entities {
		if strings.EqualFold(ent.Name, name) {
			return &ent
		}
	}
	return nil
}

// Entities iterates over all entities currently in the room.
func (mgr *Manager) Entities(yield func(ent Entity) bool) {
	mgr.mtxEnts.RLock()
	for _, ent := range mgr.entities {
		mgr.mtxEnts.RUnlock()
		if !yield(ent) {
			break
		}
		mgr.mtxEnts.RLock()
	}
	mgr.mtxEnts.RUnlock()
}

// EntityCount returns the number of entities in the room.
func (mgr *Manager) EntityCount() int {
	return len(mgr.entities)
}

func (mgr *Manager) enterRoom(model string, id int) (info Info, ok bool) {
	mgr.mtxRoom.Lock()
	defer mgr.mtxRoom.Unlock()

	mgr.roomModel = model
	mgr.roomId = id
	mgr.isInRoom = true

	info, ok = mgr.infoCache[id]
	if ok {
		mgr.roomInfo = &info
	} else {
		mgr.roomInfo = nil
	}

	return
}

func (mgr *Manager) leaveRoom() {
	if mgr.isInRoom {
		mgr.mtxRoom.Lock()
		defer mgr.mtxRoom.Unlock()

		id := mgr.roomId
		info := mgr.roomInfo

		mgr.usersPacketCount = 0

		mgr.isInRoom = false
		mgr.roomModel = ""
		mgr.roomId = 0
		mgr.roomInfo = nil
		mgr.isOwner = false
		mgr.hasRights = false
		mgr.heightmap = []string{}
		mgr.clearObjects()
		mgr.clearItems()
		mgr.clearEntities()

		mgr.left.Dispatch(Args{Id: id, Info: info})

		dbg.Printf("left room")
	}
}

func (mgr *Manager) updateCache(info Info) {
	mgr.mtxCache.Lock()
	defer mgr.mtxCache.Unlock()
	mgr.infoCache[info.Id] = info
}

func (mgr *Manager) addObjects(load bool, objs []Object) {
	mgr.mtxObjs.Lock()
	defer mgr.mtxObjs.Unlock()

	for _, object := range objs {
		mgr.objects[object.Id] = object
	}

	if load {
		dbg.Printf("loaded %d objects", len(objs))
	} else {
		dbg.Printf("added object %s (ID: %d)", objs[0].Class, objs[0].Id)
	}
}

func (mgr *Manager) updateObject(obj Object) (pre Object, ok bool) {
	mgr.mtxObjs.Lock()
	defer mgr.mtxObjs.Unlock()

	if pre, ok = mgr.objects[obj.Id]; ok {
		mgr.objects[obj.Id] = obj
		dbg.Printf("updated object %s (ID: %d)", obj.Class, obj.Id)
	} else {
		dbg.Printf("WARNING: failed to find object to update (ID: %d)", obj.Id)
	}

	return
}

func (mgr *Manager) processSlideObjectBundle(bundle SlideObjectBundle) SlideArgs {
	mgr.mtxObjs.Lock()
	defer mgr.mtxObjs.Unlock()
	mgr.mtxEnts.Lock()
	defer mgr.mtxEnts.Unlock()

	var pSource *Object
	if bundle.RollerId != 0 {
		if source, ok := mgr.objects[bundle.RollerId]; ok {
			pSource = &source
		} else {
			dbg.Printf("failed to find source (ID: %d)", bundle.RollerId)
		}
	}

	args := SlideArgs{
		From:          bundle.From,
		To:            bundle.To,
		Source:        pSource,
		SlideMoveType: bundle.SlideMoveType,
	}

	for _, bundleObj := range bundle.Objects {
		obj, ok := mgr.objects[bundleObj.Id]
		if ok {
			obj.X = args.To.X
			obj.Y = args.To.Y
			obj.Z = bundleObj.ToZ
			mgr.objects[obj.Id] = obj
			args.ObjectSlides = append(args.ObjectSlides, SlideObjectArgs{
				Object: obj,
				From: Tile{
					X: bundle.From.X,
					Y: bundle.From.Y,
					Z: bundleObj.FromZ,
				},
				To: Tile{
					X: bundle.To.X,
					Y: bundle.To.Y,
					Z: bundleObj.ToZ,
				},
			})
		} else {
			dbg.Printf("failed to find object (ID: %d)", bundleObj.Id)
		}
	}

	if bundle.SlideMoveType != SlideMoveTypeNone {
		if ent, ok := mgr.entities[bundle.Entity.Id]; ok {
			ent.X = bundle.To.X
			ent.Y = bundle.To.Y
			ent.Z = bundle.Entity.ToZ
			mgr.entities[ent.Index] = ent
			args.EntitySlide = &SlideEntityArgs{
				Entity: ent,
				From: Tile{
					X: bundle.From.X,
					Y: bundle.From.Y,
					Z: bundle.Entity.FromZ,
				},
				To: Tile{
					X: bundle.To.X,
					Y: bundle.To.Y,
					Z: bundle.Entity.ToZ,
				},
			}
		} else {
			dbg.Printf("failed to find entity (ID: %d)", bundle.Entity.Id)
		}
	}

	dbg.Printf("processed slide object bundle (%d objects, with entity: %t)", len(args.ObjectSlides), args.EntitySlide != nil)
	return args
}

func (mgr *Manager) removeObject(id int) (obj Object, ok bool) {
	mgr.mtxObjs.Lock()
	defer mgr.mtxObjs.Unlock()

	if obj, ok = mgr.objects[id]; ok {
		delete(mgr.objects, id)
		dbg.Printf("removed object (ID: %d)", id)
	} else {
		dbg.Printf("WARNING: failed to remove object (ID: %d)", id)
	}

	return
}

func (mgr *Manager) clearObjects() {
	mgr.mtxObjs.Lock()
	defer mgr.mtxObjs.Unlock()
	clear(mgr.objects)
	mgr.objects = map[int]Object{}
}

func (mgr *Manager) addItems(load bool, items []Item) {
	mgr.mtxItems.Lock()
	defer mgr.mtxItems.Unlock()

	for _, item := range items {
		// TODO: check if this loop gets optimized away when !debug.Enabled
		if _, exists := mgr.items[item.Id]; exists {
			dbg.Printf("WARNING: duplicate item (ID: %d)", item.Id)
		}
		mgr.items[item.Id] = item
	}

	if load {
		dbg.Printf("loaded %d items", len(items))
	} else {
		dbg.Printf("added item %s (ID: %d)", items[0].Class, items[0].Id)
	}
}

func (mgr *Manager) updateItem(item Item) (pre Item, ok bool) {
	mgr.mtxItems.Lock()
	defer mgr.mtxItems.Unlock()

	if pre, ok = mgr.items[item.Id]; ok {
		mgr.items[item.Id] = item
		dbg.Printf("updated item %s (ID: %d)", item.Class, item.Id)
	} else {
		dbg.Printf("WARNING: failed to find item to update (ID: %d)", item.Id)
	}

	return
}

func (mgr *Manager) removeItem(id int) (item Item, ok bool) {
	mgr.mtxItems.Lock()
	defer mgr.mtxItems.Unlock()

	if item, ok = mgr.items[id]; ok {
		delete(mgr.items, item.Id)
		dbg.Printf("removed item %s (ID: %d)", item.Class, item.Id)
	} else {
		dbg.Printf("WARNING: failed to find item to remove (ID: %d)", item.Id)
	}

	return
}

func (mgr *Manager) clearItems() {
	mgr.mtxItems.Lock()
	defer mgr.mtxItems.Unlock()
	clear(mgr.items)
	mgr.items = map[int]Item{}
}

func (mgr *Manager) addEntities(ents []Entity) {
	mgr.mtxEnts.Lock()
	defer mgr.mtxEnts.Unlock()

	for _, entity := range ents {
		// TODO: check if this branch gets optimized away when !debug.Enabled
		if _, exists := mgr.entities[entity.Index]; exists {
			dbg.Printf("WARNING: duplicate entity index: %d", entity.Index)
		}
		mgr.entities[entity.Index] = entity
	}
}

func (mgr *Manager) updateEntities(statuses []EntityStatus) []EntityUpdateArgs {
	mgr.mtxEnts.Lock()
	defer mgr.mtxEnts.Unlock()

	updates := make([]EntityUpdateArgs, 0, len(statuses))

	for _, status := range statuses {
		pre, ok := mgr.entities[status.Index]
		if ok {
			cur := pre
			cur.Tile = status.Tile
			cur.Action = status.Action
			mgr.entities[status.Index] = cur
		} else {
			dbg.Printf("WARNING: failed to find entity to update (index: %d)", status.Index)
		}
	}

	return updates
}

func (mgr *Manager) removeEntity(index int) (ent Entity, ok bool) {
	mgr.mtxEnts.Lock()
	defer mgr.mtxEnts.Unlock()

	if ent, ok = mgr.entities[index]; ok {
		delete(mgr.entities, index)
		dbg.Printf("removed entity %q (index: %d)", ent.Name, ent.Index)
	} else {
		dbg.Printf("WARNING: failed to find entity to remove (index: %d)", index)
	}

	return
}

func (mgr *Manager) clearEntities() {
	mgr.mtxEnts.Lock()
	defer mgr.mtxEnts.Unlock()
	clear(mgr.entities)
	mgr.entities = map[int]Entity{}
}

// handlers

func (mgr *Manager) handleFlatInfo(e *g.Intercept) {
	var info Info
	e.Packet.Read(&info)

	mgr.updateCache(info)

	dbg.Printf("cached room info (ID: %d)", info.Id)
}

func (mgr *Manager) handleOpcOk(e *g.Intercept) {
	mgr.leaveRoom()
}

func (mgr *Manager) handleRoomReady(e *g.Intercept) {
	if mgr.isInRoom {
		dbg.Printf("WARNING: already in room")
	}

	s := e.Packet.ReadString()
	fields := strings.Fields(s)
	if len(fields) != 2 {
		dbg.Printf("WARNING: string fields length != 2: %q (%v)", s, fields)
		return
	}

	model := fields[0]
	roomId, err := strconv.Atoi(fields[1])
	if err != nil {
		dbg.Printf("WARNING: room ID is not an integer: %s", fields[1])
		return
	}

	mgr.enterRoom(model, roomId)

	if mgr.roomInfo != nil {
		mgr.entered.Dispatch(Args{Id: roomId, Info: mgr.roomInfo})
		dbg.Printf("entered room %q by %s (ID: %d)", mgr.roomInfo.Name, mgr.roomInfo.Owner, mgr.roomInfo.Id)
	} else {
		mgr.entered.Dispatch(Args{Id: roomId})
		dbg.Println("WARNING: failed to get room info from cache")
		dbg.Printf("entered room (ID: %d)", roomId)
	}
}

func (mgr *Manager) handleRoomRights(e *g.Intercept) {
	if !mgr.isInRoom {
		return
	}

	switch {
	case e.Is(in.ROOM_RIGHTS):
		mgr.hasRights = true
		mgr.rightsUpdated.Dispatch()
	case e.Is(in.ROOM_RIGHTS_2):
		mgr.hasRights = false
		mgr.rightsUpdated.Dispatch()
	case e.Is(in.ROOM_RIGHTS_3):
		mgr.isOwner = true
	}
}

func (mgr *Manager) handleHeightmap(e *g.Intercept) {
	if !mgr.isInRoom {
		return
	}

	mgr.heightmap = strings.Split(e.Packet.ReadString(), "\r")

	if debug.Enabled {
		if len(mgr.heightmap) > 0 {
			dbg.Printf("received heightmap (%dx%d)", len(mgr.heightmap[0]), len(mgr.heightmap))
		} else {
			dbg.Println("WARNING: empty heightmap")
		}
	}
}

func (mgr *Manager) handleActiveObjects(e *g.Intercept) {
	if !mgr.isInRoom {
		return
	}

	var objects []Object
	e.Packet.Read(&objects)

	mgr.addObjects(true, objects)

	mgr.objectsLoaded.Dispatch(ObjectsArgs{Objects: objects})
}

func (mgr *Manager) handleActiveObjectAdd(e *g.Intercept) {
	if !mgr.isInRoom {
		return
	}

	var object Object
	e.Packet.Read(&object)

	mgr.addObjects(false, []Object{object})

	mgr.objectAdded.Dispatch(ObjectArgs{Object: object})
}

func (mgr *Manager) handleActiveObjectUpdate(e *g.Intercept) {
	if !mgr.isInRoom {
		return
	}

	var cur Object
	e.Packet.Read(&cur)

	if pre, ok := mgr.updateObject(cur); ok {
		mgr.objectUpdated.Dispatch(ObjectUpdateArgs{Pre: pre, Object: cur})
	}
}

func (mgr *Manager) handleActiveObjectRemove(e *g.Intercept) {
	if !mgr.isInRoom {
		return
	}

	var obj Object
	e.Packet.Read(&obj)

	if obj, ok := mgr.removeObject(obj.Id); ok {
		mgr.objectRemoved.Dispatch(ObjectArgs{Object: obj})
	}
}

func (mgr *Manager) handleSlideObjectBundle(e *g.Intercept) {
	if !mgr.isInRoom {
		return
	}

	var bundle SlideObjectBundle
	e.Packet.Read(&bundle)

	args := mgr.processSlideObjectBundle(bundle)

	mgr.slide.Dispatch(args)
}

func (mgr *Manager) handleItems(e *g.Intercept) {
	if !mgr.isInRoom {
		return
	}

	var items Items
	e.Packet.Read(&items)

	mgr.addItems(true, items)

	mgr.itemsLoaded.Dispatch(ItemsArgs{Items: items})
}

func (mgr *Manager) handleAddOrUpdateItem(e *g.Intercept) {
	if !mgr.isInRoom {
		return
	}

	var item Item
	e.Packet.Read(&item)

	if e.Is(in.ITEMS_2) {
		mgr.addItems(false, []Item{item})
		mgr.itemAdded.Dispatch(ItemArgs{Item: item})
	} else {
		if pre, ok := mgr.updateItem(item); ok {
			mgr.itemUpdated.Dispatch(ItemUpdateArgs{Pre: pre, Item: item})
		}
	}
}

func (mgr *Manager) handleRemoveItem(e *g.Intercept) {
	if !mgr.isInRoom {
		return
	}

	strId := e.Packet.ReadString()
	id, err := strconv.Atoi(strId)
	if err != nil {
		dbg.Printf("WARNING: invalid item id: %s", strId)
		return
	}

	if item, ok := mgr.removeItem(id); ok {
		mgr.itemRemoved.Dispatch(ItemArgs{Item: item})
	}
}

func (mgr *Manager) handleUsers(e *g.Intercept) {
	if !mgr.isInRoom {
		return
	}

	var ents []Entity
	e.Packet.Read(&ents)

	mgr.addEntities(ents)

	if mgr.usersPacketCount < 3 {
		mgr.usersPacketCount++
	}

	mgr.entitiesAdded.Dispatch(EntitiesArgs{
		Entered:  mgr.usersPacketCount >= 3,
		Entities: ents,
	})

	dbg.Printf("added %d entities", len(ents))
}

func (mgr *Manager) handleStatus(e *g.Intercept) {
	if !mgr.isInRoom {
		return
	}

	var statuses []EntityStatus
	e.Packet.Read(&statuses)

	updates := mgr.updateEntities(statuses)

	for _, update := range updates {
		mgr.entityUpdated.Dispatch(update)
	}
}

func (mgr *Manager) handleChat(e *g.Intercept) {
	if !mgr.isInRoom {
		return
	}

	index := e.Packet.ReadInt()
	msg := e.Packet.ReadString()
	var chatType ChatType
	if e.Is(in.CHAT) {
		chatType = Talk
	} else if e.Is(in.CHAT_2) {
		chatType = Whisper
	} else if e.Is(in.CHAT_3) {
		chatType = Shout
	} else {
		dbg.Printf("WARNING: unknown chat header: %q", e.Name())
	}

	if entity, ok := mgr.entities[index]; ok {
		mgr.entityChat.Dispatch(EntityChatArgs{
			EntityArgs: EntityArgs{Entity: entity},
			Type:       chatType,
			Message:    msg,
		})
		var indicator string
		switch chatType {
		case Talk:
			indicator = "[-]"
		case Shout:
			indicator = "[!]"
		case Whisper:
			indicator = "[*]"
		}
		dbg.Printf("%s %s: %s", indicator, entity.Name, msg)
	} else {
		dbg.Printf("WARNING: failed to find entity (index: %d)", index)
	}
}

func (mgr *Manager) handleLogout(e *g.Intercept) {
	if !mgr.isInRoom {
		return
	}

	s := e.Packet.ReadString()
	index, err := strconv.Atoi(s)
	if err != nil {
		dbg.Printf("WARNING: invalid index: %q", s)
		return
	}

	if entity, ok := mgr.removeEntity(index); ok {
		mgr.entityLeft.Dispatch(EntityArgs{Entity: entity})
	}
}

func (mgr *Manager) handleClc(e *g.Intercept) {
	mgr.leaveRoom()
}
