package service

import (
	"context"
	"time"

	deal_no_mercy "the-deal/internal/engine/deal-no-mercy"
	monopoly_deal "the-deal/internal/engine/monopoly-deal"
	"the-deal/internal/errors"
	"the-deal/internal/event"
	"the-deal/internal/schema"
	"the-deal/internal/schema/room_schema"
	"the-deal/internal/store"
	"the-deal/internal/token"

	"github.com/adsrivatsa/randomize"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

func (c *Controller) PublishEvent(ctx context.Context, roomID uuid.UUID, e *room_schema.ServerMessage) error {
	s := &schema.ServerMessage{
		Payload: &schema.ServerMessage_RoomMessage{
			RoomMessage: e,
		},
	}

	buf, err := proto.Marshal(s)
	if err != nil {
		return err
	}

	return c.bus.Publish(ctx, event.RoomChannelPre+roomID.String(), event.NewServerMessageEvent(buf))
}

func (c *Controller) ListenRoomEvents(ctx context.Context, tp token.Payload, callback func(message *schema.ServerMessage)) error {
	rp, err := c.store.GetRoomPlayer(ctx, tp.PlayerID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return errors.EntityNotFound(errors.EntityRoom, err)
		}
		return err
	}

	ch, err := c.bus.Subscribe(ctx, event.RoomChannelPre+rp.RoomID.String())
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case e := <-ch:
			if e.Kind == event.KindServerMessage {
				var msg schema.ServerMessage
				err = proto.Unmarshal(e.Message, &msg)
				if err != nil {
					return err
				}
				callback(&msg)
			}
		}
	}
}

func (c *Controller) GetRoom(ctx context.Context, tp token.Payload) (LongRoom, error) {
	r, err := c.store.GetRoomByPlayer(ctx, tp.PlayerID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return LongRoom{}, errors.EntityNotFound(errors.EntityRoom, err)
		}
		return LongRoom{}, err
	}

	ps, err := c.store.GetPlayersByRoom(ctx, r.RoomID)
	if err != nil {
		return LongRoom{}, err
	}

	players := make([]ShortPlayer, len(ps))
	for i, p := range ps {
		players[i] = ShortPlayer{
			PlayerID:    p.PlayerID,
			DisplayName: p.DisplayName,
			ImageUrl:    p.ImageUrl,
			IsHost:      p.IsHost,
			IsReady:     p.IsReady,
		}
	}

	return LongRoom{
		RoomID:      r.RoomID,
		DisplayName: r.DisplayName,
		Capacity:    r.Capacity,
		Occupied:    r.Occupied,
		Players:     players,
		Game:        r.Game,
		Settings:    r.Settings,
		IsPrivate:   r.IsPrivate,
	}, nil
}

func (c *Controller) GetRoomJoinedMessage(ctx context.Context, tp token.Payload) (*schema.ServerMessage, error) {
	r, err := c.store.GetRoomByPlayer(ctx, tp.PlayerID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return nil, errors.EntityNotFound(errors.EntityRoom, err)
		}
		return nil, err
	}

	ps, err := c.store.GetPlayersByRoom(ctx, r.RoomID)
	if err != nil {
		return nil, err
	}

	players := make([]*room_schema.Player, len(ps))
	for i, p := range ps {
		players[i] = &room_schema.Player{
			PlayerId:    p.PlayerID.String(),
			DisplayName: p.DisplayName,
			AvatarUrl:   p.ImageUrl,
			IsReady:     p.IsReady,
			IsHost:      p.IsHost,
		}
	}

	return &schema.ServerMessage{
		Payload: &schema.ServerMessage_RoomMessage{
			RoomMessage: &room_schema.ServerMessage{
				Payload: &room_schema.ServerMessage_RoomJoined{
					RoomJoined: &room_schema.RoomJoined{
						Room: &room_schema.Room{
							RoomId:      r.RoomID.String(),
							DisplayName: r.DisplayName,
							Players:     players,
							Capacity:    r.Capacity,
							Occupied:    r.Occupied,
							Game:        r.Game.Proto(),
							Settings:    r.Settings,
							IsPrivate:   r.IsPrivate,
						},
					},
				},
			},
		},
	}, nil
}

func (c *Controller) ListRooms(ctx context.Context, args ListRoomsParams) (ListRoomsRes, error) {
	var game store.NullGameType
	if args.Game != nil {
		game.GameType = *args.Game
		game.Valid = true
	}

	rps, err := c.store.ListRooms(ctx, store.ListRoomsParams{
		Limit:  args.Limit,
		Offset: args.Offset,
		Search: args.Search,
		Game:   game,
	})
	if err != nil {
		return ListRoomsRes{}, err
	}

	res := ListRoomsRes{}

	rooms := make([]LongRoom, len(rps))
	for i, rp := range rps {
		res.TotalCount = rp.TotalCount
		rooms[i] = LongRoom{
			RoomID:      rp.RoomID,
			DisplayName: rp.RoomDisplayName,
			Capacity:    rp.RoomCapacity,
			Occupied:    rp.RoomOccupied,
			Players: []ShortPlayer{
				{
					PlayerID:    rp.PlayerID,
					DisplayName: rp.HostDisplayName,
					ImageUrl:    rp.HostImageUrl,
				},
			},
			Game:     rp.RoomGame,
			Settings: rp.RoomSettings,
		}
	}

	res.Rooms = rooms
	return res, nil
}

func (c *Controller) CreateRoom(ctx context.Context, tp token.Payload, args CreateRoomParams) (Room, error) {
	_, err := c.store.GetRoomPlayer(ctx, tp.PlayerID)
	if err == nil { // room exists
		return Room{}, errors.EntityAlreadyExists(errors.EntityRoom)
	}
	if errors.DBErrorCode(err) != errors.NoDataFound {
		return Room{}, err
	}

	_, err = c.store.GetGameByPlayer(ctx, tp.PlayerID)
	if err == nil {
		return Room{}, errors.EntityAlreadyExists(errors.EntityGame)
	}
	if errors.DBErrorCode(err) != errors.NoDataFound {
		return Room{}, err
	}

	var r store.Room
	err = c.store.ExecTx(ctx, func(q *store.Queries) error {
		var err error
		r, _, err = c.createRoomTx(ctx, q, createRoomTxParams{
			DisplayName: args.DisplayName,
			Capacity:    args.Capacity,
			Game:        args.Game,
			Settings:    args.Settings,
			IsPrivate:   args.IsPrivate,
			IsQuickPlay: false,
			PlayerID:    tp.PlayerID,
			IsHost:      true,
			IsReady:     false,
		})
		return err
	})
	if err != nil {
		return Room{}, err
	}

	return Room{
		RoomID:      r.RoomID,
		DisplayName: r.DisplayName,
		Capacity:    r.Capacity,
		Occupied:    r.Occupied,
		Game:        r.Game,
		Settings:    r.Settings,
		IsPrivate:   r.IsPrivate,
	}, nil
}

func (c *Controller) JoinRoom(ctx context.Context, tp token.Payload, roomID uuid.UUID) error {
	_, err := c.store.GetRoomPlayer(ctx, tp.PlayerID)
	if err == nil { // room exists
		return errors.EntityAlreadyExists(errors.EntityRoomPlayer)
	}
	if errors.DBErrorCode(err) != errors.NoDataFound {
		return err
	}

	var rp store.RoomPlayer
	err = c.store.ExecTx(ctx, func(q *store.Queries) error {
		var err error
		_, rp, err = c.joinRoomTx(ctx, q, joinRoomTxParams{
			RoomID:   roomID,
			PlayerID: tp.PlayerID,
			IsHost:   false,
			IsReady:  false,
		})
		return err
	})
	if err != nil {
		return err
	}

	p, err := c.store.GetPlayer(ctx, store.GetPlayerParams{
		PlayerID: &tp.PlayerID,
	})
	if err != nil {
		return err
	}

	player := &room_schema.Player{
		PlayerId:    p.PlayerID.String(),
		DisplayName: p.DisplayName,
		AvatarUrl:   p.ImageUrl,
		IsReady:     rp.IsReady,
		IsHost:      rp.IsHost,
		JoinedAt:    rp.JoinedAt.UnixMilli(),
	}

	e := &room_schema.ServerMessage{
		Payload: &room_schema.ServerMessage_PlayerJoinedRoom{
			PlayerJoinedRoom: &room_schema.PlayerJoinedRoom{
				RoomId: roomID.String(),
				Player: player,
			},
		},
	}

	return c.PublishEvent(ctx, rp.RoomID, e)
}

func (c *Controller) LeaveRoom(ctx context.Context, tp token.Payload) error {
	var newHostPlayerID *string
	deleteRoom := false
	var roomID uuid.UUID
	err := c.store.ExecTx(ctx, func(q *store.Queries) error {
		rp, err := q.GetRoomPlayerForUpdate(ctx, tp.PlayerID)
		if err != nil {
			if errors.DBErrorCode(err) == errors.NoDataFound {
				return errors.EntityNotFound(errors.EntityRoom)
			}
			return err
		}

		roomID = rp.RoomID

		if rp.IsHost { // host is leaving, find new host
			newHost, err := q.GetOldestRoomPlayer(ctx, store.GetOldestRoomPlayerParams{
				RoomID:          rp.RoomID,
				LeavingPlayerID: tp.PlayerID,
			})
			if err != nil {
				if errors.DBErrorCode(err) == errors.NoDataFound {
					// no other player exists to become new host, delete the room
					deleteRoom = true
				} else {
					return err
				}
			}

			if !deleteRoom { // set the new host
				playerID := newHost.PlayerID.String()
				newHostPlayerID = &playerID

				_, err = q.UpdateRoomPlayerHost(ctx, store.UpdateRoomPlayerHostParams{
					IsHost:   true,
					RoomID:   rp.RoomID,
					PlayerID: newHost.PlayerID,
				})
				if err != nil {
					return err
				}
			}
		}

		err = q.DeleteRoomPlayer(ctx, store.DeleteRoomPlayerParams{
			RoomID:   rp.RoomID,
			PlayerID: tp.PlayerID,
		})
		if err != nil {
			return err
		}

		if deleteRoom {
			err = q.DeleteRoom(ctx, rp.RoomID)
		} else {
			_, err = q.DecrementRoomOccupied(ctx, rp.RoomID)
		}
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	if !deleteRoom {
		e := &room_schema.ServerMessage{
			Payload: &room_schema.ServerMessage_PlayerLeftRoom{
				PlayerLeftRoom: &room_schema.PlayerLeftRoom{
					RoomId:          roomID.String(),
					PlayedId:        tp.PlayerID.String(),
					NewHostPlayerId: newHostPlayerID,
				},
			},
		}

		err = c.PublishEvent(ctx, roomID, e)
	}

	return err
}

func (c *Controller) ToggleIsReady(ctx context.Context, tp token.Payload) error {
	rp, err := c.store.GetRoomPlayer(ctx, tp.PlayerID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return errors.EntityNotFound(errors.EntityRoom)
		}
		return err
	}

	rp, err = c.store.ToggleRoomPlayerIsReady(ctx, store.ToggleRoomPlayerIsReadyParams{
		RoomID:   rp.RoomID,
		PlayerID: tp.PlayerID,
	})
	if err != nil {
		return err
	}

	e := &room_schema.ServerMessage{
		Payload: &room_schema.ServerMessage_PlayerToggledReady{
			PlayerToggledReady: &room_schema.PlayerToggledReady{
				PlayerId: tp.PlayerID.String(),
				IsReady:  rp.IsReady,
			},
		},
	}

	return c.PublishEvent(ctx, rp.RoomID, e)
}

func (c *Controller) UpdateRoomSettings(ctx context.Context, tp token.Payload, args UpdateRoomSettingsParams) error {
	rp, err := c.store.GetRoomPlayer(ctx, tp.PlayerID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return errors.EntityNotFound(errors.EntityRoom)
		}
		return err
	}

	if !rp.IsHost {
		return errors.PlayerIsNotHost
	}

	r, err := c.store.UpdateRoomSettings(ctx, store.UpdateRoomSettingsParams{
		Capacity:  args.Capacity,
		Game:      args.Game,
		Settings:  args.Settings,
		IsPrivate: args.IsPrivate,
		RoomID:    rp.RoomID,
	})
	if err != nil {
		return err
	}

	e := &room_schema.ServerMessage{
		Payload: &room_schema.ServerMessage_SettingsUpdated{
			SettingsUpdated: &room_schema.SettingsUpdated{
				Capacity:  r.Capacity,
				Game:      r.Game.Proto(),
				Settings:  r.Settings,
				IsPrivate: r.IsPrivate,
			},
		},
	}

	return c.PublishEvent(ctx, rp.RoomID, e)
}

func (c *Controller) HandleRoomEvent(ctx context.Context, tp token.Payload, msg *schema.ClientMessage_RoomMessage) error {
	switch p := msg.RoomMessage.GetPayload().(type) {
	case *room_schema.ClientMessage_Chat:
		return c.handleRoomChat(ctx, tp, p)
	default:
		return nil
	}
}

func (c *Controller) handleRoomChat(ctx context.Context, tp token.Payload, msg *room_schema.ClientMessage_Chat) error {
	rp, err := c.store.GetRoomPlayer(ctx, tp.PlayerID)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return errors.EntityNotFound(errors.EntityRoom)
		}
		return err
	}

	e := &room_schema.ServerMessage{
		Payload: &room_schema.ServerMessage_ChatReceived{
			ChatReceived: &room_schema.ChatReceived{
				PlayerId: tp.PlayerID.String(),
				Payload:  msg.Chat.Payload,
			},
		},
	}

	return c.PublishEvent(ctx, rp.RoomID, e)
}

func (c *Controller) LeaveQuickPlayWaitingIfPending(ctx context.Context, tp token.Payload) error {
	_, err := c.store.GetGameByPlayer(ctx, tp.PlayerID)
	if err == nil {
		return nil
	}
	if errors.DBErrorCode(err) != errors.NoDataFound {
		return err
	}

	err = c.LeaveRoom(ctx, tp)
	if err == nil {
		return nil
	}

	if errors.IsEntityNotFound(err) {
		return nil
	}
	return err
}

func (c *Controller) JoinQuickPlayRoom(ctx context.Context, tp token.Payload, game store.GameType) (*uuid.UUID, error) {
	lockGameKey, err := quickPlayGameLockKey(game)
	if err != nil {
		return nil, err
	}

	_, err = c.store.GetRoomPlayer(ctx, tp.PlayerID)
	if err == nil {
		return nil, errors.EntityAlreadyExists(errors.EntityRoom)
	}
	if errors.DBErrorCode(err) != errors.NoDataFound {
		return nil, err
	}

	_, err = c.store.GetGameByPlayer(ctx, tp.PlayerID)
	if err == nil {
		return nil, errors.EntityAlreadyExists(errors.EntityGame)
	}
	if errors.DBErrorCode(err) != errors.NoDataFound {
		return nil, err
	}

	displayName, err := randomize.Do[string]()
	if err != nil {
		return nil, err
	}

	var capacity int32
	var settings []byte
	switch game {
	case store.GameTypeMonopolyDeal:
		capacity = 2
		set := monopoly_deal.DefaultSettings()
		settings, err = set.Encode()
	case store.GameTypeDealNoMercy:
		capacity = 2
		set := deal_no_mercy.DefaultSettings()
		settings, err = set.Encode()
	default:
		return nil, errors.GameNotSupported
	}
	if err != nil {
		return nil, err
	}

	var r store.Room
	var rp store.RoomPlayer
	var gameID uuid.UUID
	var gameStarted bool
	var mdAction monopoly_deal.Action
	var dnmAction deal_no_mercy.Action
	var deadline time.Time
	err = c.store.ExecTx(ctx, func(q *store.Queries) error {
		err = q.QuickPlayBucketXactLock(ctx, store.QuickPlayBucketXactLockParams{
			LockClass: quickPlayAdvisoryLockClass,
			GameKey:   lockGameKey,
		})
		if err != nil {
			return err
		}

		r, err = q.GetQuickPlayRoomForUpdate(ctx, game)
		if err == nil {
			r, rp, err = c.joinRoomTx(ctx, q, joinRoomTxParams{
				RoomID:   r.RoomID,
				PlayerID: tp.PlayerID,
				IsHost:   false,
				IsReady:  true,
			})
			if err != nil {
				return err
			}

			if r.Occupied >= r.Capacity {
				ps, err := q.GetPlayersByRoom(ctx, r.RoomID)
				if err != nil {
					if errors.DBErrorCode(err) == errors.NoDataFound {
						return errors.EntityNotFound(errors.EntityRoom)
					}
					return err
				}

				playerIDs := make([]uuid.UUID, 0, len(ps))
				for _, p := range ps {
					playerIDs = append(playerIDs, p.PlayerID)
				}

				switch r.Game {
				case store.GameTypeMonopolyDeal:
					gameID, mdAction, deadline, err = c.MonopolyDealController.CreateGameFromRoomTx(ctx, q, r.RoomID, playerIDs)
				case store.GameTypeDealNoMercy:
					gameID, dnmAction, deadline, err = c.DealNoMercyController.CreateGameFromRoomTx(ctx, q, r.RoomID, playerIDs)
				default:
					return errors.GameNotSupported
				}
				if err != nil {
					return err
				}
				gameStarted = true
			}

			return nil
		}
		if errors.DBErrorCode(err) != errors.NoDataFound {
			return err
		}

		r, rp, err = c.createRoomTx(ctx, q, createRoomTxParams{
			DisplayName: displayName,
			Capacity:    capacity,
			Game:        game,
			Settings:    settings,
			IsPrivate:   true,
			IsQuickPlay: true,
			PlayerID:    tp.PlayerID,
			IsHost:      true,
			IsReady:     true,
		})

		return err
	})
	if err != nil {
		return nil, err
	}

	p, err := c.store.GetPlayer(ctx, store.GetPlayerParams{
		PlayerID: &tp.PlayerID,
	})
	if err != nil {
		return nil, err
	}

	if gameStarted {
		switch game {
		case store.GameTypeMonopolyDeal:
			err = c.MonopolyDealController.PublishAction(ctx, gameID, mdAction, deadline)
		case store.GameTypeDealNoMercy:
			err = c.DealNoMercyController.PublishAction(ctx, gameID, dnmAction, deadline)
		default:
			return nil, errors.GameNotSupported
		}
		if err != nil {
			return nil, err
		}

		e := &room_schema.ServerMessage{
			Payload: &room_schema.ServerMessage_GameStarted{
				GameStarted: &room_schema.GameStarted{
					GameId: gameID.String(),
				},
			},
		}

		err = c.PublishEvent(ctx, r.RoomID, e)
		if err != nil {
			return nil, err
		}
		gid := gameID
		return &gid, nil
	}

	player := &room_schema.Player{
		PlayerId:    p.PlayerID.String(),
		DisplayName: p.DisplayName,
		AvatarUrl:   p.ImageUrl,
		IsReady:     rp.IsReady,
		IsHost:      rp.IsHost,
		JoinedAt:    rp.JoinedAt.UnixMilli(),
	}

	e := &room_schema.ServerMessage{
		Payload: &room_schema.ServerMessage_PlayerJoinedRoom{
			PlayerJoinedRoom: &room_schema.PlayerJoinedRoom{
				RoomId: r.RoomID.String(),
				Player: player,
			},
		},
	}

	err = c.PublishEvent(ctx, r.RoomID, e)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
