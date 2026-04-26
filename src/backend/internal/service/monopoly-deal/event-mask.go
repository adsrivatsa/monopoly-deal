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
	//case *monopoly_deal_schema.ServerMessage_StartTurnRes:
	//	if p.StartTurnRes.PlayerId != tp.PlayerID.String() {
	//		msg = &monopoly_deal_schema.ServerMessage{
	//			Payload: &monopoly_deal_schema.ServerMessage_StartTurnMaskedRes{
	//				StartTurnMaskedRes: &monopoly_deal_schema.StartTurnMaskedRes{
	//					SeqNum:   msg.GetStartTurnRes().GetSeqNum(),
	//					PlayerId: msg.GetStartTurnRes().GetPlayerId(),
	//					NumCards: int32(len(msg.GetStartTurnRes().GetCards())),
	//				},
	//			},
	//		}
	//	}

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
		//case *monopoly_deal_schema.ServerMessage_DiscardCardsRes:
		//	if p.DiscardCardsRes.PlayerId != tp.PlayerID.String() {
		//		msg = &monopoly_deal_schema.ServerMessage{
		//			Payload: &monopoly_deal_schema.ServerMessage_DiscardCardsMaskedRes{
		//				DiscardCardsMaskedRes: &monopoly_deal_schema.DiscardCardsMaskedRes{
		//					SeqNum:   msg.GetDiscardCardsRes().GetSeqNum(),
		//					PlayerId: msg.GetDiscardCardsRes().GetPlayerId(),
		//					NumCards: int32(len(msg.GetDiscardCardsRes().GetCards())),
		//				},
		//			},
		//		}
		//	}

	}

	return &schema.ServerMessage{
		Payload: &schema.ServerMessage_MonopolyDealMessage{
			MonopolyDealMessage: msg,
		},
	}
}
