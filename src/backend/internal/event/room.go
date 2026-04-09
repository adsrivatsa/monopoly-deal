package event

import (
	"context"
	"fmt"
	"monopoly-deal/internal/schema"
	"time"

	"google.golang.org/protobuf/proto"
)

func (b *Bus) SetRoom(ctx context.Context, room *schema.Room) error {
	key := RoomStatePre + room.RoomId
	return b.set(ctx, key, room, time.Hour*24)
}

func (b *Bus) GetRoom(ctx context.Context, roomId string) (*schema.Room, error) {
	key := RoomStatePre + roomId

	var room schema.Room
	err := b.get(ctx, key, &room)
	return &room, err
}

func (b *Bus) ListRooms(ctx context.Context, callback func(msg *schema.Room)) error {
	return b.list(ctx, RoomStatePre, func(data []byte) {
		var room schema.Room
		err := proto.Unmarshal(data, &room)
		if err != nil {
			fmt.Println(err)
			return
		}
		callback(&room)
	})
}
func (b *Bus) DeleteRoom(ctx context.Context, roomId string) error {
	key := RoomStatePre + roomId
	return b.delete(ctx, key)
}
