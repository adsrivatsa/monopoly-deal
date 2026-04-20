package main

import (
	"context"
	stderrors "errors"
	"fun-kames/internal/errors"
	"fun-kames/internal/schema"
	"fun-kames/internal/schema/monopoly_deal_schema"
	"fun-kames/internal/token"
	"net/http"
)

func (s *Server) MonopolyDealSocket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tp, err := tokenFromRequest(r, token.AccessToken)
	if err != nil {
		ErrorHTTP(w, err)
		return
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		ErrorHTTP(w, err)
		return
	}

	sock, ctx := newSocket(conn, ctx)

	s.gameSocketsMu.Lock()
	oldSock, ok := s.gameSockets[tp.PlayerID]
	if ok {
		oldSock.close(errors.DuplicateSocket)
	}
	s.gameSockets[tp.PlayerID] = sock
	s.gameSocketsMu.Unlock()
	defer func() {
		s.gameSocketsMu.Lock()
		if s2, ok := s.gameSockets[tp.PlayerID]; ok && s2 == oldSock {
			delete(s.gameSockets, tp.PlayerID)
		}
		s.gameSocketsMu.Unlock()
	}()

	msg := s.services.GetMonopolyDealGame(ctx, tp)
	sock.send(msg)

	go s.ping(ctx, sock)
	go s.handleClientMonopolyDealMessages(ctx, sock, tp)

	callback := func(message *schema.ServerMessage) {
		sock.send(message)
	}

	err = s.services.ListenGameEvents(ctx, tp, callback)
	if err != nil {
		sock.error(err)
		return
	}
}

func (s *Server) handleClientMonopolyDealMessages(ctx context.Context, sock *socket, tp token.Payload) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		msg := sock.read()
		if msg == nil {
			return
		}

		switch p := msg.GetPayload().(type) {
		case *schema.ClientMessage_MonopolyDealMessage:
			err := s.services.HandleMonopolyDealEvent(ctx, tp, p)
			if err != nil {
				var intErr errors.Error
				if !stderrors.As(err, &intErr) {
					intErr = errors.Internal(err)
				}

				sock.send(&schema.ServerMessage{
					Payload: &schema.ServerMessage_MonopolyDealMessage{
						MonopolyDealMessage: &monopoly_deal_schema.ServerMessage{
							Payload: &monopoly_deal_schema.ServerMessage_Error{
								Error: &monopoly_deal_schema.Error{
									Message: intErr.Message,
									Status:  int32(intErr.Status),
									Code:    intErr.Code,
								},
							},
						},
					},
				})
			}
		default:
			sock.error(errors.InvalidMessageType[schema.ClientMessage_MonopolyDealMessage]())
			return
		}
	}
}
