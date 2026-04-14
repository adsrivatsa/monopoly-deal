package main

import (
	"monopoly-deal/internal/errors"
	"monopoly-deal/internal/schema"
	"monopoly-deal/internal/service"
	"monopoly-deal/internal/token"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (s *Server) roomRoutes() *chi.Mux {
	router := chi.NewRouter()

	router.Post("/list", s.ListRooms)
	router.Post("/", s.CreateRoom)
	router.Get("/join/{"+ROOM_ID+"}", s.JoinRoom)
	router.Get("/leave", s.LeaveRoom)
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

	WriteHTTP(w, http.StatusOK, ListRoomRes(rooms))
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

	room, err := s.services.CreateRoom(ctx, tp, service.CreateRoomParams{
		DisplayName: args.DisplayName,
		Capacity:    int32(args.Capacity),
	})
	if err != nil {
		ErrorHTTP(w, err)
		return
	}

	WriteHTTP(w, http.StatusOK, Room(room))
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

func (s *Server) RoomSocket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tp, err := tokenFromRequest(r, token.AccessToken)
	if err != nil {
		ErrorHTTP(w, err)
		return
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		ErrorHTTP(w, errors.Internal(err))
		return
	}

	sock := newSocket(conn, ctx)

	s.roomSocketsMu.Lock()
	oldSock, ok := s.roomSockets[tp.PlayerID]
	if ok {
		oldSock.close(errors.DuplicateSocket)
	}
	s.roomSockets[tp.PlayerID] = sock
	s.roomSocketsMu.Unlock()

	go s.foreverPing(ctx, sock)

	callback := func(message *schema.ServerMessage) {
		sock.send(message)
	}

	err = s.services.ListenRoomEvents(ctx, tp, callback)
	if err != nil {
		sock.error(err)
		return
	}
}
