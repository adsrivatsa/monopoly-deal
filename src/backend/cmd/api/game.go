package main

import (
	"fun-kames/internal/store"
	"fun-kames/internal/token"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *Server) gameRoutes() *chi.Mux {
	router := chi.NewRouter()

	router.Post("/", s.CreateGame)
	router.Get("/socket", s.GameSocket)

	return router
}

func (s *Server) CreateGame(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tp, err := tokenFromRequest(r, token.AccessToken)
	if err != nil {
		ErrorHTTP(w, err)
		return
	}

	err = s.services.CreateGame(ctx, tp)
	if err != nil {
		ErrorHTTP(w, err)
		return
	}

	WriteHTTP(w, http.StatusOK, nil)
}

func (s *Server) GameSocket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tp, err := tokenFromRequest(r, token.AccessToken)
	if err != nil {
		ErrorHTTP(w, err)
		return
	}

	g, err := s.services.GetGame(ctx, tp)
	if err != nil {
		ErrorHTTP(w, err)
		return
	}

	switch g.Game {
	case store.GameTypeMonopolyDeal:
		s.MonopolyDealSocket(w, r)
	}
}
