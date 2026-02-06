package nav

import (
	"fmt"
	"strconv"
	"strings"

	g "xabbo.b7c.io/goearth"
	"xabbo.b7c.io/goearth/internal/debug"
)

type NodeInfo struct {
	NodeMask int
	Root     Node
}

func (info *NodeInfo) Parse(r g.PacketReader) {
	*info = NodeInfo{}
	info.NodeMask = r.ReadInt()
	r.Read(&info.Root)

	nodeMap := map[int]*Node{info.Root.Id: &info.Root}

	var skips map[int]int
	if debug.Enabled {
		skips = map[int]int{}
	}

	for r.Available() > 0 {
		var node Node
		r.Read(&node)
		// if node.Id <= 0 {
		// 	break
		// }
		if _, exists := nodeMap[node.Id]; !exists {
			nodeMap[node.Id] = &node
			if parent, ok := nodeMap[node.ParentId]; ok {
				parent.Children = append(parent.Children, node)
			} else {
				dbg.Printf("WARNING: orphaned nav node: %d %q", node.Id, getNodeName(&node))
			}
		} else if debug.Enabled {
			skips[node.Id]++
		}
	}

	if debug.Enabled && len(skips) > 0 {
		dbg.Println("skipped duplicate nodes:")
		for id := range skips {
			dbg.Printf("  %d %q (%dx)", id, getNodeName(nodeMap[id]), skips[id])
		}
	}
}

func getNodeName(node *Node) string {
	switch data := node.Data.(type) {
	case *Category:
		return data.Name
	case *Room:
		return data.Name
	default:
		return ""
	}
}

// Node represents a hierarchical tree structure of categories and rooms.
type Node struct {
	Id       int
	Type     NodeType
	ParentId int
	Parent   *Node
	Children []Node
	Data     NavNodeData
}

// Category represents a room category in the navigator.
type Category struct {
	Name      string
	MaxUsers  int
	UserCount int
}

func (*Category) isNavNodeData() {}

// Room represents a room in the navigator.
type Room struct {
	Id           int
	Name         string
	MaxUsers     int
	UserCount    int
	Port         string
	Door         string
	UnitId       string
	Casts        string
	Visible      bool
	UsersInQueue int
	Owner        string
	Description  string
	Filter       string
}

func (*Room) isNavNodeData() {}

// NavNodeData represents a category or room.
type NavNodeData interface {
	isNavNodeData()
}

type NodeType int

const (
	NodeCategory NodeType = iota
	NodePublicRoom
	NodeUserRoom
)

// Traverse performs a breadth-first search,
// calling the provided yield func for each node, including the root node.
// Returning false from the yield func breaks the iteration loop.
func (root *Node) Traverse(yield func(node *Node) bool) {
	nodes := []*Node{root}
	for len(nodes) > 0 {
		node := nodes[0]
		if !yield(node) {
			return
		}
		nodes = nodes[1:]
		for i := range node.Children {
			nodes = append(nodes, &node.Children[i])
		}
	}
}

// Find performs a breadth-first search and returns
// the first node in the hierarchy that matches the condition.
// If a node is not found, nil is returned.
func (root *Node) Find(cond func(node *Node) bool) (result *Node) {
	root.Traverse(func(node *Node) bool {
		if cond(node) {
			result = node
			return false
		}
		return true
	})
	return
}

// Rooms returns all rooms in the node hierarchy.
func (root *Node) Rooms() (rooms []Room) {
	root.Traverse(func(node *Node) bool {
		if room, ok := node.Data.(*Room); ok {
			rooms = append(rooms, *room)
		}
		return true
	})
	return
}

func (node *Node) Parse(r g.PacketReader) {
	*node = Node{}
	node.Id = r.ReadInt()
	// Have not yet encountered this case in testing,
	// and removing it makes this compatible with
	// favourite room results (its node ID is 0).
	// if node.Id <= 0 {
	//   return
	// }
	node.Type = NodeType(r.ReadInt())
	name := r.ReadString()
	userCount := r.ReadInt()
	maxUsers := r.ReadInt()
	node.ParentId = r.ReadInt()
	switch node.Type {
	case NodeCategory:
		node.Data = &Category{
			Name:      name,
			UserCount: userCount,
			MaxUsers:  maxUsers,
		}
		node.Children = []Node{}
	case NodePublicRoom:
		node.Data = &Room{
			Id:           1000 + node.Id,
			Name:         name,
			UserCount:    userCount,
			MaxUsers:     maxUsers,
			UnitId:       r.ReadString(),
			Port:         strconv.Itoa(r.ReadInt()),
			Door:         strconv.Itoa(r.ReadInt()),
			Casts:        r.ReadString(),
			UsersInQueue: r.ReadInt(),
			Visible:      r.ReadBool(),
		}
	case NodeUserRoom:
		node.Type = NodeCategory
		node.Children = parseCategoryRoomNodes(node, r)
	default:
		panic(fmt.Errorf("unknown node type: %d", node.Type))
	}
}

func parseCategoryRoomNodes(parent *Node, r g.PacketReader) []Node {
	n := r.ReadInt()
	nodes := make([]Node, 0, n)
	for range n {
		id := r.ReadInt()
		node := Node{
			Id:       id,
			Type:     NodeUserRoom,
			ParentId: parent.Id,
			Parent:   parent,
			Data: &Room{
				Id:          id,
				Name:        r.ReadString(),
				Owner:       r.ReadString(),
				Door:        r.ReadString(),
				UserCount:   r.ReadInt(),
				MaxUsers:    r.ReadInt(),
				Description: r.ReadString(),
				Visible:     true,
			},
		}
		nodes = append(nodes, node)
	}
	return nodes
}

type Rooms []Room

func (rooms *Rooms) Parse(r g.PacketReader) {
	*rooms = Rooms{}
	s := r.ReadString()
	lines := strings.Split(s, "\r")
	for _, line := range lines {
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) != 9 {
			panic(fmt.Errorf("parse RoomResults: invalid field count (%d): %v", len(fields), fields))
		}
		id, err := strconv.Atoi(fields[0])
		if err != nil {
			panic(fmt.Errorf("parse RoomResults: invalid ID: %s", fields[0]))
		}
		userCount, err := strconv.Atoi(fields[5])
		if err != nil {
			panic(fmt.Errorf("parse RoomResults: invalid UserCount: %s", fields[5]))
		}
		maxUsers, err := strconv.Atoi(fields[6])
		if err != nil {
			panic(fmt.Errorf("parse RoomResults: invalid MaxUsers: %s", fields[6]))
		}
		node := Room{
			Id:          id,
			Name:        fields[1],
			Owner:       fields[2],
			Door:        fields[3],
			Port:        fields[4],
			UserCount:   userCount,
			MaxUsers:    maxUsers,
			Filter:      fields[7],
			Description: fields[8],
			Visible:     true,
		}
		*rooms = append(*rooms, node)
	}
}
