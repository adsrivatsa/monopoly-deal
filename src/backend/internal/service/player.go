package service

import (
	"context"
	"monopoly-deal/internal/errors"
	"monopoly-deal/internal/store"
	"monopoly-deal/internal/token"
)

type Player = store.Player

type CreatePlayerParams = store.CreatePlayerParams

var createPlayerErrMap = map[errors.DBViolation]func(err error) errors.Error{
	errors.UniqueViolation: errors.EntityAlreadyExists(errors.EntityPlayer),
}

func (c *Controller) CreatePlayer(ctx context.Context, args CreatePlayerParams) (Player, error) {
	p, err := c.store.CreatePlayer(ctx, args)
	return p, errors.DBError(createPlayerErrMap, err)
}

type GetPlayerParams = store.GetPlayerParams

var getPlayerErrMap = map[errors.DBViolation]func(err error) errors.Error{
	errors.NoDataFound: errors.EntityNotFound(errors.EntityPlayer),
}

func (c *Controller) GetPlayer(ctx context.Context, tp token.Payload, args GetPlayerParams) (Player, error) {
	if args.PlayerID == nil && args.Email == nil {
		args.PlayerID = &tp.PlayerID
	}
	p, err := c.store.GetPlayer(ctx, args)
	return p, errors.DBError(getPlayerErrMap, err)
}

var updatePlayerErrMap = map[errors.DBViolation]func(err error) errors.Error{
	errors.NoDataFound: errors.EntityNotFound(errors.EntityPlayer),
}

func (c *Controller) UpdatePlayer(ctx context.Context, tp token.Payload, name string) (Player, error) {
	args := store.UpdatePlayerParams{
		DisplayName: name,
		PlayerID:    tp.PlayerID,
	}
	p, err := c.store.UpdatePlayer(ctx, args)
	return p, errors.DBError(updatePlayerErrMap, err)
}
