package main

import (
	"context"
	"fmt"
	monopoly_deal "fun-kames/internal/engine/monopoly-deal"
	"fun-kames/internal/errors"
	"fun-kames/internal/schema"
	"fun-kames/internal/service"
	"fun-kames/internal/store"
	"fun-kames/internal/token"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (s *Server) roomRoutes() *chi.Mux {
	router := chi.NewRouter()

	router.Get("/", s.GetRoom)
	router.Post("/list", s.ListRooms)
	router.Post("/", s.CreateRoom)
	router.Patch("/join/{"+ROOM_ID+"}", s.JoinRoom)
	router.Patch("/leave", s.LeaveRoom)
	router.Patch("/ready", s.ToggleIsReady)
	router.Put("/settings", s.UpdateRoomSettings)
	router.Get("/socket", s.RoomSocket)

	return router
}

func (s *Server) ListRooms(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	_, err := tokenFromRequest(r, token.AccessToken)
	if err != nil {
		ErrorHTTP(w, err)
		return
	}

	args, err := ReadAndValidate[ListRoomsParams](w, r)
	if err != nil {
		ErrorHTTP(w, err)
		return
	}

	rooms, err := s.services.ListRooms(ctx, service.ListRoomsParams(args))
	if err != nil {
		ErrorHTTP(w, err)
		return
	}

	WriteHTTP(w, http.StatusOK, rooms)
}

func (s *Server) GetRoom(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tp, err := tokenFromRequest(r, token.AccessToken)
	if err != nil {
		ErrorHTTP(w, err)
		return
	}

	room, err := s.services.GetRoom(ctx, tp)
	if err != nil {
		ErrorHTTP(w, err)
		return
	}

	WriteHTTP(w, http.StatusOK, room)
}

func (s *Server) CreateRoom(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tp, err := tokenFromRequest(r, token.AccessToken)
	if err != nil {
		ErrorHTTP(w, err)
		return
	}

	args, err := ReadAndValidate[CreateRoomParams](w, r)
	if err != nil {
		ErrorHTTP(w, err)
		return
	}

	switch args.Game {
	case store.GameTypeMonopolyDeal:
		var settings monopoly_deal.Settings
		err = settings.Decode(args.Settings)
		if err != nil {
			ErrorHTTP(w, err)
			return
		}

		err = Validate(settings)
		if err != nil {
			ErrorHTTP(w, err)
			return
		}
	default:
		ErrorHTTP(w, errors.GameNotSupported)
		return
	}

	room, err := s.services.CreateRoom(ctx, tp, service.CreateRoomParams{
		DisplayName: args.DisplayName,
		Capacity:    int32(args.Capacity),
		Game:        args.Game,
		Settings:    args.Settings,
	})
	if err != nil {
		ErrorHTTP(w, err)
		return
	}

	WriteHTTP(w, http.StatusOK, room)
}

func (s *Server) JoinRoom(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	roomIDStr := chi.URLParam(r, ROOM_ID)
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		ErrorHTTP(w, errors.InvalidUUID(err))
		return
	}

	tp, err := tokenFromRequest(r, token.AccessToken)
	if err != nil {
		ErrorHTTP(w, err)
		return
	}

	err = s.services.JoinRoom(ctx, tp, roomID)
	if err != nil {
		ErrorHTTP(w, err)
		return
	}

	WriteHTTP(w, http.StatusOK, nil)
}

func (s *Server) LeaveRoom(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tp, err := tokenFromRequest(r, token.AccessToken)
	if err != nil {
		ErrorHTTP(w, err)
		return
	}

	err = s.services.LeaveRoom(ctx, tp)
	if err != nil {
		ErrorHTTP(w, err)
		return
	}

	WriteHTTP(w, http.StatusOK, nil)
}

func (s *Server) ToggleIsReady(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tp, err := tokenFromRequest(r, token.AccessToken)
	if err != nil {
		ErrorHTTP(w, err)
		return
	}

	err = s.services.ToggleIsReady(ctx, tp)
	if err != nil {
		ErrorHTTP(w, err)
		return
	}

	WriteHTTP(w, http.StatusOK, nil)
}

func (s *Server) UpdateRoomSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tp, err := tokenFromRequest(r, token.AccessToken)
	if err != nil {
		ErrorHTTP(w, err)
		return
	}

	args, err := ReadAndValidate[UpdateRoomSettingsParams](w, r)
	if err != nil {
		ErrorHTTP(w, err)
		return
	}

	switch args.Game {
	case store.GameTypeMonopolyDeal:
		var settings monopoly_deal.Settings
		err = settings.Decode(args.Settings)
		if err != nil {
			ErrorHTTP(w, err)
			return
		}

		err = Validate(settings)
		if err != nil {
			ErrorHTTP(w, err)
			return
		}
	default:
		ErrorHTTP(w, errors.GameNotSupported)
		return
	}

	err = s.services.UpdateRoomSettings(ctx, tp, service.UpdateRoomSettingsParams{
		Capacity: int32(args.Capacity),
		Game:     args.Game,
		Settings: args.Settings,
	})
	if err != nil {
		ErrorHTTP(w, err)
		return
	}

	WriteHTTP(w, http.StatusOK, nil)
}

func (s *Server) RoomSocket(w http.ResponseWriter, r *http.Request) {
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

	s.roomSocketsMu.Lock()
	oldSock, ok := s.roomSockets[tp.PlayerID]
	if ok {
		oldSock.close(errors.DuplicateSocket)
	}
	s.roomSockets[tp.PlayerID] = sock
	s.roomSocketsMu.Unlock()
	defer func() {
		s.roomSocketsMu.Lock()
		if s2, ok := s.roomSockets[tp.PlayerID]; ok && s2 == oldSock {
			delete(s.roomSockets, tp.PlayerID)
		}
		s.roomSocketsMu.Unlock()
	}()

	go s.ping(ctx, sock)
	go s.handleClientRoomMessages(ctx, sock, tp)

	callback := func(message *schema.ServerMessage) {
		sock.send(message)
	}

	err = s.services.ListenRoomEvents(ctx, tp, callback)
	if err != nil {
		sock.error(err)
		return
	}
}

func (s *Server) handleClientRoomMessages(ctx context.Context, sock *socket, tp token.Payload) {
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
		case *schema.ClientMessage_RoomMessage:
			err := s.services.HandleRoomEvent(ctx, tp, p)
			if err != nil {
				fmt.Println(err)
			}
		default:
			sock.error(errors.InvalidMessageType[schema.ClientMessage]())
			return
		}
	}
}
