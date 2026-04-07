package main

import (
	"monopoly-deal/internal/config"
	"monopoly-deal/internal/store"
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

	srv := NewServer(cfg, pool)
	err = srv.Start()
	if err != nil {
		panic(err)
	}
}
