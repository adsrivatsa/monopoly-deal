package monopoly_deal

import (
	"the-deal/internal/schema"
	"the-deal/internal/schema/monopoly_deal_schema"
	"the-deal/internal/token"
)

func (c *Controller) MaskEvents(tp token.Payload, msg *monopoly_deal_schema.ServerMessage) *schema.ServerMessage {
	actionWrap, ok := msg.GetPayload().(*monopoly_deal_schema.ServerMessage_Action)
	if !ok {
		return nil
	}

	action := actionWrap.Action

	switch p := action.Payload.(type) {
	case *monopoly_deal_schema.Action_ActionStartTurn:
		if action.PlayerId != tp.PlayerID.String() {
			msg = &monopoly_deal_schema.ServerMessage{
				Payload: &monopoly_deal_schema.ServerMessage_Action{
					Action: &monopoly_deal_schema.Action{
						PlayerId:       action.PlayerId,
						Kind:           action.Kind,
						SeqNum:         action.SeqNum,
						TurnDeadlineMs: action.TurnDeadlineMs,
						Payload: &monopoly_deal_schema.Action_MaskedActionStartTurn{
							MaskedActionStartTurn: &monopoly_deal_schema.MaskedActionStartTurn{
								NumCards:  int32(len(p.ActionStartTurn.GetCards())),
								MovesLeft: p.ActionStartTurn.GetMovesLeft(),
							},
						},
					},
				},
			}
		}

	case *monopoly_deal_schema.Action_ActionPlayPassGo:
		if action.PlayerId != tp.PlayerID.String() {
			msg = &monopoly_deal_schema.ServerMessage{
				Payload: &monopoly_deal_schema.ServerMessage_Action{
					Action: &monopoly_deal_schema.Action{
						PlayerId:       action.PlayerId,
						Kind:           action.Kind,
						SeqNum:         action.SeqNum,
						TurnDeadlineMs: action.TurnDeadlineMs,
						Payload: &monopoly_deal_schema.Action_MaskedActionPlayedPassGo{
							MaskedActionPlayedPassGo: &monopoly_deal_schema.MaskedActionPlayPassGo{
								LastPlayedCard: p.ActionPlayPassGo.GetLastPlayedCard(),
								NumCards:       int32(len(p.ActionPlayPassGo.GetCards())),
							},
						},
					},
				},
			}
		}

	//case *monopoly_deal_schema.ServerMessage_PendingRentCreated:
	//	if p.PendingRentCreated.PendingRent.PlayerId != tp.PlayerID.String() {
	//		return nil
	//	}
	//
	//case *monopoly_deal_schema.ServerMessage_PendingRentResolved:
	//	if p.PendingRentResolved.PlayerId != tp.PlayerID.String() {
	//		return nil
	//	}
	//

	case *monopoly_deal_schema.Action_ActionDiscardCards:
		if action.PlayerId != tp.PlayerID.String() {
			msg = &monopoly_deal_schema.ServerMessage{
				Payload: &monopoly_deal_schema.ServerMessage_Action{
					Action: &monopoly_deal_schema.Action{
						PlayerId:       action.PlayerId,
						Kind:           action.Kind,
						SeqNum:         action.SeqNum,
						TurnDeadlineMs: action.TurnDeadlineMs,
						Payload: &monopoly_deal_schema.Action_MaskedActionDiscardCards{
							MaskedActionDiscardCards: &monopoly_deal_schema.MaskedActionDiscardCards{
								NumCards: int32(len(p.ActionDiscardCards.Cards)),
							},
						},
					},
				},
			}
		}

	}

	return &schema.ServerMessage{
		Payload: &schema.ServerMessage_MonopolyDealMessage{
			MonopolyDealMessage: msg,
		},
	}
}
