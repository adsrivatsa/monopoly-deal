package main

import (
	"monopoly-deal/internal/service"
	"monopoly-deal/internal/token"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *Server) playerRoutes() *chi.Mux {
	router := chi.NewRouter()

	router.Post("/", s.GetPlayer)

	return router
}

func (s *Server) GetPlayer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tp, err := tokenFromRequest(r, token.AccessToken)
	if err != nil {
		ErrorHTTP(w, err)
		return
	}

	args, err := ReadAndValidate[GetPlayerParams](w, r)
	if err != nil {
		ErrorHTTP(w, err)
		return
	}

	p, err := s.services.GetPlayer(ctx, tp, service.GetPlayerParams{
		PlayerID: args.PlayerID,
		Email:    args.Email,
	})
	if err != nil {
		ErrorHTTP(w, err)
		return
	}

	WriteHTTP(w, http.StatusOK, Player{
		PlayerID:    p.PlayerID,
		DisplayName: p.DisplayName,
		Email:       p.Email,
		ImageUrl:    p.ImageUrl,
	})
}
