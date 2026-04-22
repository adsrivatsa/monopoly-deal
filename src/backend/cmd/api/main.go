package main

import (
	"log/slog"
	"os"
	"the-deal/internal/config"
	"the-deal/internal/event"
	"the-deal/internal/store"
	"time"
)

func main() {
	envPath := "../dev.env"
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
