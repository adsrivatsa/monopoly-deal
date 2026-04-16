package service

import (
	"context"
	"fun-kames/internal/errors"
	"fun-kames/internal/store"
	"fun-kames/internal/token"
)

func (c *Controller) CreatePlayer(ctx context.Context, args CreatePlayerParams) (Player, error) {
	p, err := c.store.CreatePlayer(ctx, args)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return p, errors.EntityNotFound(errors.EntityPlayer, err)
		}
		return p, errors.Internal(err)
	}

	return p, nil
}

func (c *Controller) GetPlayer(ctx context.Context, tp token.Payload, args GetPlayerParams) (Player, error) {
	if args.PlayerID == nil && args.Email == nil {
		args.PlayerID = &tp.PlayerID
	}

	p, err := c.store.GetPlayer(ctx, args)
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return p, errors.EntityNotFound(errors.EntityPlayer, err)
		}
		return p, errors.Internal(err)
	}

	return p, nil
}

func (c *Controller) UpdatePlayer(ctx context.Context, tp token.Payload, displayName string) (Player, error) {
	p, err := c.store.UpdatePlayer(ctx, store.UpdatePlayerParams{
		DisplayName: displayName,
		PlayerID:    tp.PlayerID,
	})
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return p, errors.EntityNotFound(errors.EntityPlayer, err)
		}
		return p, errors.Internal(err)
	}

	return p, nil
}
