package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"the-deal/internal/config"
	"the-deal/internal/schema"
	"the-deal/internal/token"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

// Client is a scripted stand-in for a browser: it holds a signed session
// cookie (minted with the same secret + gorilla cookie store the server uses)
// and speaks HTTP + protobuf websockets exactly like the frontend.
type Client struct {
	name     string
	playerID uuid.UUID
	cookie   *http.Cookie
	baseURL  string
	wsBase   string

	http *http.Client
}

const (
	sessionName = "session"
)

// mintCookie builds a gorilla session cookie carrying the access-token JWT the
// same way the server's tokenMiddleware expects to read it.
func mintCookie(cfg config.Config, maker token.Maker, playerID uuid.UUID) (*http.Cookie, error) {
	access, _, err := maker.CreateToken(token.Payload{
		TokenID:  uuid.New(),
		PlayerID: playerID,
	}, token.AccessToken)
	if err != nil {
		return nil, err
	}

	store := sessions.NewCookieStore([]byte(cfg.CookieSecret))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   100 * 365 * 24 * 60 * 60,
		HttpOnly: true,
		Secure:   cfg.IsProduction,
		SameSite: http.SameSiteLaxMode,
	}

	// Encode the session via a throwaway request/recorder so we reuse the
	// exact securecookie codecs the server uses.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	sess, _ := store.Get(req, sessionName)
	sess.Values[token.AccessToken] = access

	rec := httptest.NewRecorder()
	if err := sess.Save(req, rec); err != nil {
		return nil, err
	}

	setCookie := rec.Result().Cookies()
	if len(setCookie) == 0 {
		return nil, fmt.Errorf("no cookie produced")
	}
	return setCookie[0], nil
}

func NewClient(name string, cfg config.Config, maker token.Maker, playerID uuid.UUID) (*Client, error) {
	cookie, err := mintCookie(cfg, maker, playerID)
	if err != nil {
		return nil, err
	}
	return &Client{
		name:     name,
		playerID: playerID,
		cookie:   cookie,
		baseURL:  fmt.Sprintf("http://127.0.0.1:%d", cfg.BackendPort),
		wsBase:   fmt.Sprintf("ws://127.0.0.1:%d", cfg.BackendPort),
		http:     &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (c *Client) doJSON(method, path string, body any) ([]byte, int, error) {
	var rdr io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return nil, 0, err
		}
		rdr = bytes.NewReader(buf)
	}
	req, err := http.NewRequest(method, c.baseURL+path, rdr)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(c.cookie)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	return out, resp.StatusCode, nil
}

// GetStatic fetches a static asset path (used to spot-check card image URLs).
func (c *Client) GetStatic(path string) (int, string, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return 0, "", err
	}
	req.AddCookie(c.cookie)
	resp, err := c.http.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	return resp.StatusCode, resp.Header.Get("Content-Type"), nil
}

// Socket is a websocket connection speaking the top-level gateway proto.
type Socket struct {
	name  string
	conn  *websocket.Conn
	mu    sync.Mutex
	inbox chan *schema.ServerMessage
	ctx   context.Context
	stop  context.CancelFunc
}

func (c *Client) dialSocket(path string) (*Socket, error) {
	u, err := url.Parse(c.wsBase + path)
	if err != nil {
		return nil, err
	}
	header := http.Header{}
	header.Set("Cookie", c.cookie.String())
	header.Set("Origin", "http://127.0.0.1")

	conn, resp, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		if resp != nil {
			b, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("dial %s: %w (status %d: %s)", path, err, resp.StatusCode, string(b))
		}
		return nil, fmt.Errorf("dial %s: %w", path, err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	s := &Socket{
		name:  c.name,
		conn:  conn,
		inbox: make(chan *schema.ServerMessage, 256),
		ctx:   ctx,
		stop:  cancel,
	}
	go s.readLoop()
	return s, nil
}

func (s *Socket) readLoop() {
	for {
		_, data, err := s.conn.ReadMessage()
		if err != nil {
			close(s.inbox)
			return
		}
		msg := &schema.ServerMessage{}
		if err := proto.Unmarshal(data, msg); err != nil {
			continue
		}
		select {
		case s.inbox <- msg:
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *Socket) sendClient(msg *schema.ClientMessage) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.conn.WriteMessage(websocket.BinaryMessage, data)
}

func (s *Socket) Close() {
	s.stop()
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = s.conn.Close()
}

// nextMatching waits up to timeout for a server message satisfying pred,
// draining (and ignoring) any others. Pings are ignored automatically.
func (s *Socket) nextMatching(timeout time.Duration, pred func(*schema.ServerMessage) bool) (*schema.ServerMessage, error) {
	deadline := time.After(timeout)
	for {
		select {
		case msg, ok := <-s.inbox:
			if !ok {
				return nil, fmt.Errorf("[%s] socket closed while waiting", s.name)
			}
			if msg.GetPing() != nil {
				continue
			}
			if pred(msg) {
				return msg, nil
			}
		case <-deadline:
			return nil, fmt.Errorf("[%s] timeout waiting for message", s.name)
		}
	}
}

// drain empties the inbox of any currently buffered messages.
func (s *Socket) drain() {
	for {
		select {
		case <-s.inbox:
		default:
			return
		}
	}
}

func short(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

func trunc(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return strings.TrimSpace(s)
}
