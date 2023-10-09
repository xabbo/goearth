package goearth

import "strings"

// Defines a message direction.
type Direction int

const (
	// Represents the incoming (to client) direction.
	INCOMING Direction = 1 << iota
	// Represents the outgoing (to server) direction.
	OUTGOING
)

func (d Direction) String() string {
	switch d {
	case INCOMING:
		return "incoming"
	case OUTGOING:
		return "outgoing"
	default:
		return "unknown"
	}
}

// Represents a message name and direction.
type Identifier struct {
	Dir  Direction
	Name string
}

// Defines information about a message.
type MsgInfo struct {
	Id        int
	Hash      string
	Name      string
	Structure string
	Outgoing  bool
	Source    string
}

// Defines a message name, direction and value.
type Header struct {
	dir   Direction
	value uint16
	name  string
}

func inHeader(value uint16) *Header {
	return &Header{INCOMING, value, ""}
}

func outHeader(value uint16) *Header {
	return &Header{OUTGOING, value, ""}
}

// Creates a new header with the specified direction, value and name.
func NewHeader(dir Direction, value uint16, name string) *Header {
	return &Header{dir, value, name}
}

// Gets the direction of the header.
func (h *Header) Dir() Direction {
	return h.dir
}

// Gets the value of the header.
func (h *Header) Value() uint16 {
	return h.value
}

// Gets the name of the header.
func (h *Header) Name() string {
	return h.name
}

// Defines a map of incoming and outgoing headers.
type Headers struct {
	valueMap map[Direction]map[uint16]*Header
	nameMap  map[Direction]map[string]*Header
}

// Creates a new header map.
func NewHeaders() *Headers {
	return &Headers{
		valueMap: map[Direction]map[uint16]*Header{
			INCOMING: {}, OUTGOING: {},
		},
		nameMap: map[Direction]map[string]*Header{
			INCOMING: {}, OUTGOING: {},
		},
	}
}

// Resets the header map.
func (h *Headers) Reset() {
	for _, dir := range []Direction{INCOMING, OUTGOING} {
		for k := range h.valueMap[dir] {
			delete(h.valueMap[dir], k)
		}
		for k := range h.nameMap[dir] {
			delete(h.nameMap[dir], k)
		}
	}
}

// Adds a header to the map.
func (h *Headers) Add(header *Header) {
	h.valueMap[header.dir][header.value] = header
	h.nameMap[header.dir][strings.ToLower(header.name)] = header
}

// Gets the header with the specified identifier. Returns nil if it does not exist.
func (h *Headers) Get(identifier Identifier) *Header {
	return h.ByName(identifier.Dir, identifier.Name)
}

// Gets the header with the specified direction and value. Returns nil if it does not exist.
func (h *Headers) ByValue(dir Direction, value uint16) *Header {
	return h.valueMap[dir][value]
}

// Gets the header with the specified direction and name. Returns nil if it does not exist.
func (h *Headers) ByName(dir Direction, name string) *Header {
	return h.nameMap[dir][strings.ToLower(name)]
}
