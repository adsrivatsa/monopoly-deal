package main

import (
	"fmt"
	deal_no_mercy "the-deal/internal/engine/deal-no-mercy"
	"the-deal/internal/schema"
	dnm "the-deal/internal/schema/deal_no_mercy_schema"
	"time"

	"github.com/google/uuid"
)

const waitT = 4 * time.Second

// sockFor returns the socket for the given player.
func (g *GameCtx) sockFor(pid uuid.UUID) *Socket {
	if pid == g.a.playerID {
		return g.sockA
	}
	return g.sockB
}

// otherSock returns the opponent's socket.
func (g *GameCtx) otherSock(pid uuid.UUID) *Socket {
	if pid == g.a.playerID {
		return g.sockB
	}
	return g.sockA
}

// send emits a client message from the given player's socket.
func (g *GameCtx) send(pid uuid.UUID, m *dnm.ClientMessage) error {
	return g.sockFor(pid).sendClient(gw(m))
}

// awaitAction waits for an Action of the given kind on a socket.
func awaitAction(s *Socket, kind dnm.ActionKind) (*dnm.Action, error) {
	m, err := s.nextMatching(waitT, func(sm *schema.ServerMessage) bool {
		a := dnmAction(sm)
		return a != nil && a.GetKind() == kind
	})
	if err != nil {
		return nil, err
	}
	return dnmAction(m), nil
}

// awaitError waits for a dnm Error message.
func awaitError(s *Socket) (*dnm.Error, error) {
	m, err := s.nextMatching(waitT, func(sm *schema.ServerMessage) bool {
		return dnmErr(sm) != nil
	})
	if err != nil {
		return nil, err
	}
	return dnmErr(m), nil
}

// reload fetches a fresh authoritative GameState by reconnecting a throwaway
// read from the DB (used to inspect internal state deterministically).
func (h *Harness) reload(gameID uuid.UUID) *deal_no_mercy.Game {
	game, err := h.db.loadGame(h.ctx, gameID)
	must(err, "reload game")
	return game
}

// currentPlayer returns the uuid of whose turn it is per authoritative state.
func (h *Harness) currentPlayer(g *GameCtx) uuid.UUID {
	game := h.reload(g.gameID)
	return game.Players[game.CurrPlayerIdx]
}

// writeState decodes, mutates via fn, and persists the authoritative state.
func (h *Harness) writeState(g *GameCtx, fn func(b *stateBuilder)) {
	game := h.reload(g.gameID)
	b := newBuilder(game)
	fn(b)
	must(h.db.saveGame(h.ctx, g.gameID, game), "save state")
	// Drain any buffered socket noise so subsequent awaits are clean.
	g.sockA.drain()
	g.sockB.drain()
}

// resync makes both sockets re-request a fresh GameState (reconnect) so their
// view matches DB writes we made out of band. Returns the two states.
func (g *GameCtx) resync(h *Harness) (*dnm.GameState, *dnm.GameState) {
	// Reconnect both sockets to force an authoritative snapshot.
	g.sockA.Close()
	g.sockB.Close()
	var err error
	g.sockA, err = g.a.dialSocket("/game/socket")
	must(err, "redial Alice")
	g.sockB, err = g.b.dialSocket("/game/socket")
	must(err, "redial Bob")
	mA, err := g.sockA.nextMatching(waitT, func(m *schema.ServerMessage) bool { return dnmState(m) != nil })
	must(err, "Alice resync state")
	mB, err := g.sockB.nextMatching(waitT, func(m *schema.ServerMessage) bool { return dnmState(m) != nil })
	must(err, "Bob resync state")
	return dnmState(mA), dnmState(mB)
}

// ---------------------------------------------------------------------------
// Scenario 2: core turn loop
// ---------------------------------------------------------------------------

func (h *Harness) scenarioTurnLoop(g *GameCtx) {
	// Put the current player in a known state: hand with a money card, a
	// property card, and filler so we exercise play money + play property +
	// hand-limit discard, then complete turn.
	cur := h.currentPlayer(g)

	h.writeState(g, func(b *stateBuilder) {
		b.setTurn(cur)
		// hand: 1 money (5M), 1 property (brown1), 6 filler money(1M) => 8 cards,
		// over the 7-card limit so a discard is required at turn end.
		b.setHand(cur,
			deal_no_mercy.AssetKeyMoney5,
			deal_no_mercy.AssetKeyBrown1,
			deal_no_mercy.AssetKeyMoney1,
			deal_no_mercy.AssetKeyMoney1,
			deal_no_mercy.AssetKeyMoney1,
			deal_no_mercy.AssetKeyMoney1,
			deal_no_mercy.AssetKeyMoney1,
			deal_no_mercy.AssetKeyMoney1,
		)
		b.setBank(cur)
		b.clearProperties(cur)
	})
	state, _ := g.resync(h)
	hand := state.GetYourHand().GetCards()

	var moneyCard, propCard string
	for _, c := range hand {
		if c.GetAssetKey() == dnm.AssetKey_ASSET_KEY_MONEY_5 && moneyCard == "" {
			moneyCard = c.GetCardId()
		}
		if c.GetCategory() == dnm.Category_CATEGORY_PURE_PROPERTY && propCard == "" {
			propCard = c.GetCardId()
		}
	}

	// Play money.
	must(g.send(cur, &dnm.ClientMessage{Payload: &dnm.ClientMessage_PlayMoney{
		PlayMoney: &dnm.PlayMoney{CardId: moneyCard},
	}}), "send play money")
	act, err := awaitAction(g.sockFor(cur), dnm.ActionKind_ACTION_KIND_PLAY_MONEY)
	if err != nil {
		h.record("play money", false, err.Error())
	} else {
		h.record("play money", act.GetPlayerId() == cur.String(), "bank +5M accepted")
	}
	// Opponent also sees the play (money plays are public).
	_, oerr := awaitAction(g.otherSock(cur), dnm.ActionKind_ACTION_KIND_PLAY_MONEY)
	h.record("opponent sees play money", oerr == nil, boolStr(oerr == nil, "received", errStr(oerr)))

	// Play property.
	must(g.send(cur, &dnm.ClientMessage{Payload: &dnm.ClientMessage_PlayProperty{
		PlayProperty: &dnm.PlayProperty{CardId: propCard},
	}}), "send play property")
	pact, err := awaitAction(g.sockFor(cur), dnm.ActionKind_ACTION_KIND_PLAY_PROPERTY)
	h.record("play property", err == nil && pact != nil, boolStr(err == nil, "brown set created", errStr(err)))

	// Third play: bank a 1M so moves are exhausted (3 plays used).
	var thirdMoney string
	{
		game := h.reload(g.gameID)
		for _, c := range game.Hands[cur] {
			if c.AssetKey == deal_no_mercy.AssetKeyMoney1 {
				thirdMoney = string(c.ID)
				break
			}
		}
	}
	must(g.send(cur, &dnm.ClientMessage{Payload: &dnm.ClientMessage_PlayMoney{
		PlayMoney: &dnm.PlayMoney{CardId: thirdMoney},
	}}), "send third play")
	_, _ = awaitAction(g.sockFor(cur), dnm.ActionKind_ACTION_KIND_PLAY_MONEY)
	g.otherSock(cur).drain()

	// Now hand should be 5 cards; limit is 7, so no discard needed... we need
	// >7 to force a discard. Re-check: started 8, played 3 => 5. Adjust plan:
	// instead over-fill the hand directly so discard is required.
	h.writeState(g, func(b *stateBuilder) {
		// leave current player's turn intact but overfill the hand to 9.
		b.g.MovesLeft = 0
		b.setHand(cur,
			deal_no_mercy.AssetKeyMoney1, deal_no_mercy.AssetKeyMoney1, deal_no_mercy.AssetKeyMoney1,
			deal_no_mercy.AssetKeyMoney1, deal_no_mercy.AssetKeyMoney1, deal_no_mercy.AssetKeyMoney1,
			deal_no_mercy.AssetKeyMoney1, deal_no_mercy.AssetKeyMoney1, deal_no_mercy.AssetKeyMoney1,
		)
	})
	state, _ = g.resync(h)
	hand = state.GetYourHand().GetCards()
	// Complete turn should be rejected while over the hand limit.
	must(g.send(cur, &dnm.ClientMessage{Payload: &dnm.ClientMessage_CompleteTurn{CompleteTurn: &dnm.CompleteTurn{}}}), "send complete over-limit")
	eerr, _ := awaitError(g.sockFor(cur))
	h.record("complete-turn rejected over hand limit", eerr != nil, boolStr(eerr != nil, msgOf(eerr), "no error returned"))

	// Discard down to 7 (drop 2 cards).
	discardIDs := []string{hand[0].GetCardId(), hand[1].GetCardId()}
	must(g.send(cur, &dnm.ClientMessage{Payload: &dnm.ClientMessage_DiscardCards{
		DiscardCards: &dnm.DiscardCards{CardIds: discardIDs},
	}}), "send discard")
	dact, err := awaitAction(g.sockFor(cur), dnm.ActionKind_ACTION_KIND_DISCARD_CARDS)
	h.record("hand-limit discard", err == nil && dact != nil, boolStr(err == nil, "discarded 2", errStr(err)))

	// Opponent sees masked discard (count only, no card ids revealed).
	m, oerr := g.otherSock(cur).nextMatching(waitT, func(sm *schema.ServerMessage) bool {
		a := dnmAction(sm)
		return a != nil && a.GetKind() == dnm.ActionKind_ACTION_KIND_DISCARD_CARDS
	})
	maskedDiscardOK := false
	var mdDetail string
	if oerr == nil {
		a := dnmAction(m)
		md := a.GetMaskedActionDiscardCards()
		real := a.GetActionDiscardCards()
		maskedDiscardOK = md != nil && real == nil
		if md != nil {
			mdDetail = fmt.Sprintf("masked count=%d (no ids)", md.GetNumCards())
		} else {
			mdDetail = "opponent got UNMASKED discard (leak)"
		}
	} else {
		mdDetail = errStr(oerr)
	}
	h.record("opponent sees masked discard (count only)", maskedDiscardOK, mdDetail)

	// Complete turn now (hand at 7). Opponent should see a masked start-turn.
	must(g.send(cur, &dnm.ClientMessage{Payload: &dnm.ClientMessage_CompleteTurn{CompleteTurn: &dnm.CompleteTurn{}}}), "send complete turn")
	// The actor gets a real start-turn for the NEXT player only if it's them;
	// here turn passes to opponent, so opponent gets the real start-turn.
	stAct, err := awaitAction(g.otherSock(cur), dnm.ActionKind_ACTION_KIND_START_TURN)
	turnPassed := err == nil && stAct != nil
	h.record("complete turn advances to opponent", turnPassed, boolStr(err == nil, "opponent got start-turn", errStr(err)))

	// The player who just ended their turn should see a MASKED start-turn for
	// the opponent (they can't see the drawn cards).
	m2, oerr2 := g.sockFor(cur).nextMatching(waitT, func(sm *schema.ServerMessage) bool {
		a := dnmAction(sm)
		return a != nil && a.GetKind() == dnm.ActionKind_ACTION_KIND_START_TURN
	})
	maskedStartOK := false
	var msDetail string
	if oerr2 == nil {
		a := dnmAction(m2)
		maskedStartOK = a.GetMaskedActionStartTurn() != nil && a.GetActionStartTurn() == nil
		msDetail = boolStr(maskedStartOK, "masked (draw count only)", "UNMASKED start-turn leaked to opponent")
	} else {
		msDetail = errStr(oerr2)
	}
	h.record("bystander sees masked start-turn", maskedStartOK, msDetail)
}

// other returns the non-cur player id.
func (g *GameCtx) other(pid uuid.UUID) uuid.UUID {
	if pid == g.a.playerID {
		return g.b.playerID
	}
	return g.a.playerID
}

func boolStr(ok bool, y, n string) string {
	if ok {
		return y
	}
	return n
}

func errStr(err error) string {
	if err == nil {
		return "ok"
	}
	return err.Error()
}

func msgOf(e *dnm.Error) string {
	if e == nil {
		return ""
	}
	return e.GetMessage()
}
