package event

import (
	"context"
	"time"

	"google.golang.org/protobuf/proto"
)

func (b *Bus) set(ctx context.Context, key string, value proto.Message, expiration time.Duration) error {
	payload, err := proto.Marshal(value)
	if err != nil {
		return err
	}

	return b.client.Set(ctx, key, payload, expiration).Err()
}

func (b *Bus) get(ctx context.Context, key string, out proto.Message) error {
	data, err := b.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}

	err = proto.Unmarshal(data, out)
	return err
}

func (b *Bus) list(ctx context.Context, prefix string, callback func([]byte)) error {
	iter := b.client.Scan(ctx, 0, prefix+"*", 0).Iterator()

	for iter.Next(ctx) {
		key := iter.Val()

		data, err := b.client.Get(ctx, key).Bytes()
		if err != nil {
			return err
		}

		callback(data)
	}

	return iter.Err()
}

func (b *Bus) delete(ctx context.Context, key string) error {
	return b.client.Del(ctx, key).Err()
}
