package event

import (
	"context"
	"fmt"
	"fun-kames/internal/schema"

	"github.com/go-redis/redis/v8"
	"google.golang.org/protobuf/proto"
)

const (
	RoomChannelPre = "room-channel:"
)

type Bus struct {
	client *redis.Client
}

func NewBus(client *redis.Client) *Bus {
	return &Bus{client: client}
}

func (b *Bus) Publish(ctx context.Context, channel string, event *schema.ServerMessage) error {
	payload, err := proto.Marshal(event)
	if err != nil {
		return err
	}

	return b.client.Publish(ctx, channel, payload).Err()
}

func (b *Bus) Subscribe(ctx context.Context, channel string) (chan *schema.ServerMessage, error) {
	sub := b.client.Subscribe(ctx, channel)

	msgCh := make(chan *schema.ServerMessage, 32)
	go func() {
		defer sub.Close()
		defer close(msgCh)

		for {
			msg, err := sub.ReceiveMessage(ctx)
			if err != nil {
				fmt.Println(err)
				return
			}

			var out schema.ServerMessage
			err = proto.Unmarshal([]byte(msg.Payload), &out)
			if err != nil {
				fmt.Println(err)
				return
			}

			select {
			case msgCh <- &out:
			case <-ctx.Done():
				return
			}
		}
	}()

	return msgCh, nil
}
