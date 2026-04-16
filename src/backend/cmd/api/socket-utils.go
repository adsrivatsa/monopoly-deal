package main

import (
	"context"
	"fun-kames/internal/schema"
	"time"
)

func (s *Server) ping(ctx context.Context, sock *socket) {
	ticker := time.NewTicker(s.cfg.WebsocketPingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			msg := &schema.ServerMessage{
				Payload: &schema.ServerMessage_Ping{
					Ping: &schema.Ping{
						TimeUnixMs: time.Now().UnixMilli(),
					},
				},
			}
			sock.send(msg)
		}
	}
}
