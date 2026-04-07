package main

import (
	"fmt"
	"monopoly-deal/internal/config"
	"monopoly-deal/internal/service"
	"monopoly-deal/internal/token"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

const (
	PROVIDER = "provider"
)

type Server struct {
	cfg         config.Config
	controller  *service.Controller
	router      *chi.Mux
	tokenMaker  token.Maker
	cookieStore *sessions.CookieStore
	sessionName string
}

func NewServer(cfg config.Config, pool *pgxpool.Pool) *Server {
	initialiseGoth(cfg)

	sessionName := "session"

	cookieStore := sessions.NewCookieStore([]byte(cfg.CookieSecret))
	cookieStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   100 * 365 * 24 * 60 * 60,
		HttpOnly: true,
		Secure:   cfg.IsProduction,
	}

	durations := map[token.TokenType]time.Duration{
		token.AccessToken:  cfg.AccessTokenDuration,
		token.RefreshToken: cfg.RefreshTokenDuration,
	}
	tokenMaker := token.NewMaker(durations, cfg.CookieSecret)

	s := &Server{
		cfg:         cfg,
		controller:  service.NewController(cfg, pool),
		tokenMaker:  tokenMaker,
		cookieStore: cookieStore,
		sessionName: sessionName,
	}

	s.addRoutes()

	return s
}

func initialiseGoth(cfg config.Config) {
	store := sessions.NewCookieStore([]byte(cfg.CookieSecret))
	store.Options = &sessions.Options{
		Path:     "/",
		HttpOnly: true,
		Secure:   cfg.IsProduction,
	}
	gothic.Store = store
	goth.UseProviders(
		google.New(cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.GoogleClientRedirect, "profile", "email"),
	)
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

	router.Mount("/auth", s.authRoutes())

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
