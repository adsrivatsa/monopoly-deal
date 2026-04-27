package monopoly_deal

import (
	"the-deal/internal/schema"
	"the-deal/internal/schema/monopoly_deal_schema"
	"the-deal/internal/token"
)

func (c *Controller) MaskEvent(tp token.Payload, msg *monopoly_deal_schema.ServerMessage) *schema.ServerMessage {
	var serverMsg *schema.ServerMessage

	switch p := msg.GetPayload().(type) {
	case *monopoly_deal_schema.ServerMessage_Action:
		serverMsg = &schema.ServerMessage{
			Payload: &schema.ServerMessage_MonopolyDealMessage{
				MonopolyDealMessage: &monopoly_deal_schema.ServerMessage{
					Payload: &monopoly_deal_schema.ServerMessage_Action{
						Action: c.MaskAction(tp, p.Action),
					},
				},
			},
		}

	case *monopoly_deal_schema.ServerMessage_ActionHistory:
		serverMsg = &schema.ServerMessage{
			Payload: &schema.ServerMessage_MonopolyDealMessage{
				MonopolyDealMessage: &monopoly_deal_schema.ServerMessage{
					Payload: &monopoly_deal_schema.ServerMessage_ActionHistory{
						ActionHistory: &monopoly_deal_schema.ActionHistory{
							Action: c.MaskAction(tp, p.ActionHistory.Action),
						},
					},
				},
			},
		}

	default:
		serverMsg = &schema.ServerMessage{
			Payload: &schema.ServerMessage_MonopolyDealMessage{
				MonopolyDealMessage: msg,
			},
		}
	}

	return serverMsg
}

func (c *Controller) MaskAction(tp token.Payload, msg *monopoly_deal_schema.Action) *monopoly_deal_schema.Action {
	switch p := msg.Payload.(type) {
	case *monopoly_deal_schema.Action_ActionStartTurn:
		if msg.PlayerId != tp.PlayerID.String() {
			msg = &monopoly_deal_schema.Action{
				PlayerId:       msg.PlayerId,
				Kind:           msg.Kind,
				SeqNum:         msg.SeqNum,
				TurnDeadlineMs: msg.TurnDeadlineMs,
				Payload: &monopoly_deal_schema.Action_MaskedActionStartTurn{
					MaskedActionStartTurn: &monopoly_deal_schema.MaskedActionStartTurn{
						NumCards:  int32(len(p.ActionStartTurn.GetCards())),
						MovesLeft: p.ActionStartTurn.GetMovesLeft(),
					},
				},
			}
		}

	case *monopoly_deal_schema.Action_ActionPlayPassGo:
		if msg.PlayerId != tp.PlayerID.String() {
			msg = &monopoly_deal_schema.Action{
				PlayerId:       msg.PlayerId,
				Kind:           msg.Kind,
				SeqNum:         msg.SeqNum,
				TurnDeadlineMs: msg.TurnDeadlineMs,
				Payload: &monopoly_deal_schema.Action_MaskedActionPlayedPassGo{
					MaskedActionPlayedPassGo: &monopoly_deal_schema.MaskedActionPlayPassGo{
						LastPlayedCard: p.ActionPlayPassGo.GetLastPlayedCard(),
						NumCards:       int32(len(p.ActionPlayPassGo.GetCards())),
					},
				},
			}
		}

	case *monopoly_deal_schema.Action_ActionDiscardCards:
		if msg.PlayerId != tp.PlayerID.String() {
			msg = &monopoly_deal_schema.Action{
				PlayerId:       msg.PlayerId,
				Kind:           msg.Kind,
				SeqNum:         msg.SeqNum,
				TurnDeadlineMs: msg.TurnDeadlineMs,
				Payload: &monopoly_deal_schema.Action_MaskedActionDiscardCards{
					MaskedActionDiscardCards: &monopoly_deal_schema.MaskedActionDiscardCards{
						NumCards: int32(len(p.ActionDiscardCards.Cards)),
					},
				},
			}
		}
	}

	return msg
}
