package token

import (
	"encoding/json"
	stderrors "errors"
	"fun-kames/internal/errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTPayload struct {
	jwt.Claims
	MarshalledPayload string `json:"payload"`
	IssuedAt          int64  `json:"iat"`
	ExpiresAt         int64  `json:"exp"`
}

func (j JWTPayload) GetExpirationTime() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Unix(j.ExpiresAt, 0)), nil
}

func (j JWTPayload) GetIssuedAt() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Unix(j.IssuedAt, 0)), nil
}

func (j JWTPayload) GetNotBefore() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Unix(j.IssuedAt, 0)), nil
}

func (j JWTPayload) GetIssuer() (string, error) {
	return "", nil
}

func (j JWTPayload) GetSubject() (string, error) {
	return "", nil
}

func (j JWTPayload) GetAudience() (jwt.ClaimStrings, error) {
	return nil, nil
}

type Maker struct {
	tokenDuration map[TokenType]time.Duration
	secretKey     []byte
	signingMethod jwt.SigningMethod
	parser        *jwt.Parser
}

func (j Maker) CreateToken(payload Payload, tokenType TokenType) (string, Payload, error) {
	now := time.Now()

	payload.SetIssuedAt(now)
	payload.SetExpiresAt(now.Add(j.tokenDuration[tokenType]))
	payload.SetType(tokenType)

	js, err := json.Marshal(payload)
	if err != nil {
		return "", payload, errors.Internal(err)
	}

	tokenPayload := JWTPayload{
		MarshalledPayload: string(js),
		IssuedAt:          payload.GetIssuedAt().Unix(),
		ExpiresAt:         payload.GetExpiresAt().Unix(),
	}
	token := jwt.NewWithClaims(j.signingMethod, tokenPayload)
	signed, err := token.SignedString(j.secretKey)
	if err != nil {
		return "", payload, errors.Internal(err)
	}

	return signed, payload, nil
}

func (j Maker) VerifyToken(s string, tokenType TokenType) (Payload, error) {
	var payload Payload
	token, err := j.parser.Parse(s, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.InvalidToken
		}
		return j.secretKey, nil
	})

	if err != nil || !token.Valid {
		if stderrors.Is(err, jwt.ErrTokenInvalidClaims) {
			return payload, errors.InvalidToken
		}
		return payload, err
	}

	tokenPayload, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return payload, errors.InvalidTokenContent
	}

	if err := json.Unmarshal([]byte(tokenPayload["payload"].(string)), &payload); err != nil {
		return payload, errors.Internal(err)
	}

	if payload.GetType() != tokenType {
		return payload, errors.InvalidTokenType
	}

	return payload, nil
}

func NewMaker(durations map[TokenType]time.Duration, secret string) Maker {
	signingMethod := jwt.SigningMethodHS256

	return Maker{
		tokenDuration: durations,
		secretKey:     []byte(secret),
		signingMethod: signingMethod,
		parser: jwt.NewParser(
			jwt.WithValidMethods([]string{signingMethod.Name}),
			jwt.WithExpirationRequired(),
		),
	}
}
