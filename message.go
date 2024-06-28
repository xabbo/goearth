package goearth

import "strings"

// Generate message identifiers.
//go:generate go run .generate/messages/main.go --dir . --variant flash-windows
//go:generate go run .generate/messages/main.go --dir shockwave --variant shockwave-windows

// Defines a message direction.
type Direction int

const (
	// Represents the incoming (to client) direction.
	In Direction = 1 << iota
	// Represents the outgoing (to server) direction.
	Out
)

// Id creates an identifier using the provided direction and name.
func (d Direction) Id(name string) Identifier {
	return Identifier{d, name}
}

func (d Direction) String() string {
	switch d {
	case In:
		return "incoming"
	case Out:
		return "outgoing"
	default:
		return "unknown"
	}
}

// Defines a message direction and name.
type Identifier struct {
	Dir  Direction
	Name string
}

// Defines a message direction and value.
type Header struct {
	Dir   Direction
	Value uint16
}

// Defines a message direction, value, and name.
type NamedHeader struct {
	Header
	Name string
}

// Returns true if the header has the same direction and name as the specified identifier.
func (header *NamedHeader) Is(identifier Identifier) bool {
	return header.Dir == identifier.Dir && header.Name == identifier.Name
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

func outHeader(value uint16) *NamedHeader {
	return &NamedHeader{Header{Out, value}, ""}
}

// Creates a new header with the specified direction, value and name.
func NewHeader(dir Direction, value uint16, name string) *NamedHeader {
	return &NamedHeader{Header{dir, value}, name}
}

// Defines a map of incoming and outgoing headers.
type Headers struct {
	valueMap map[Direction]map[uint16]*NamedHeader
	nameMap  map[Direction]map[string]*NamedHeader
}

// Creates a new header map.
func NewHeaders() *Headers {
	return &Headers{
		valueMap: map[Direction]map[uint16]*NamedHeader{
			In: {}, Out: {},
		},
		nameMap: map[Direction]map[string]*NamedHeader{
			In: {}, Out: {},
		},
	}
}

// Resets the header map.
func (h *Headers) Reset() {
	for _, dir := range []Direction{In, Out} {
		for k := range h.valueMap[dir] {
			delete(h.valueMap[dir], k)
		}
		for k := range h.nameMap[dir] {
			delete(h.nameMap[dir], k)
		}
	}
}

// Adds a header to the map.
func (h *Headers) Add(header *NamedHeader) {
	h.valueMap[header.Dir][header.Value] = header
	h.nameMap[header.Dir][strings.ToLower(header.Name)] = header
}

// Gets the header with the specified identifier. Returns nil if it does not exist.
func (h *Headers) Get(identifier Identifier) *NamedHeader {
	return h.ByName(identifier.Dir, identifier.Name)
}

// Gets the header with the specified direction and value. Returns nil if it does not exist.
func (h *Headers) ByValue(dir Direction, value uint16) *NamedHeader {
	return h.valueMap[dir][value]
}

// Gets the header with the specified direction and name. Returns nil if it does not exist.
func (h *Headers) ByName(dir Direction, name string) *NamedHeader {
	return h.nameMap[dir][strings.ToLower(name)]
}
