package catalog

import (
	"strings"

	g "xabbo.b7c.io/goearth"
)

// Index is a map of catalog page ID -> page name.
type Index map[string]string

func (index *Index) Parse(r g.PacketReader) {
	*index = map[string]string{}

	lines := strings.Split(r.ReadString(), "\r")
	for _, line := range lines {
		fields := strings.SplitN(line, "\t", 2)
		if len(fields) != 2 {
			continue
		}
		(*index)[fields[0]] = fields[1]
	}
}

type Page struct {
	Id          string
	Name        string
	Layout      string
	HeaderText  string
	HeaderImage string
}

type Product struct {
	Type       string
	ClassId    int
	Params     string
	Count      int
	Expiration int
}

type Offer struct {
	Code           int
	Name           string
	PriceInCredits int
	PriceInPixels  int
	Products       []Product
}
