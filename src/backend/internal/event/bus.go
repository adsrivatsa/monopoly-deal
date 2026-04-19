package event

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis/v8"
)

const (
	RoomChannelPre = "room-channel:"
	GameChannelPre = "game-channel:"
)

type Bus struct {
	client *redis.Client
}

func NewBus(client *redis.Client) *Bus {
	return &Bus{client: client}
}

type Kind int

const (
	KindUnknown Kind = iota
	KindServerMessage
	KindMonopolyDealGameState
)

type Event struct {
	Kind    Kind   `json:"kind"`
	Message []byte `json:"message"`
}

func NewServerMessageEvent(message []byte) Event {
	return Event{KindServerMessage, message}
}

func NewMonopolyDealGameStateEvent(message []byte) Event {
	return Event{KindMonopolyDealGameState, message}
}

func (b *Bus) Publish(ctx context.Context, channel string, event Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return b.client.Publish(ctx, channel, payload).Err()
}

func (b *Bus) Subscribe(ctx context.Context, channel string) (chan Event, error) {
	sub := b.client.Subscribe(ctx, channel)

	msgCh := make(chan Event, 32)
	go func() {
		defer sub.Close()
		defer close(msgCh)

		for {
			msg, err := sub.ReceiveMessage(ctx)
			if err != nil {
				fmt.Println(err)
				return
			}

			var out Event
			err = json.Unmarshal([]byte(msg.Payload), &out)
			if err != nil {
				fmt.Println(err)
				return
			}

			select {
			case msgCh <- out:
			case <-ctx.Done():
				return
			}
		}
	}()

	return msgCh, nil
}
