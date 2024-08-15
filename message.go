package goearth

import (
	"fmt"
	"strings"
)

// Generate message identifiers.
//go:generate go run .generate/messages/main.go --dir . --variant flash-windows
//go:generate go run .generate/messages/main.go --dir shockwave --variant shockwave-windows

// Defines a message direction.
type Direction int

const (
	Unknown Direction = 0
	// Represents the incoming (to client) direction.
	In Direction = 1 << (iota - 1)
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

func (d Direction) ShortString() string {
	switch d {
	case In:
		return "in"
	case Out:
		return "out"
	default:
		return "?"
	}
}

// Identifier defines a message direction and name.
type Identifier struct {
	Dir  Direction
	Name string
}

func (id Identifier) String() string {
	return id.Dir.ShortString() + ":" + id.Name
}

// Header defines a message direction and value.
type Header struct {
	Dir   Direction
	Value uint16
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

// Defines a map of incoming and outgoing headers.
type Headers struct {
	identifiers map[Direction]map[string]Header
	names       map[Header]string
}

// Creates a new header map.
func NewHeaders() *Headers {
	return &Headers{
		identifiers: map[Direction]map[string]Header{In: {}, Out: {}},
		names:       map[Header]string{},
	}
}

// Resets the header map.
func (h *Headers) Reset() {
	for k := range h.identifiers {
		delete(h.identifiers, k)
	}
	for k := range h.names {
		delete(h.names, k)
	}
}

// Adds a header to the map.
func (h *Headers) Add(name string, header Header) {
	h.identifiers[header.Dir][strings.ToLower(name)] = header
	h.names[header] = name
}

// Gets the header with the specified identifier. Panics if it does not exist.
func (h *Headers) Get(id Identifier) (header Header) {
	header, ok := h.TryGet(id)
	if !ok {
		panic(fmt.Errorf("header does not exist for identifier %s", id))
	}
	return
}

// Gets the header with the specified identifier. Panics if it does not exist.
func (h *Headers) TryGet(id Identifier) (header Header, ok bool) {
	header, ok = h.identifiers[id.Dir][strings.ToLower(id.Name)]
	return
}

// Name gets the name of the header. Returns an empty string if it does not exist.
func (h *Headers) Name(header Header) string {
	return h.names[header]
}

// Is returns whether the specified header matches the identifier.
func (h *Headers) Is(header Header, id Identifier) bool {
	return header == h.identifiers[id.Dir][strings.ToLower(id.Name)]
}
