package main

import (
	"log/slog"
	"monopoly-deal/internal/config"
	"monopoly-deal/internal/store"
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

	handler := slog.NewTextHandler(os.Stdout, nil)
	logger := slog.New(handler)

	srv := NewServer(cfg, logger, pool)
	err = srv.Start()
	if err != nil {
		panic(err)
	}
}
