package main

import (
	"the-deal/internal/schema"
	dnm "the-deal/internal/schema/deal_no_mercy_schema"
)

// gw boxes a deal_no_mercy client message into the top-level gateway
// ClientMessage the socket expects.
func gw(m *dnm.ClientMessage) *schema.ClientMessage {
	return &schema.ClientMessage{
		Payload: &schema.ClientMessage_DealNoMercyMessage{
			DealNoMercyMessage: m,
		},
	}
}

// dnmState extracts the deal_no_mercy GameState from a server message, or nil.
func dnmState(m *schema.ServerMessage) *dnm.GameState {
	d := m.GetDealNoMercyMessage()
	if d == nil {
		return nil
	}
	return d.GetGameState()
}

// dnmAction extracts a deal_no_mercy Action from a server message, or nil.
func dnmAction(m *schema.ServerMessage) *dnm.Action {
	d := m.GetDealNoMercyMessage()
	if d == nil {
		return nil
	}
	return d.GetAction()
}

func dnmWon(m *schema.ServerMessage) *dnm.WonGame {
	d := m.GetDealNoMercyMessage()
	if d == nil {
		return nil
	}
	return d.GetWonGame()
}

func dnmErr(m *schema.ServerMessage) *dnm.Error {
	d := m.GetDealNoMercyMessage()
	if d == nil {
		return nil
	}
	return d.GetError()
}
