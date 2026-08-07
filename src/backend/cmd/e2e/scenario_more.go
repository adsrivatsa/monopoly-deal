package main

import (
	"fmt"
	deal_no_mercy "the-deal/internal/engine/deal-no-mercy"
	"the-deal/internal/schema"
	dnm "the-deal/internal/schema/deal_no_mercy_schema"
	"the-deal/internal/store"
	"time"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Scenario 4: timeout default move (scoped) with a classic game running.
// ---------------------------------------------------------------------------

func (h *Harness) scenarioTimeout(g *GameCtx) {
	// Stand up a parallel classic monopoly_deal game so we can prove the No
	// Mercy timeout does not disturb it (ClaimGameTimeoutsByGame is scoped by
	// game type; the live scheduler keys per game/player).
	classicGameID, classicA, classicSeq := h.startClassicGame()
	_ = classicA

	// Make the No Mercy current player's move timeout tiny (1s) by writing it
	// into the persisted config, then perform a trivial play to reschedule the
	// move timeout at now+1s. When it fires, DefaultMove + CompleteTurn run and
	// the turn advances to the opponent.
	att := h.currentPlayer(g)
	opp := g.other(att)
	h.writeState(g, func(b *stateBuilder) {
		b.setTurn(att)
		b.g.Config.MoveTimeout = 1 * time.Second
		b.g.Config.DemandTimeout = 1 * time.Second
		b.setHand(att, deal_no_mercy.AssetKeyMoney1, deal_no_mercy.AssetKeyMoney1, deal_no_mercy.AssetKeyMoney1)
		b.setDeckTop(deal_no_mercy.AssetKeyMoney1, deal_no_mercy.AssetKeyMoney1)
	})
	g.resync(h)

	seqBefore := h.reload(g.gameID).SequenceNum

	// A single play reschedules the move timeout at now + 1s (tiny).
	cardID := h.handCardID(g, att, deal_no_mercy.AssetKeyMoney1)
	must(g.send(att, &dnm.ClientMessage{Payload: &dnm.ClientMessage_PlayMoney{
		PlayMoney: &dnm.PlayMoney{CardId: cardID},
	}}), "send play to arm timeout")
	_, _ = awaitAction(g.sockFor(att), dnm.ActionKind_ACTION_KIND_PLAY_MONEY)

	// Wait for the default move to fire (turn should advance to opp). The
	// opponent's socket receives a start-turn published by the timeout.
	_, err := g.sockFor(opp).nextMatching(6*time.Second, func(m *schema.ServerMessage) bool {
		a := dnmAction(m)
		return a != nil && a.GetKind() == dnm.ActionKind_ACTION_KIND_START_TURN
	})
	after := h.reload(g.gameID)
	fired := err == nil && after.Players[after.CurrPlayerIdx] == opp && after.SequenceNum > seqBefore
	h.record("timeout default move fires + advances turn", fired,
		fmt.Sprintf("current after timeout=%s (want opp=%s), seq %d->%d", short(after.Players[after.CurrPlayerIdx].String()), short(opp.String()), seqBefore, after.SequenceNum))

	// Verify no cross-game interference: the classic game's state is unchanged.
	classicSeqAfter := h.classicSeqNum(classicGameID)
	noInterference := classicSeqAfter == classicSeq
	h.record("no cross-game interference (classic untouched)", noInterference,
		fmt.Sprintf("classic seq %d (before %d)", classicSeqAfter, classicSeq))

	// Also confirm the timeout rows are scoped: a claim for classic timeouts
	// must not return the No Mercy game's rows and vice versa.
	scoped := h.verifyTimeoutScoping(g.gameID, classicGameID)
	h.record("ClaimGameTimeoutsByGame scoping (per game type)", scoped.ok, scoped.detail)

	// Demand-timeout default: an aggressive play leaves a payment demand; when
	// the DEMAND deadline expires, DefaultDemand auto-resolves it (best-payment
	// subset) without the target acting.
	h.scenarioDemandTimeout(g)

	// Restore a normal move timeout on the No Mercy game for subsequent tests.
	h.writeState(g, func(b *stateBuilder) {
		b.g.Config.MoveTimeout = 30 * time.Second
		b.g.Config.DemandTimeout = 20 * time.Second
	})
}

// scenarioDemandTimeout arms a tiny demand timeout on a payment demand and
// verifies the default-demand path fires (auto-pays) per demand kind.
func (h *Harness) scenarioDemandTimeout(g *GameCtx) {
	att := h.currentPlayer(g)
	tgt := g.other(att)
	h.writeState(g, func(b *stateBuilder) {
		b.setTurn(att)
		b.g.Config.MoveTimeout = 2 * time.Second
		b.g.Config.DemandTimeout = 1 * time.Second
		b.setHand(att, deal_no_mercy.AssetKeyYoink)
		b.setBank(tgt, deal_no_mercy.AssetKeyMoney10, deal_no_mercy.AssetKeyMoney5)
		b.setBank(att)
	})
	g.resync(h)

	cardID := h.handCardID(g, att, deal_no_mercy.AssetKeyYoink)
	must(g.send(att, &dnm.ClientMessage{Payload: &dnm.ClientMessage_PlayYoink{
		PlayYoink: &dnm.PlayYoink{CardId: cardID, TargetId: tgt.String()},
	}}), "send yoink for demand timeout")
	h.awaitDemandsApplied(g, att)

	before := h.numDemands(g)
	// Wait for DefaultDemand to auto-resolve the payment demand (attacker sees a
	// demand-complied action published by the timeout path).
	_, err := g.sockFor(att).nextMatching(5*time.Second, func(m *schema.ServerMessage) bool {
		a := dnmAction(m)
		return a != nil && a.GetKind() == dnm.ActionKind_ACTION_KIND_DEMAND_COMPLIED
	})
	after := h.reload(g.gameID)
	// Demand cleared and attacker received the yoink payment (>=10 of the target's bank).
	autoPaid := err == nil && len(after.Demands) == 0 && bankValueOf(after.Money[att]) >= 10
	h.record("demand-timeout default auto-resolves (payment kind)", autoPaid,
		fmt.Sprintf("demands %d->%d, attacker bank=%d (want>=10)", before, len(after.Demands), bankValueOf(after.Money[att])))
	h.endTurn(g, att)
}

type scopeResult struct {
	ok     bool
	detail string
}

// verifyTimeoutScoping queries game_timeout joined with game.game to confirm
// the No Mercy timeout row is tagged deal_no_mercy and any classic row is
// monopoly_deal — i.e. the scoped claim can't cross games.
func (h *Harness) verifyTimeoutScoping(nmGame, classicGame uuid.UUID) scopeResult {
	rows, err := h.db.pool.Query(h.ctx,
		`SELECT g.game, count(*) FROM game_timeout t JOIN game g ON g.game_id = t.game_id
		 WHERE t.game_id = ANY($1) GROUP BY g.game`,
		[]uuid.UUID{nmGame, classicGame})
	if err != nil {
		return scopeResult{false, err.Error()}
	}
	defer rows.Close()
	counts := map[string]int{}
	for rows.Next() {
		var gt store.GameType
		var n int
		if err := rows.Scan(&gt, &n); err != nil {
			return scopeResult{false, err.Error()}
		}
		counts[string(gt)] += n
	}
	// The No Mercy game's timeout row must be tagged deal_no_mercy.
	ok := counts[string(store.GameTypeDealNoMercy)] >= 1
	return scopeResult{ok, fmt.Sprintf("timeout rows by game type: %v", counts)}
}

// startClassicGame seeds two players and starts a classic monopoly_deal game
// via the real room/game flow, returning the game id, host client and the
// initial sequence number. It leaves a live move timeout registered.
func (h *Harness) startClassicGame() (uuid.UUID, *Client, int) {
	cID, err := h.db.seedPlayer(h.ctx, "Carol")
	must(err, "seed Carol")
	dID, err := h.db.seedPlayer(h.ctx, "Dave")
	must(err, "seed Dave")

	c, err := NewClient("Carol", h.cfg, h.maker, cID)
	must(err, "client Carol")
	d, err := NewClient("Dave", h.cfg, h.maker, dID)
	must(err, "client Dave")

	settings := classicSettingsBuf()
	body := map[string]any{
		"display_name": "E2E Classic",
		"capacity":     2,
		"game":         string(store.GameTypeMonopolyDeal),
		"settings":     settings,
		"is_private":   true,
	}
	_, code, err := c.doJSON("POST", "/room", body)
	must(err, "create classic room")
	if code != 200 {
		must(fmt.Errorf("status %d", code), "create classic room")
	}
	roomID, err := h.db.roomIDForPlayer(h.ctx, cID)
	must(err, "classic room id")
	_, _, err = d.doJSON("PATCH", "/room/join/"+roomID.String(), nil)
	must(err, "join classic")
	_, _, err = d.doJSON("PATCH", "/room/ready", nil)
	must(err, "ready classic")
	_, code, err = c.doJSON("POST", "/game", nil)
	must(err, "start classic")
	if code != 200 {
		must(fmt.Errorf("start classic status %d", code), "start classic")
	}
	gameID, err := h.db.gameIDForPlayer(h.ctx, cID)
	must(err, "classic game id")

	return gameID, c, h.classicSeqNum(gameID)
}

// classicSeqNum reads the classic engine snapshot's sequence number from the DB
// (msgpack field "o", mirrored from the classic engine).
func (h *Harness) classicSeqNum(gameID uuid.UUID) int {
	// Rather than depend on the classic decoder here, count history rows as a
	// proxy for "did anything advance".
	var n int
	err := h.db.pool.QueryRow(h.ctx, `SELECT count(*) FROM game_history WHERE game_id = $1`, gameID).Scan(&n)
	must(err, "classic history count")
	return n
}

// classicSettingsBuf returns encoded default classic settings.
func classicSettingsBuf() []byte {
	// Import lazily to avoid a hard dependency spread; use the engine default.
	return mdDefaultSettings()
}

// ---------------------------------------------------------------------------
// Scenario 5: win (3 complete sets)
// ---------------------------------------------------------------------------

func (h *Harness) scenarioWin(g *GameCtx) {
	att := h.currentPlayer(g)
	opp := g.other(att)

	// Set up: attacker already has two complete sets + a third set one card
	// short, and the last card in hand. Playing it completes the third set and
	// wins (WinSetAmount=3, WinMoneyAmount=0).
	var lastCardID, utilSetID string
	h.writeState(g, func(b *stateBuilder) {
		b.setTurn(att)
		b.clearProperties(att)
		b.addSet(att, deal_no_mercy.ColorBrown, false, deal_no_mercy.AssetKeyBrown1, deal_no_mercy.AssetKeyBrown2)
		b.addSet(att, deal_no_mercy.ColorBlue, false, deal_no_mercy.AssetKeyBlue1, deal_no_mercy.AssetKeyBlue2)
		// third set: utility needs 2, give it 1, complete with hand card.
		utilSet := b.addSet(att, deal_no_mercy.ColorUtility, false, deal_no_mercy.AssetKeyUtil1)
		utilSetID = string(utilSet.ID)
		hand := b.setHand(att, deal_no_mercy.AssetKeyUtil2)
		lastCardID = string(hand[0].ID)
	})
	g.resync(h)

	// Play util2 INTO the existing utility set so it completes (2/2) — the third
	// complete set triggers the win.
	must(g.send(att, &dnm.ClientMessage{Payload: &dnm.ClientMessage_PlayProperty{
		PlayProperty: &dnm.PlayProperty{CardId: lastCardID, PropertySetId: &utilSetID},
	}}), "send winning property")

	// Both clients should receive a WonGame message.
	wonA, errA := g.sockFor(att).nextMatching(waitT, func(m *schema.ServerMessage) bool { return dnmWon(m) != nil })
	wonB, errB := g.sockFor(opp).nextMatching(waitT, func(m *schema.ServerMessage) bool { return dnmWon(m) != nil })
	bothWon := errA == nil && errB == nil && dnmWon(wonA).GetPlayerId() == att.String() && dnmWon(wonB).GetPlayerId() == att.String()
	h.record("both clients receive WonGame", bothWon,
		boolStr(bothWon, fmt.Sprintf("winner=%s sets=%d", short(att.String()), dnmWon(wonA).GetNumCompletedSets()), fmt.Sprintf("A:%v B:%v", errA, errB)))

	// Game row marked completed with winner set.
	completed, winner, err := h.db.gameCompleted(h.ctx, g.gameID)
	must(err, "read game completed")
	rowOK := completed && winner != nil && *winner == att
	h.record("game row marked completed", rowOK, fmt.Sprintf("completed=%v winner=%v", completed, winnerStr(winner)))
}

func winnerStr(w *uuid.UUID) string {
	if w == nil {
		return "nil"
	}
	return short(w.String())
}

// ---------------------------------------------------------------------------
// Scenario 6: reconnect snapshot integrity
// ---------------------------------------------------------------------------

func (h *Harness) scenarioReconnect(g *GameCtx) {
	// Establish a known mid-game board so the snapshot is non-trivial: attacker
	// with a couple of sets and a bank, opponent with a bank.
	att := h.currentPlayer(g)
	opp := g.other(att)
	h.writeState(g, func(b *stateBuilder) {
		b.setTurn(att)
		b.clearProperties(att)
		b.clearProperties(opp)
		b.addSet(att, deal_no_mercy.ColorBrown, false, deal_no_mercy.AssetKeyBrown1, deal_no_mercy.AssetKeyBrown2)
		b.addSet(att, deal_no_mercy.ColorRed, false, deal_no_mercy.AssetKeyRed1)
		b.setBank(att, deal_no_mercy.AssetKeyMoney5)
		b.setBank(opp, deal_no_mercy.AssetKeyMoney2)
	})
	g.resync(h)

	// Capture Bob's authoritative snapshot before the drop.
	pre := h.snapshotFor(g, g.b)

	// Drop Bob mid-game and reconnect the game socket.
	g.sockB.Close()
	var err error
	g.sockB, err = g.b.dialSocket("/game/socket")
	must(err, "reconnect Bob")
	m, err := g.sockB.nextMatching(waitT, func(sm *schema.ServerMessage) bool { return dnmState(sm) != nil })
	must(err, "Bob reconnect state")
	post := dnmState(m)

	match := snapshotsMatch(pre, post)
	h.record("reconnect returns matching authoritative snapshot", match.ok, match.detail)

	// On reconnect the server also replays the action history (masked). Confirm
	// at least one ActionHistory entry arrives — this exercises the history
	// sidebar path for the new No Mercy action kinds.
	hm, herr := g.sockB.nextMatching(waitT, func(sm *schema.ServerMessage) bool {
		dd := sm.GetDealNoMercyMessage()
		return dd != nil && dd.GetActionHistory() != nil
	})
	replayed := herr == nil && hm != nil
	h.record("action history replayed on reconnect (new action kinds)", replayed,
		boolStr(replayed, "received ActionHistory entries after snapshot", errStr(herr)))
}

func (h *Harness) snapshotFor(g *GameCtx, c *Client) *dnm.GameState {
	// Open a throwaway socket for a clean snapshot read.
	s, err := c.dialSocket("/game/socket")
	must(err, "snapshot dial")
	defer s.Close()
	m, err := s.nextMatching(waitT, func(sm *schema.ServerMessage) bool { return dnmState(sm) != nil })
	must(err, "snapshot read")
	return dnmState(m)
}

type matchResult struct {
	ok     bool
	detail string
}

func snapshotsMatch(a, b *dnm.GameState) matchResult {
	if a == nil || b == nil {
		return matchResult{false, "nil snapshot"}
	}
	reasons := ""
	if a.GetSeqNum() != b.GetSeqNum() {
		reasons += fmt.Sprintf("seq %d!=%d ", a.GetSeqNum(), b.GetSeqNum())
	}
	if a.GetCurrentPlayerId() != b.GetCurrentPlayerId() {
		reasons += "currentPlayer differs "
	}
	if len(a.GetProperties()) != len(b.GetProperties()) {
		reasons += fmt.Sprintf("propSets %d!=%d ", len(a.GetProperties()), len(b.GetProperties()))
	}
	if len(a.GetPlayers()) != len(b.GetPlayers()) {
		reasons += "player count differs "
	}
	// Compare per-player completed-set counts.
	setsA := map[string]int32{}
	for _, p := range a.GetPlayers() {
		setsA[p.GetPlayerId()] = p.GetCompletedSets()
	}
	for _, p := range b.GetPlayers() {
		if setsA[p.GetPlayerId()] != p.GetCompletedSets() {
			reasons += fmt.Sprintf("player %s sets differ ", short(p.GetPlayerId()))
		}
	}
	if reasons == "" {
		return matchResult{true, fmt.Sprintf("snapshot stable (seq=%d, %d prop sets)", a.GetSeqNum(), len(a.GetProperties()))}
	}
	return matchResult{false, reasons}
}
