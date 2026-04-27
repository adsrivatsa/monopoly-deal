package monopoly_deal

import (
	"fmt"
)

type Identifier string

const (
	NullIdentifier Identifier = "null"
)

type IdentifierGenerator struct {
	Next uint16 `json:"next" msgpack:"a"`
}

func NewIdentifierGenerator() *IdentifierGenerator {
	return &IdentifierGenerator{
		Next: 1,
	}
}

func (g *IdentifierGenerator) New() Identifier {
	id := Identifier(fmt.Sprintf("%03x", g.Next))
	g.Next++
	return id
}
