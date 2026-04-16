package main

import (
	"fun-kames/internal/service"
	"fun-kames/internal/token"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *Server) playerRoutes() *chi.Mux {
	router := chi.NewRouter()

	router.Post("/", s.GetPlayer)
	router.Put("/", s.UpdatePlayer)

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

func (s *Server) UpdatePlayer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tp, err := tokenFromRequest(r, token.AccessToken)
	if err != nil {
		ErrorHTTP(w, err)
		return
	}

	args, err := ReadAndValidate[UpdatePlayerParams](w, r)
	if err != nil {
		ErrorHTTP(w, err)
		return
	}

	p, err := s.services.UpdatePlayer(ctx, tp, args.DisplayName)
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
