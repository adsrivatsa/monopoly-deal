package event

import (
	"context"
	"fmt"
	"monopoly-deal/internal/schema"
	"time"

	"github.com/go-redis/redis/v8"
	"google.golang.org/protobuf/proto"
)

const (
	LobbyChannel   = "lobby-chan"
	RoomStatePre   = "room-state:"
	RoomChannelPre = "room-channel:"
)

type Bus struct {
	client *redis.Client
}

func NewBus(client *redis.Client) *Bus {
	return &Bus{client: client}
}

func (p *Bus) Publish(ctx context.Context, channel string, event *schema.ServerMessage) error {
	payload, err := proto.Marshal(event)
	if err != nil {
		return err
	}

	return p.client.Publish(ctx, channel, payload).Err()
}

func (p *Bus) Subscribe(ctx context.Context, channel string) (chan *schema.ServerMessage, error) {
	sub := p.client.Subscribe(ctx, channel)

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

func (p *Bus) Set(ctx context.Context, key string, value *schema.ServerMessage, expiration time.Duration) error {
	payload, err := proto.Marshal(value)
	if err != nil {
		return err
	}

	return p.client.Set(ctx, key, payload, expiration).Err()
}

func (p *Bus) Get(ctx context.Context, key string) (*schema.ServerMessage, error) {
	data, err := p.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var out schema.ServerMessage
	err = proto.Unmarshal(data, &out)
	return &out, err
}

func (p *Bus) List(ctx context.Context, prefix string, callback func(key string, state *schema.ServerMessage)) error {
	iter := p.client.Scan(ctx, 0, prefix+"*", 0).Iterator()

	for iter.Next(ctx) {
		key := iter.Val()

		data, err := p.client.Get(ctx, key).Bytes()
		if err != nil {
			return err
		}

		var state schema.ServerMessage
		err = proto.Unmarshal(data, &state)
		if err != nil {
			return err
		}

		callback(key, &state)
	}

	return iter.Err()
}
