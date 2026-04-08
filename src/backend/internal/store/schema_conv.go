package store

import (
	"fmt"
	"monopoly-deal/internal/schema"
)

func (r RoomStatus) SchemaRoomStatus() schema.RoomStatus {
	switch r {
	case RoomStatusCompleted:
		return schema.RoomStatus_COMPLETED
	case RoomStatusGame:
		return schema.RoomStatus_GAME
	case RoomStatusLobby:
		return schema.RoomStatus_LOBBY
	default:
		panic(fmt.Sprintf("unexpected store.RoomStatus: %#v", r))
	}
}
