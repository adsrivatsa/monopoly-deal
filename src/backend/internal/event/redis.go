package event

import (
	"context"
	"fmt"
	"fun-kames/internal/config"
	"time"

	"github.com/go-redis/redis/v8"
)

func NewRedisClient(cfg config.Config, waitTime time.Duration) *redis.Client {
	count := 0

	for {
		client, err := newRedisClient(cfg)
		if err == nil {
			return client
		}

		count++
		if count >= 5 {
			fmt.Println("unable to connect: ", err)
			fmt.Printf("retrying in %d ms...", waitTime.Milliseconds())
			time.Sleep(waitTime)
			count = 0
		}
	}
}

func newRedisClient(cfg config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisHost,
		Password: cfg.RedisPassword,
		DB:       0,
	})

	err := client.Ping(context.Background()).Err()
	return client, err
}
