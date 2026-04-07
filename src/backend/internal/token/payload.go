package token

import (
	"encoding/gob"
	"time"

	"github.com/google/uuid"
)

func init() {
	gob.Register(TokenType(0))
}

type TokenType int

const (
	AccessToken TokenType = iota
	RefreshToken
)

type Payload struct {
	TokenID   uuid.UUID `json:"token_id"`
	PlayerID  uuid.UUID `json:"player_id"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
	TokenType TokenType `json:"token_type"`
}

func (t *Payload) GetIssuedAt() time.Time {
	return t.IssuedAt
}

func (t *Payload) GetExpiresAt() time.Time {
	return t.ExpiresAt
}

func (t *Payload) GetType() TokenType {
	return t.TokenType
}

func (t *Payload) SetIssuedAt(t2 time.Time) {
	t.IssuedAt = t2
}

func (t *Payload) SetExpiresAt(t2 time.Time) {
	t.ExpiresAt = t2
}

func (t *Payload) SetType(tokenType TokenType) {
	t.TokenType = tokenType
}
