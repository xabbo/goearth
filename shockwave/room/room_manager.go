package room

import (
	"strconv"
	"strings"

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
	itemRemoved   g.Event[ItemArgs]
	entitiesAdded g.Event[EntitiesArgs]
	entityUpdated g.Event[EntityUpdateArgs]
	entityChat    g.Event[EntityChatArgs]
	entityLeft    g.Event[EntityArgs]
	left          g.Event[Args]

	infoCache map[int]Info

	usersPacketCount int

	IsInRoom  bool
	RoomModel string
	RoomId    int
	RoomInfo  *Info
	IsOwner   bool // IsOwner indicates whether the user is the owner of the current room.
	HasRights bool // HasRights indicates whether the user has rights in the current room.
	Heightmap []string

	Objects  map[int]Object
	Items    map[int]Item
	Entities map[int]Entity
}

func NewManager(ext *g.Ext) *Manager {
	mgr := &Manager{
		ext:       ext,
		infoCache: map[int]Info{},
		Objects:   map[int]Object{},
		Items:     map[int]Item{},
		Entities:  map[int]Entity{},
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
	ext.Intercept(in.ITEMS_2).With(mgr.handleItems2)
	ext.Intercept(in.REMOVEITEM).With(mgr.handleRemoveItem)
	ext.Intercept(in.USERS).With(mgr.handleUsers)
	ext.Intercept(in.STATUS).With(mgr.handleStatus)
	ext.Intercept(in.CHAT, in.CHAT_2, in.CHAT_3).With(mgr.handleChat)
	ext.Intercept(in.LOGOUT).With(mgr.handleLogout)
	ext.Intercept(in.CLC).With(mgr.handleClc)
	return mgr
}

func (mgr *Manager) leaveRoom() {
	if mgr.IsInRoom {
		id := mgr.RoomId
		info := mgr.RoomInfo

		mgr.usersPacketCount = 0

		mgr.IsInRoom = false
		mgr.RoomModel = ""
		mgr.RoomId = 0
		mgr.RoomInfo = info
		mgr.IsOwner = false
		mgr.HasRights = false
		mgr.Heightmap = []string{}
		clear(mgr.Objects)
		clear(mgr.Items)
		clear(mgr.Entities)

		mgr.left.Dispatch(Args{Id: id, Info: info})

		dbg.Printf("left room")
	}
}

// EntityByName gets the entity with the specified name.
// Names are case-insensitive.
// Returns nil if it does not exist.
func (mgr *Manager) EntityByName(name string) *Entity {
	// TODO add maps
	for _, ent := range mgr.Entities {
		if strings.EqualFold(ent.Name, name) {
			return &ent
		}
	}
	return nil
}

// handlers

func (mgr *Manager) handleFlatInfo(e *g.Intercept) {
	var info Info
	e.Packet.Read(&info)

	mgr.infoCache[info.Id] = info

	dbg.Printf("cached room info (ID: %d)", info.Id)
}

func (mgr *Manager) handleOpcOk(e *g.Intercept) {
	mgr.leaveRoom()
}

func (mgr *Manager) handleRoomReady(e *g.Intercept) {
	if mgr.IsInRoom {
		dbg.Printf("WARNING: already in room")
	}

	s := e.Packet.ReadString()
	fields := strings.Fields(s)
	if len(fields) != 2 {
		dbg.Printf("WARNING: string fields length != 2: %q (%v)", s, fields)
		return
	}

	roomId, err := strconv.Atoi(fields[1])
	if err != nil {
		dbg.Printf("WARNING: room ID is not an integer: %s", fields[1])
		return
	}

	mgr.RoomModel = fields[0]
	mgr.RoomId = roomId
	mgr.IsInRoom = true

	if info, ok := mgr.infoCache[roomId]; ok {
		mgr.entered.Dispatch(Args{Id: roomId, Info: &info})
		dbg.Printf("entered room %q by %s (ID: %d)", info.Name, info.Owner, info.Id)
	} else {
		mgr.entered.Dispatch(Args{Id: roomId})
		dbg.Println("WARNING: failed to get room info from cache")
		dbg.Printf("entered room (ID: %d)", roomId)
	}
}

func (mgr *Manager) handleRoomRights(e *g.Intercept) {
	if !mgr.IsInRoom {
		return
	}

	switch {
	case e.Is(in.ROOM_RIGHTS):
		mgr.HasRights = true
		mgr.rightsUpdated.Dispatch()
	case e.Is(in.ROOM_RIGHTS_2):
		mgr.HasRights = false
		mgr.rightsUpdated.Dispatch()
	case e.Is(in.ROOM_RIGHTS_3):
		mgr.IsOwner = true
	}
}

func (mgr *Manager) handleHeightmap(e *g.Intercept) {
	if !mgr.IsInRoom {
		return
	}

	mgr.Heightmap = strings.Split(e.Packet.ReadString(), "\r")

	if debug.Enabled {
		if len(mgr.Heightmap) > 0 {
			dbg.Printf("received heightmap (%dx%d)", len(mgr.Heightmap[0]), len(mgr.Heightmap))
		} else {
			dbg.Println("WARNING: empty heightmap")
		}
	}
}

func (mgr *Manager) handleActiveObjects(e *g.Intercept) {
	if !mgr.IsInRoom {
		return
	}

	var objects []Object
	e.Packet.Read(&objects)

	for _, object := range objects {
		mgr.Objects[object.Id] = object
	}

	mgr.objectsLoaded.Dispatch(ObjectsArgs{Objects: objects})

	dbg.Printf("loaded %d objects", len(objects))
}

func (mgr *Manager) handleActiveObjectAdd(e *g.Intercept) {
	if !mgr.IsInRoom {
		return
	}

	var object Object
	e.Packet.Read(&object)

	mgr.Objects[object.Id] = object

	mgr.objectAdded.Dispatch(ObjectArgs{Object: object})

	dbg.Printf("added object %s (ID: %d)", object.Class, object.Id)
}

func (mgr *Manager) handleActiveObjectUpdate(e *g.Intercept) {
	if !mgr.IsInRoom {
		return
	}

	var cur Object
	e.Packet.Read(&cur)

	if prev, ok := mgr.Objects[cur.Id]; ok {
		mgr.Objects[cur.Id] = cur
		mgr.objectUpdated.Dispatch(ObjectUpdateArgs{Prev: prev, Cur: cur})
		dbg.Printf("updated object %s (ID: %d)", cur.Class, cur.Id)
	} else {
		dbg.Printf("WARNING: failed to find object to update (ID: %d)", cur.Id)
	}
}

func (mgr *Manager) handleActiveObjectRemove(e *g.Intercept) {
	if !mgr.IsInRoom {
		return
	}

	var object Object
	e.Packet.Read(&object)

	if _, ok := mgr.Objects[object.Id]; ok {
		delete(mgr.Objects, object.Id)
		mgr.objectRemoved.Dispatch(ObjectArgs{Object: object})
		dbg.Printf("removed object (ID: %d)", object.Id)
	} else {
		dbg.Printf("WARNING: failed to remove object (ID: %d)", object.Id)
	}
}

func (mgr *Manager) handleSlideObjectBundle(e *g.Intercept) {
	if !mgr.IsInRoom {
		return
	}

	var bundle SlideObjectBundle
	e.Packet.Read(&bundle)

	var pSource *Object
	if bundle.RollerId != 0 {
		if source, ok := mgr.Objects[bundle.RollerId]; ok {
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
		obj, ok := mgr.Objects[bundleObj.Id]
		if ok {
			obj.X = args.To.X
			obj.Y = args.To.Y
			obj.Z = bundleObj.ToZ
			mgr.Objects[obj.Id] = obj
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
		if ent, ok := mgr.Entities[bundle.Entity.Id]; ok {
			ent.X = bundle.To.X
			ent.Y = bundle.To.Y
			ent.Z = bundle.Entity.ToZ
			mgr.Entities[ent.Index] = ent
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

	mgr.slide.Dispatch(args)

	dbg.Printf("processed slide object bundle (%d objects, with entity: %t)", len(args.ObjectSlides), args.EntitySlide != nil)
}

func (mgr *Manager) handleItems(e *g.Intercept) {
	if !mgr.IsInRoom {
		return
	}

	var items Items
	e.Packet.Read(&items)

	for _, item := range items {
		// TODO: check if this loop gets optimized away when !debug.Enabled
		if _, exists := mgr.Items[item.Id]; exists {
			dbg.Printf("WARNING: duplicate item (ID: %d)", item.Id)
		}
		mgr.Items[item.Id] = item
	}

	mgr.itemsLoaded.Dispatch(ItemsArgs{Items: items})

	dbg.Printf("loaded %d items", len(items))
}

func (mgr *Manager) handleItems2(e *g.Intercept) {
	if !mgr.IsInRoom {
		return
	}

	var item Item
	e.Packet.Read(&item)

	if _, exists := mgr.Items[item.Id]; exists {
		dbg.Printf("WARNING: duplicate item (ID: %d)", item.Id)
	}
	mgr.Items[item.Id] = item

	mgr.itemAdded.Dispatch(ItemArgs{Item: item})

	dbg.Printf("added item %s (ID: %d)", item.Class, item.Id)
}

func (mgr *Manager) handleRemoveItem(e *g.Intercept) {
	if !mgr.IsInRoom {
		return
	}

	strId := e.Packet.ReadString()
	id, err := strconv.Atoi(strId)
	if err != nil {
		dbg.Printf("WARNING: invalid item id: %s", strId)
		return
	}

	if item, ok := mgr.Items[id]; ok {
		delete(mgr.Items, item.Id)
		mgr.itemRemoved.Dispatch(ItemArgs{Item: item})
		dbg.Printf("removed item %s (ID: %d)", item.Class, item.Id)
	} else {
		dbg.Printf("WARNING: failed to remove item (ID: %d)", item.Id)
	}
}

func (mgr *Manager) handleUsers(e *g.Intercept) {
	if !mgr.IsInRoom {
		return
	}

	var ents []Entity
	e.Packet.Read(&ents)

	for _, entity := range ents {
		// TODO: check if this branch gets optimized away when !debug.Enabled
		if _, exists := mgr.Entities[entity.Index]; exists {
			dbg.Printf("WARNING: duplicate entity index: %d", entity.Index)
		}
		mgr.Entities[entity.Index] = entity
	}

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
	if !mgr.IsInRoom {
		return
	}

	n := e.Packet.ReadInt()
	for range n {
		var status EntityStatus
		e.Packet.Read(&status)
		entity, ok := mgr.Entities[status.Index]
		if !ok {
			dbg.Printf("WARNING: failed to find entity to update (index: %d)", status.Index)
			continue
		}

		cur := entity
		cur.Tile = status.Tile
		cur.Action = status.Action
		mgr.Entities[status.Index] = cur

		mgr.entityUpdated.Dispatch(EntityUpdateArgs{
			Prev: entity,
			Cur:  cur,
		})

		dbg.Printf("status update for %q (index: %d): %q",
			entity.Name, entity.Index, status.Action)
	}
}

func (mgr *Manager) handleChat(e *g.Intercept) {
	if !mgr.IsInRoom {
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

	if entity, ok := mgr.Entities[index]; ok {
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
	if !mgr.IsInRoom {
		return
	}

	s := e.Packet.ReadString()
	index, err := strconv.Atoi(s)
	if err != nil {
		dbg.Printf("WARNING: invalid index: %q", s)
		return
	}

	if entity, ok := mgr.Entities[index]; ok {
		delete(mgr.Entities, index)
		mgr.entityLeft.Dispatch(EntityArgs{Entity: entity})
		dbg.Printf("removed entity %q (index: %d)", entity.Name, entity.Index)
	} else {
		dbg.Printf("WARNING: failed to remove entity (index: %d)", index)
	}
}

func (mgr *Manager) handleClc(e *g.Intercept) {
	mgr.leaveRoom()
}
