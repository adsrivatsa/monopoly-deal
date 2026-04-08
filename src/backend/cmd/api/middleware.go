package main

import (
	"context"
	"monopoly-deal/internal/errors"
	"monopoly-deal/internal/token"
	"net/http"

	"github.com/gorilla/sessions"
)

func tokenFromRequest(r *http.Request, tokenType token.TokenType) (token.Payload, error) {
	ctx := r.Context()
	var payload token.Payload

	payload, ok := ctx.Value(tokenType).(token.Payload)
	if !ok {
		return token.Payload{}, errors.InvalidToken
	}

	return payload, nil
}

func tokenMiddleware(tokenMaker token.Maker, cookieStore *sessions.CookieStore, sessionName string, tokenType token.TokenType) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sess, _ := cookieStore.Get(r, sessionName)
			t, ok := sess.Values[tokenType].(string)
			if !ok || t == "" {
				ErrorHTTP(w, errors.InvalidToken)
				return
			}

			payload, err := tokenMaker.VerifyToken(t, tokenType)
			if err != nil {
				ErrorHTTP(w, err)
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, tokenType, payload)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}
