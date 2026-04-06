package service

import "monopoly-deal/internal/config"

type Controller struct {
	cfg config.Config
}

func NewController(cfg config.Config) *Controller {
	c := &Controller{
		cfg: cfg,
	}

	return c
}

// func (c Controller)
