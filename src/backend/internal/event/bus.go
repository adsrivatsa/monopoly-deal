package event

import (
	"context"
	"encoding/json"

	"github.com/go-redis/redis/v8"
)

type Bus struct {
	client *redis.Client
}

func NewBus(client *redis.Client) *Bus {
	return &Bus{client: client}
}

func (p *Bus) Publish(ctx context.Context, channel string, event any) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.client.Publish(ctx, channel, payload).Err()
}

func (p *Bus) Set(ctx context.Context, key string, value any) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return p.client.Set(ctx, key, payload, 0).Err()
}

func (p *Bus) Get(ctx context.Context, key string, out any) error {
	data, err := p.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}

	return json.Unmarshal(data, out)
}
