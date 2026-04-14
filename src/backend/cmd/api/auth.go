package main

import (
	"context"
	"monopoly-deal/internal/errors"
	"monopoly-deal/internal/service"
	"monopoly-deal/internal/token"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/markbates/goth/gothic"
)

func (s *Server) authRoutes() *chi.Mux {
	router := chi.NewRouter()

	router.Get("/{"+PROVIDER+"}/login", s.Login)
	router.Get("/{"+PROVIDER+"}/callback", s.LoginCallback)
	router.Get("/{"+PROVIDER+"}/logout", s.Logout)

	router.Route("/refresh", func(r chi.Router) {
		r.Use(tokenMiddleware(s.tokenMaker, s.cookieStore, s.sessionName, token.RefreshToken))
		r.Get("/", s.Refresh)
	})

	return router
}

func (s *Server) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	provider := chi.URLParam(r, PROVIDER)
	r = r.WithContext(context.WithValue(ctx, PROVIDER, provider))

	sess, _ := s.cookieStore.Get(r, s.sessionName)
	accessToken, ok1 := sess.Values[token.AccessToken].(string)
	refreshToken, ok2 := sess.Values[token.RefreshToken].(string)
	if !ok1 || !ok2 || accessToken == "" || refreshToken == "" {
		gothic.BeginAuthHandler(w, r)
		return
	}

	tp, err1 := s.tokenMaker.VerifyToken(accessToken, token.AccessToken)
	_, err2 := s.tokenMaker.VerifyToken(refreshToken, token.RefreshToken)
	if err1 != nil && err2 != nil {
		gothic.BeginAuthHandler(w, r)
		return
	}

	_, err := s.services.GetPlayer(ctx, tp, service.GetPlayerParams{})
	if err == nil {
		fullURL, err := url.JoinPath(s.cfg.FrontendDomain, s.cfg.FrontendLobbyRoute)
		if err != nil {
			ErrorHTTP(w, errors.Internal(err))
		}

		http.Redirect(w, r, fullURL, http.StatusFound)
		return
	}

	s.Logout(w, r)
	s.Login(w, r)
}

func (s *Server) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	provider := chi.URLParam(r, PROVIDER)
	r = r.WithContext(context.WithValue(ctx, PROVIDER, provider))

	err := gothic.Logout(w, r)
	if err != nil {
		ErrorHTTP(w, errors.Internal(err))
		return
	}

	sess, _ := s.cookieStore.Get(r, s.sessionName)
	sess.Values[token.AccessToken] = ""
	sess.Values[token.RefreshToken] = ""
	err = sess.Save(r, w)
	if err != nil {
		ErrorHTTP(w, errors.Internal(err))
		return
	}

	fullURL, err := url.JoinPath(s.cfg.FrontendDomain, s.cfg.FrontendLoginRoute)
	if err != nil {
		ErrorHTTP(w, errors.Internal(err))
		return
	}
	http.Redirect(w, r, fullURL, http.StatusFound)
}

func (s *Server) LoginCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	provider := chi.URLParam(r, PROVIDER)
	r = r.WithContext(context.WithValue(r.Context(), PROVIDER, provider))
	oauthUser, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		ErrorHTTP(w, errors.Unauthenticated(err))
		return
	}

	args := service.GetPlayerParams{Email: &oauthUser.Email}
	p, err := s.services.GetPlayer(ctx, token.Payload{}, args)
	if err != nil {
		// create user
		args := service.CreatePlayerParams{
			DisplayName: oauthUser.Name,
			Email:       oauthUser.Email,
			ImageUrl:    oauthUser.AvatarURL,
		}
		p, err = s.services.CreatePlayer(ctx, args)
		if err != nil {
			ErrorHTTP(w, err)
			return
		}
	}

	tp := token.Payload{
		PlayerID: p.PlayerID,
	}
	accessToken, _, err := s.tokenMaker.CreateToken(tp, token.AccessToken)
	if err != nil {
		ErrorHTTP(w, err)
		return
	}

	tp.TokenID = p.RefreshTokenID
	refreshToken, _, err := s.tokenMaker.CreateToken(tp, token.RefreshToken)
	if err != nil {
		ErrorHTTP(w, err)
		return
	}

	sess, _ := s.cookieStore.Get(r, s.sessionName)
	sess.Values[token.AccessToken] = accessToken
	sess.Values[token.RefreshToken] = refreshToken
	err = sess.Save(r, w)
	if err != nil {
		ErrorHTTP(w, errors.Internal(err))
		return
	}

	fullUrl, err := url.JoinPath(s.cfg.FrontendDomain, s.cfg.FrontendLobbyRoute)
	if err != nil {
		ErrorHTTP(w, errors.Internal(err))
		return
	}

	http.Redirect(w, r, fullUrl, http.StatusFound)
}

func (s *Server) Refresh(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tp, err := tokenFromRequest(r, token.RefreshToken)
	if err != nil {
		ErrorHTTP(w, err)
		return
	}

	p, err := s.services.GetPlayer(ctx, tp, service.GetPlayerParams{})
	if err != nil {
		ErrorHTTP(w, err)
		return
	}

	if p.RefreshTokenID != tp.TokenID {
		ErrorHTTP(w, errors.InvalidToken)
		return
	}

	tp.TokenID = uuid.Nil
	accessToken, _, err := s.tokenMaker.CreateToken(tp, token.AccessToken)
	if err != nil {
		ErrorHTTP(w, err)
		return
	}

	sess, _ := s.cookieStore.Get(r, s.sessionName)
	sess.Values[token.AccessToken] = accessToken
	err = sess.Save(r, w)
	if err != nil {
		ErrorHTTP(w, errors.Internal(err))
		return
	}

	WriteHTTP(w, http.StatusOK, nil)
}
