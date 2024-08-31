package nav

import (
	g "xabbo.b7c.io/goearth"
	"xabbo.b7c.io/goearth/internal/debug"
	"xabbo.b7c.io/goearth/shockwave/in"
	"xabbo.b7c.io/goearth/shockwave/out"
)

var dbg = debug.NewLogger("[nav]")

type Manager struct {
	ix g.Interceptor
}

func NewManager(ix g.Interceptor) *Manager {
	mgr := &Manager{ix: ix}
	return mgr
}

func (mgr *Manager) Navigate(nodeId int) *Node {
	mgr.ix.Send(out.NAVIGATE, false /* hide full */, nodeId, 1 /* depth */)
	if pkt := mgr.ix.Recv(in.NAVNODEINFO).If(nodeIdEq(nodeId)).TimeoutSec(10).Block().Wait(); pkt != nil {
		var navNodeInfo NodeInfo
		navNodeInfo.Parse(pkt, &pkt.Pos)
		return &navNodeInfo.Root
	} else {
		return nil
	}
}

func (mgr *Manager) Search(query string) (rooms Rooms, ok bool) {
	mgr.ix.Send(out.SRCHF, query)
	if pkt := mgr.ix.Recv(in.FLAT_RESULTS_2).TimeoutSec(10).Block().Wait(); pkt != nil {
		rooms.Parse(pkt, &pkt.Pos)
		ok = true
	}
	return
}

func (mgr *Manager) GetOwnRooms() (rooms Rooms, ok bool) {
	mgr.ix.Send(out.SUSERF)
	if pkt := mgr.ix.Recv(in.FLAT_RESULTS).TimeoutSec(10).Block().Wait(); pkt != nil {
		rooms.Parse(pkt, &pkt.Pos)
		ok = true
	}
	return
}

func (mgr *Manager) GetFavouriteRooms() (rooms Rooms, ok bool) {
	mgr.ix.Send(out.GETFVRF, false)
	if pkt := mgr.ix.Recv(in.FAVOURITEROOMRESULTS).TimeoutSec(10).Block().Wait(); pkt != nil {
		var nodeInfo NodeInfo
		nodeInfo.Parse(pkt, &pkt.Pos)
		nodeInfo.Root.Traverse(func(node *Node) bool {
			if room, ok := node.Data.(*Room); ok {
				rooms = append(rooms, *room)
			}
			return true
		})
		ok = true
	}
	return
}

func nodeIdEq(nodeId int) func(*g.Packet) bool {
	return func(p *g.Packet) bool {
		p.ReadInt()
		return p.ReadInt() == nodeId
	}
}
