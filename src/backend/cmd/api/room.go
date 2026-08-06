package main

import (
	"context"
	"fmt"
	"net/http"
	monopoly_deal "the-deal/internal/engine/monopoly-deal"
	"the-deal/internal/errors"
	"the-deal/internal/schema"
	"the-deal/internal/schema/room_schema"
	"the-deal/internal/service"
	"the-deal/internal/store"
	"the-deal/internal/token"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (s *Server) roomRoutes() *chi.Mux {
	router := chi.NewRouter()

	router.Get("/", s.GetRoom)
	router.Post("/list", s.ListRooms)
	router.Post("/", s.CreateRoom)
	router.Get("/quick-play/{"+GAME_TYPE+"}/socket", s.QuickPlaySocket)
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
		IsPrivate:   args.IsPrivate,
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

func (s *Server) QuickPlaySocket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tp, err := tokenFromRequest(r, token.AccessToken)
	if err != nil {
		ErrorHTTP(w, err)
		return
	}

	gameTypeStr := chi.URLParam(r, GAME_TYPE)
	game := store.GameType(gameTypeStr)
	if !game.Valid() {
		ErrorHTTP(w, errors.GameNotSupported)
		return
	}
	switch game {
	case store.GameTypeMonopolyDeal:
	default:
		ErrorHTTP(w, errors.GameNotSupported)
		return
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		ErrorHTTP(w, err)
		return
	}

	sock, sockCtx := newSocket(conn, ctx)

	s.roomSocketsMu.Lock()
	oldSock, ok := s.roomSockets[tp.PlayerID]
	if ok {
		oldSock.close(errors.DuplicateSocket)
	}
	s.roomSockets[tp.PlayerID] = sock
	s.roomSocketsMu.Unlock()
	defer func() {
		s.roomSocketsMu.Lock()
		if s2, ok := s.roomSockets[tp.PlayerID]; ok && s2 == sock {
			delete(s.roomSockets, tp.PlayerID)
		}
		s.roomSocketsMu.Unlock()

		leaveCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = s.services.LeaveQuickPlayWaitingIfPending(leaveCtx, tp)
	}()

	go s.ping(sockCtx, sock)

	startedGameID, err := s.services.JoinQuickPlayRoom(ctx, tp, game)
	if err != nil {
		sock.error(err)
		return
	}

	if startedGameID != nil {
		s.sendGameStartedAndWait(sock, *startedGameID)
		return
	}

	roomJoined, err := s.services.GetRoomJoinedMessage(ctx, tp)
	if err != nil {
		s.quickPlayRoomGone(ctx, sock, tp, err)
		return
	}
	sock.send(roomJoined)

	go s.handleClientRoomMessages(sockCtx, sock, tp)

	callback := func(message *schema.ServerMessage) {
		sock.send(message)
	}

	err = s.services.ListenRoomEvents(sockCtx, tp, callback)
	if err != nil {
		s.quickPlayRoomGone(ctx, sock, tp, err)
		return
	}
}

// quickPlayRoomGone runs when the player's quick-play room disappears between
// joining and subscribing: the room row is deleted when the game starts, so a
// lookup failure here usually means another player filled the room and the
// game began before this connection could subscribe.
func (s *Server) quickPlayRoomGone(ctx context.Context, sock *socket, tp token.Payload, cause error) {
	g, err := s.services.GetGame(ctx, tp)
	if err != nil {
		sock.error(cause)
		return
	}

	s.sendGameStartedAndWait(sock, g.GameID)
}

func (s *Server) sendGameStartedAndWait(sock *socket, gameID uuid.UUID) {
	sock.send(&schema.ServerMessage{
		Payload: &schema.ServerMessage_RoomMessage{
			RoomMessage: &room_schema.ServerMessage{
				Payload: &room_schema.ServerMessage_GameStarted{
					GameStarted: &room_schema.GameStarted{
						GameId: gameID.String(),
					},
				},
			},
		},
	})

	// Client navigates to /game/socket; keep connection open until they disconnect.
	for sock.read() != nil {
	}
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
		Capacity:  int32(args.Capacity),
		Game:      args.Game,
		Settings:  args.Settings,
		IsPrivate: args.IsPrivate,
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
		if s2, ok := s.roomSockets[tp.PlayerID]; ok && s2 == sock {
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
			sock.error(errors.InvalidMessageType[schema.ClientMessage_RoomMessage]())
			return
		}
	}
}
