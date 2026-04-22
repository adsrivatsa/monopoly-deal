package monopoly_deal

import (
	"the-deal/internal/schema"
	"the-deal/internal/schema/monopoly_deal_schema"
	"the-deal/internal/token"
)

func (c *Controller) MaskEvents(tp token.Payload, msg *monopoly_deal_schema.ServerMessage) *schema.ServerMessage {
	switch p := msg.Payload.(type) {
	case *monopoly_deal_schema.ServerMessage_StartTurnRes:
		if p.StartTurnRes.PlayerId != tp.PlayerID.String() {
			msg = &monopoly_deal_schema.ServerMessage{
				Payload: &monopoly_deal_schema.ServerMessage_StartTurnMaskedRes{
					StartTurnMaskedRes: &monopoly_deal_schema.StartTurnMaskedRes{
						SeqNum:   msg.GetStartTurnRes().GetSeqNum(),
						PlayerId: msg.GetStartTurnRes().GetPlayerId(),
						NumCards: int32(len(msg.GetStartTurnRes().GetCards())),
					},
				},
			}
		}
	case *monopoly_deal_schema.ServerMessage_PlayPassGoRes:
		if p.PlayPassGoRes.PlayerId != tp.PlayerID.String() {
			msg = &monopoly_deal_schema.ServerMessage{
				Payload: &monopoly_deal_schema.ServerMessage_PlayPassGoMaskedRes{
					PlayPassGoMaskedRes: &monopoly_deal_schema.PlayPassGoMaskedRes{
						SeqNum:         msg.GetPlayPassGoRes().GetSeqNum(),
						PlayerId:       msg.GetPlayPassGoRes().GetPlayerId(),
						NumCards:       int32(len(msg.GetPlayPassGoRes().GetCards())),
						LastPlayedCard: msg.GetPlayPassGoRes().GetLastPlayedCard(),
					},
				},
			}
		}

	case *monopoly_deal_schema.ServerMessage_DemandCreated:
		if p.DemandCreated.Demand.PlayerId != tp.PlayerID.String() {
			return nil
		}

	case *monopoly_deal_schema.ServerMessage_PendingRentCreated:
		if p.PendingRentCreated.PendingRent.PlayerId != tp.PlayerID.String() {
			return nil
		}

	case *monopoly_deal_schema.ServerMessage_PendingRentResolved:
		if p.PendingRentResolved.PlayerId != tp.PlayerID.String() {
			return nil
		}

	case *monopoly_deal_schema.ServerMessage_DiscardCardsRes:
		if p.DiscardCardsRes.PlayerId != tp.PlayerID.String() {
			msg = &monopoly_deal_schema.ServerMessage{
				Payload: &monopoly_deal_schema.ServerMessage_DiscardCardsMaskedRes{
					DiscardCardsMaskedRes: &monopoly_deal_schema.DiscardCardsMaskedRes{
						SeqNum:   msg.GetDiscardCardsRes().GetSeqNum(),
						PlayerId: msg.GetDiscardCardsRes().GetPlayerId(),
						NumCards: int32(len(msg.GetDiscardCardsRes().GetCards())),
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
