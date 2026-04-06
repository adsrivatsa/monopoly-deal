package main

import (
	"fmt"
	"monopoly-deal/internal/config"
	"monopoly-deal/internal/service"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
)

type Server struct {
	cfg        config.Config
	controller *service.Controller
	router     *chi.Mux
}

func NewServer(cfg config.Config) *Server {
	s := &Server{
		cfg:        cfg,
		controller: service.NewController(cfg),
	}

	s.addRoutes()

	return s
}

func (s *Server) addRoutes() {
	router := chi.NewRouter()

	router.Use(
		render.SetContentType(render.ContentTypeJSON),
		middleware.StripSlashes,
		middleware.Recoverer,
		middleware.Heartbeat("/ping"),
		middleware.DefaultLogger,
		cors.Handler(cors.Options{
			AllowedOrigins:   []string{"https://*", "http://*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: true,
			MaxAge:           300,
		}),
	)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		Write(w, http.StatusOK, "success")
	})

	s.router = router
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%s", s.cfg.BackendPort)
	srv := &http.Server{
		Addr:    addr,
		Handler: s.router,
	}
	return srv.ListenAndServe()
}
