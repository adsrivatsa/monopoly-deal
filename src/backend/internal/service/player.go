package service

import (
	"context"
	"the-deal/internal/errors"
	"the-deal/internal/store"
	"the-deal/internal/token"

	"github.com/google/uuid"
)

func (c *Controller) CreatePlayer(ctx context.Context, args CreatePlayerParams) (Player, error) {
	playerID, err := uuid.NewV7()
	if err != nil {
		return Player{}, err
	}

	refreshTokenID, err := uuid.NewRandom()
	if err != nil {
		return Player{}, err
	}

	p, err := c.store.CreatePlayer(ctx, store.CreatePlayerParams{
		PlayerID:       playerID,
		DisplayName:    args.DisplayName,
		Email:          args.Email,
		ImageUrl:       args.ImageUrl,
		RefreshTokenID: refreshTokenID,
	})
	if err != nil {
		if errors.DBErrorCode(err) == errors.NoDataFound {
			return p, errors.EntityNotFound(errors.EntityPlayer, err)
		}
		return p, err
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
		return p, err
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
		return p, err
	}

	return p, nil
}
