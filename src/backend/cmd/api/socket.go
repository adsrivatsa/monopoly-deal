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

func (s *Server) socketRoutes() *chi.Mux {
	router := chi.NewRouter()

	router.Get("/lobby", s.LobbySocket)

	return router
}

func (s *Server) foreverPing(ctx context.Context, conn *websocket.Conn) {
	ticker := time.NewTicker(s.cfg.WebsocketPingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			message := &schema.LobbyMessage{
				Payload: &schema.LobbyMessage_Ping{
					Ping: &schema.Ping{
						TimeUnixMs: time.Now().UnixMilli(),
					},
				},
			}
			data, _ := proto.Marshal(message)
			err := conn.WriteMessage(websocket.BinaryMessage, data)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func (s *Server) foreverListRooms(ctx context.Context, conn *websocket.Conn) {

}

var upgrader = websocket.Upgrader{}

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

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go s.foreverPing(ctx, conn)

	go s.foreverListRooms(ctx, conn)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			break
		}

		var message schema.LobbyMessage
		err = proto.Unmarshal(msg, &message)
		if err != nil {
			LobbyError(conn, err)
			continue
		}

		switch p := message.Payload.(type) {
		case *schema.LobbyMessage_CreateRoom:
			err = s.controller.CreateRoom(ctx, tp, p.CreateRoom)

		case *schema.LobbyMessage_JoinRoom:
			err = s.controller.JoinRoom(ctx, tp, p.JoinRoom)

		case *schema.LobbyMessage_LeaveRoom:

		}
		if err != nil {
			LobbyError(conn, err)
		}
	}

	fmt.Println("disconnected")
}
