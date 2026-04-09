package main

import (
	"context"
	"fmt"
	"monopoly-deal/internal/errors"
	"monopoly-deal/internal/schema"
	"monopoly-deal/internal/token"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

func (s *Server) lobbyRoutes() *chi.Mux {
	router := chi.NewRouter()

	router.Get("/socket", s.LobbySocket)

	return router
}

func (s *Server) foreverPing(ctx context.Context, sock *socket) {
	ticker := time.NewTicker(s.cfg.WebsocketPingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			msg := &schema.ServerMessage{
				Payload: &schema.ServerMessage_Ping{
					Ping: &schema.Ping{
						TimeUnixMs: time.Now().UnixMilli(),
					},
				},
			}
			sock.send(msg)
		}
	}
}

func (s *Server) foreverListRooms(ctx context.Context, sock *socket) {
	serverMsgCh, err := s.controller.SubscribeLobbyEvents(ctx)
	if err != nil {
		sock.error(err)
		return
	}

	go func() {
		err = s.controller.ListRooms(ctx, func(msg *schema.ServerMessage) {
			sock.send(msg)
		})
		if err != nil {
			sock.error(err)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-serverMsgCh:
			if !ok {
				return
			}
			sock.send(msg)
		}
	}
}

func (s *Server) LobbySocket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tp, err := tokenFromRequest(r, token.AccessToken)
	if err != nil {
		ErrorHTTP(w, err)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		ErrorHTTP(w, errors.Internal(err))
		return
	}

	sock, ctx := newSocket(conn, ctx)

	s.lobbyMu.Lock()
	oldSock, ok := s.lobbySockets[tp.PlayerID]
	if ok {
		oldSock.close(errors.DuplicateSocket)
	}
	s.lobbySockets[tp.PlayerID] = sock
	s.lobbyMu.Unlock()

	go s.foreverPing(ctx, sock)

	go s.foreverListRooms(ctx, sock)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				fmt.Println(err)
			}
			break
		}

		var message schema.ClientMessage
		err = proto.Unmarshal(msg, &message)
		if err != nil {
			sock.error(err)
			continue
		}

		switch p := message.Payload.(type) {
		case *schema.ClientMessage_LobbyMessage:
			s.foreverHandleLobby(ctx, sock, tp, p.LobbyMessage)
		default:
			sock.error(errors.InvalidMessageType[*schema.ClientMessage_LobbyMessage]())
			continue
		}
	}

	s.lobbyMu.Lock()
	if current, exists := s.lobbySockets[tp.PlayerID]; exists && current == sock {
		delete(s.lobbySockets, tp.PlayerID)
	}
	s.lobbyMu.Unlock()

	sock.close(nil)
	fmt.Println("disconnected")
}

func (s *Server) foreverHandleLobby(ctx context.Context, sock *socket, tp token.Payload, p *schema.ClientLobbyMessage) {
	var err error
	switch act := p.Payload.(type) {
	case *schema.ClientLobbyMessage_CreateRoom:
		err = s.controller.CreateRoom(ctx, tp, act.CreateRoom)

	}
	if err != nil {
		sock.error(err)
	}
}
