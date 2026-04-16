package main

import (
	"fun-kames/internal/config"
	"fun-kames/internal/event"
	"fun-kames/internal/store"
	"log/slog"
	"os"
	"time"
)

func main() {
	envPath := "../.env"
	cfg, err := config.Load(envPath)
	if err != nil {
		panic(err)
	}

	pool := store.NewPostgresPool(cfg, time.Second*5)
	defer pool.Close()

	client := event.NewRedisClient(cfg, time.Second*5)
	defer client.Close()

	handler := slog.NewTextHandler(os.Stdout, nil)
	logger := slog.New(handler)

	srv := NewServer(cfg, logger, pool, client)
	err = srv.Start()
	if err != nil {
		panic(err)
	}
}
