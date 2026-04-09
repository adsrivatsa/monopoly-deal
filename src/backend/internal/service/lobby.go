package service

import (
	"context"
	"monopoly-deal/internal/event"
	"monopoly-deal/internal/schema"
)

func (c *Controller) SubscribeLobbyEvents(ctx context.Context) (chan *schema.ServerMessage, error) {
	return c.bus.Subscribe(ctx, event.LobbyChannel)
}
