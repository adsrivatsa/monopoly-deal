package main

import (
	"context"
	"fmt"
	"monopoly-deal/internal/errors"
	"monopoly-deal/internal/schema"
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

func foreverPing(ctx context.Context, conn *websocket.Conn, interval time.Duration) {
	ticker := time.NewTicker(interval)
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
	// TODO
}

var upgrader = websocket.Upgrader{}

func (s *Server) LobbySocket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		ErrorHTTP(w, errors.Internal(err))
		return
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go foreverPing(ctx, conn, s.cfg.WebsocketPingInterval)

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
			s.CreateRoom(ctx, conn, p.CreateRoom)

		case *schema.LobbyMessage_CreateRoomRes:

		case *schema.LobbyMessage_DeleteRoom:

		case *schema.LobbyMessage_DeleteRoomRes:

		case *schema.LobbyMessage_Error:

		case *schema.LobbyMessage_JoinRoom:

		case *schema.LobbyMessage_JoinRoomRes:

		case *schema.LobbyMessage_LeaveRoom:

		case *schema.LobbyMessage_LeaveRoomRes:

		case *schema.LobbyMessage_Ping:

		case *schema.LobbyMessage_RoomList:

		}
	}

	fmt.Println("disconnected")
}
