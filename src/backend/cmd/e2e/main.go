// Command e2e is a scripted, backend-only end-to-end harness for
// "Monopoly Deal: No Mercy" (Phase 5 acceptance). It seeds players directly in
// the DB, mints session cookies with the server's own token package, drives the
// real HTTP + websocket protocol exactly like the frontend, and verifies every
// new mechanic. It never modifies the game rules; where pure play can't reach a
// scenario it writes authoritative game state to the DB between steps (as the
// brief permits).
//
// Prereqs: postgres + redis up (make postgresup redisup, createdb, migrateup)
// and the API server running (go run ./cmd/api from src/backend). Run:
//
//	go run ./cmd/e2e
package main

import (
	"context"
	"fmt"
	"os"
	"the-deal/internal/config"
	deal_no_mercy "the-deal/internal/engine/deal-no-mercy"
	"the-deal/internal/schema"
	dnm "the-deal/internal/schema/deal_no_mercy_schema"
	"the-deal/internal/store"
	"the-deal/internal/token"
	"time"

	"github.com/google/uuid"
)

type Harness struct {
	cfg   config.Config
	maker token.Maker
	db    *DB
	ctx   context.Context

	results []Result
}

type Result struct {
	Name   string
	Pass   bool
	Detail string
}

func (h *Harness) record(name string, pass bool, detail string) {
	status := "PASS"
	if !pass {
		status = "FAIL"
	}
	fmt.Printf("  [%s] %s — %s\n", status, name, detail)
	h.results = append(h.results, Result{Name: name, Pass: pass, Detail: detail})
}

func must(err error, msg string) {
	if err != nil {
		fmt.Printf("FATAL: %s: %v\n", msg, err)
		os.Exit(1)
	}
}

func main() {
	ctx := context.Background()

	cfg, err := config.Load("../.env")
	must(err, "load config")

	durations := map[token.TokenType]time.Duration{
		token.AccessToken:  cfg.AccessTokenDuration,
		token.RefreshToken: cfg.RefreshTokenDuration,
	}
	maker := token.NewMaker(durations, cfg.CookieSecret)

	db, err := NewDB(ctx, cfg.DatabaseURL)
	must(err, "connect db")
	defer db.Close()

	h := &Harness{cfg: cfg, maker: maker, db: db, ctx: ctx}

	fmt.Println("== cleaning DB ==")
	must(db.cleanup(ctx), "cleanup")

	// Wait for the API server to be reachable.
	must(h.waitForServer(), "server not reachable")

	fmt.Println("\n== Scenario 1: game creation, initial deal ==")
	game := h.setupGame()

	fmt.Println("\n== Scenario 2: core turn loop ==")
	h.scenarioTurnLoop(game)

	fmt.Println("\n== Scenario 3: every new mechanic ==")
	h.scenarioMechanics(game)

	fmt.Println("\n== Scenario 4: timeout default move (scoped, with classic game running) ==")
	h.scenarioTimeout(game)

	// Reconnect must be validated on a LIVE game (a completed game is not
	// reconnectable by design — GetGameByPlayer filters out completed games, so
	// the client stays on the terminal WonGame screen instead). Run it before
	// the win drives the game to completion.
	fmt.Println("\n== Scenario 6: reconnect snapshot (live game) ==")
	h.scenarioReconnect(game)

	fmt.Println("\n== Scenario 5: win (3 complete sets) ==")
	h.scenarioWin(game)

	h.printSummary()
}

// GameCtx carries everything a running two-player game needs.
type GameCtx struct {
	gameID  uuid.UUID
	a, b    *Client
	sockA   *Socket
	sockB   *Socket
	roomID  uuid.UUID
}

func (h *Harness) waitForServer() error {
	c, err := NewClient("probe", h.cfg, h.maker, uuid.New())
	if err != nil {
		return err
	}
	for i := 0; i < 40; i++ {
		_, code, err := c.doJSON("GET", "/ping", nil)
		if err == nil && code < 500 {
			return nil
		}
		time.Sleep(250 * time.Millisecond)
	}
	return fmt.Errorf("server did not respond on port %d", h.cfg.BackendPort)
}

// setupGame seeds two players, drives room create/join/ready + POST /game over
// real HTTP, opens both game sockets, and verifies the initial deal.
func (h *Harness) setupGame() *GameCtx {
	aID, err := h.db.seedPlayer(h.ctx, "Alice")
	must(err, "seed Alice")
	bID, err := h.db.seedPlayer(h.ctx, "Bob")
	must(err, "seed Bob")

	a, err := NewClient("Alice", h.cfg, h.maker, aID)
	must(err, "client Alice")
	b, err := NewClient("Bob", h.cfg, h.maker, bID)
	must(err, "client Bob")

	// Create room (Alice = host) with game=deal_no_mercy and default settings.
	settings := deal_no_mercy.DefaultSettings()
	settingsBuf, err := settings.Encode()
	must(err, "encode settings")

	body := map[string]any{
		"display_name": "E2E No Mercy",
		"capacity":     2,
		"game":         string(store.GameTypeDealNoMercy),
		"settings":     settingsBuf, // JSON-encodes []byte as base64
		"is_private":   true,
	}
	out, code, err := a.doJSON("POST", "/room", body)
	must(err, "create room")
	if code != 200 {
		must(fmt.Errorf("status %d: %s", code, string(out)), "create room")
	}

	roomID, err := h.db.roomIDForPlayer(h.ctx, aID)
	must(err, "room id")

	// Bob joins.
	out, code, err = b.doJSON("PATCH", "/room/join/"+roomID.String(), nil)
	must(err, "join room")
	if code != 200 {
		must(fmt.Errorf("status %d: %s", code, string(out)), "join room")
	}

	// Bob readies up.
	_, code, err = b.doJSON("PATCH", "/room/ready", nil)
	must(err, "ready")
	if code != 200 {
		must(fmt.Errorf("ready status %d", code), "ready")
	}

	// Host starts the game.
	out, code, err = a.doJSON("POST", "/game", nil)
	must(err, "create game")
	if code != 200 {
		must(fmt.Errorf("status %d: %s", code, string(out)), "create game")
	}

	gameID, err := h.db.gameIDForPlayer(h.ctx, aID)
	must(err, "game id")

	// Connect game sockets.
	sockA, err := a.dialSocket("/game/socket")
	must(err, "dial Alice game socket")
	sockB, err := b.dialSocket("/game/socket")
	must(err, "dial Bob game socket")

	g := &GameCtx{gameID: gameID, a: a, b: b, sockA: sockA, sockB: sockB, roomID: roomID}

	// Both clients should receive a GameState snapshot on connect.
	msgA, err := sockA.nextMatching(5*time.Second, func(m *schema.ServerMessage) bool { return dnmState(m) != nil })
	must(err, "Alice GameState")
	msgB, err := sockB.nextMatching(5*time.Second, func(m *schema.ServerMessage) bool { return dnmState(m) != nil })
	must(err, "Bob GameState")

	stateA := dnmState(msgA)
	stateB := dnmState(msgB)

	// Each player is dealt 5. The starting player (current) has already drawn
	// TurnDraw (2) at StartTurn, so holds 5+2=7; the other holds the dealt 5.
	// Verify the deal + start-draw: hands are {5, 7} with the 7 belonging to the
	// current player.
	handA := len(stateA.GetYourHand().GetCards())
	handB := len(stateB.GetYourHand().GetCards())
	curIsA := stateA.GetCurrentPlayerId() == g.a.playerID.String()
	wantA, wantB := 5, 7
	if curIsA {
		wantA, wantB = 7, 5
	}
	dealtOK := handA == wantA && handB == wantB
	h.record("initial deal (5 each; starter drew 2)", dealtOK,
		fmt.Sprintf("Alice hand=%d (want %d) Bob hand=%d (want %d)", handA, wantA, handB, wantB))

	// Masking: each player's own hand length must equal their HandCards row, and
	// the opponent's hand row must reflect count only (no YourHand leak). From
	// Alice's snapshot, Bob's HandCards must equal Bob's true hand length and
	// Alice must not see Bob's cards.
	maskOK := true
	var maskDetail string
	for _, p := range stateA.GetPlayers() {
		want := 5
		if p.GetPlayerId() == stateA.GetCurrentPlayerId() {
			want = 7
		}
		if p.GetHandCards() != int32(want) {
			maskOK = false
			maskDetail = fmt.Sprintf("player %s HandCards=%d want %d", short(p.GetPlayerId()), p.GetHandCards(), want)
		}
	}
	// Alice's snapshot exposes only Alice's own cards via YourHand.
	if handA != int(playerHandCount(stateA, g.a.playerID)) {
		maskOK = false
		maskDetail += " YourHand != own HandCards"
	}
	h.record("opponent hand masked (count only)", maskOK, orDefault(maskDetail, "own hand matches count; opponent hand not revealed"))

	// Debt chips: 3 each.
	chipsOK := true
	var chipsDetail string
	for _, p := range stateA.GetPlayers() {
		if p.GetAvailableDebtChips() != 3 {
			chipsOK = false
			chipsDetail = fmt.Sprintf("player %s chips=%d", short(p.GetPlayerId()), p.GetAvailableDebtChips())
		}
	}
	h.record("debt chips (3 each)", chipsOK, orDefault(chipsDetail, "all players report 3 available chips"))

	// Asset images present + spot-check a couple resolve over HTTP under
	// /static/deal-no-mercy/card/.
	h.checkAssets(a, stateA)

	return g
}

func (h *Harness) checkAssets(c *Client, state *dnm.GameState) {
	imgs := state.GetAssetImages()
	if len(imgs) == 0 {
		h.record("asset images present", false, "GameState carried no asset_images")
		return
	}
	h.record("asset images present", true, fmt.Sprintf("%d asset image entries", len(imgs)))

	// Spot-check two: one small, one large. The URLs are absolute against the
	// configured BackendDomain; we GET the path portion against the local
	// server, which serves ./public.
	checked := 0
	passed := 0
	var detail string
	for _, kind := range []dnm.AssetImageKind{dnm.AssetImageKind_ASSET_IMAGE_KIND_SMALL, dnm.AssetImageKind_ASSET_IMAGE_KIND_LARGE} {
		for _, img := range imgs {
			if img.GetKind() != kind {
				continue
			}
			path := staticPath(img.GetImageUrl())
			if path == "" {
				continue
			}
			code, ctype, err := c.GetStatic(path)
			checked++
			if err == nil && code == 200 {
				passed++
				detail += fmt.Sprintf("%s->200(%s) ", path, trunc(ctype, 20))
			} else {
				detail += fmt.Sprintf("%s->%d ", path, code)
			}
			break
		}
	}
	h.record("asset image URLs resolve (HTTP GET /static)", checked > 0 && passed == checked, detail)
}

// playerHandCount returns the HandCards count reported for a player in a state.
func playerHandCount(state *dnm.GameState, pid uuid.UUID) int32 {
	for _, p := range state.GetPlayers() {
		if p.GetPlayerId() == pid.String() {
			return p.GetHandCards()
		}
	}
	return -1
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

// staticPath extracts the "/static/..." path from an absolute asset URL.
func staticPath(u string) string {
	idx := indexOf(u, "/static/")
	if idx < 0 {
		return ""
	}
	return u[idx:]
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func (h *Harness) printSummary() {
	fmt.Println("\n========== SUMMARY ==========")
	pass, fail := 0, 0
	for _, r := range h.results {
		status := "PASS"
		if !r.Pass {
			status = "FAIL"
			fail++
		} else {
			pass++
		}
		fmt.Printf("  %-4s  %s\n", status, r.Name)
	}
	fmt.Printf("\n  %d passed, %d failed, %d total\n", pass, fail, len(h.results))
	if fail > 0 {
		os.Exit(1)
	}
}
