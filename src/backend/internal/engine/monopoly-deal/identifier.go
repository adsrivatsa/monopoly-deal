package monopoly_deal

import (
	"fmt"

	"github.com/google/uuid"
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

type IdentifierTranslator struct {
	UUIDToIdentifier map[uuid.UUID]Identifier `json:"uuid_to_identifier" msgpack:"a"`
	IdentifierToUUID map[Identifier]uuid.UUID `json:"identifier_to_uuid" msgpack:"b"`
}

func NewIdentifierTranslator(gen *IdentifierGenerator, playerUUIDs []uuid.UUID) IdentifierTranslator {
	it := IdentifierTranslator{
		UUIDToIdentifier: make(map[uuid.UUID]Identifier),
		IdentifierToUUID: make(map[Identifier]uuid.UUID),
	}

	for _, playerUUID := range playerUUIDs {
		id := gen.New()
		it.UUIDToIdentifier[playerUUID] = id
		it.IdentifierToUUID[id] = playerUUID
	}

	return it
}

func (it *IdentifierTranslator) GetIdentifier(u uuid.UUID) (Identifier, bool) {
	id, ok := it.UUIDToIdentifier[u]
	return id, ok
}

func (it *IdentifierTranslator) GetUUID(id Identifier) (uuid.UUID, bool) {
	u, ok := it.IdentifierToUUID[id]
	return u, ok
}
