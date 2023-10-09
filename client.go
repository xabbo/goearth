package goearth

import "strings"

// Defines a type of game client.
type ClientType string

const (
	// Represents the Flash client.
	FLASH ClientType = "FLASH"
	// Represents the Unity client.
	UNITY = "UNITY"
)

// Defines information about a game client.
type Client struct {
	Version    string
	Identifier string
	Type       ClientType
}

func (t *ClientType) Parse(p *Packet, pos *int) {
	*t = ClientType(p.ReadStringPtr(pos))
}

func (t ClientType) String() string {
	return strings.ToTitle(string(t))
}
