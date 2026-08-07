package deal_no_mercy

import (
	"the-deal/internal/schema"
	"the-deal/internal/schema/deal_no_mercy_schema"
	"the-deal/internal/token"
)

func (c *Controller) MaskEvent(tp token.Payload, msg *deal_no_mercy_schema.ServerMessage) *schema.ServerMessage {
	var serverMsg *schema.ServerMessage

	switch p := msg.GetPayload().(type) {
	case *deal_no_mercy_schema.ServerMessage_Action:
		serverMsg = &schema.ServerMessage{
			Payload: &schema.ServerMessage_DealNoMercyMessage{
				DealNoMercyMessage: &deal_no_mercy_schema.ServerMessage{
					Payload: &deal_no_mercy_schema.ServerMessage_Action{
						Action: c.MaskAction(tp, p.Action),
					},
				},
			},
		}

	case *deal_no_mercy_schema.ServerMessage_ActionHistory:
		serverMsg = &schema.ServerMessage{
			Payload: &schema.ServerMessage_DealNoMercyMessage{
				DealNoMercyMessage: &deal_no_mercy_schema.ServerMessage{
					Payload: &deal_no_mercy_schema.ServerMessage_ActionHistory{
						ActionHistory: &deal_no_mercy_schema.ActionHistory{
							Action: c.MaskAction(tp, p.ActionHistory.Action),
						},
					},
				},
			},
		}

	default:
		serverMsg = &schema.ServerMessage{
			Payload: &schema.ServerMessage_DealNoMercyMessage{
				DealNoMercyMessage: msg,
			},
		}
	}

	return serverMsg
}

// MaskAction hides the hidden-information portion of an action from viewers
// who are not entitled to it:
//   - ActionStartTurn / ActionPlayBigPayday: the drawn cards are visible only
//     to the drawing player; others see the count.
//   - ActionDiscardCards: the discarded cards are visible only to the
//     discarding player; others see the count.
//   - ActionDemandComplied carrying a pickpocket transfer: the stolen hand
//     cards are visible only to the two participants (the victim, who is the
//     action's PlayerId, and the attacker, who is the transfer's TargetId);
//     everyone else sees only the count and named category.
func (c *Controller) MaskAction(tp token.Payload, msg *deal_no_mercy_schema.Action) *deal_no_mercy_schema.Action {
	if msg == nil {
		return nil
	}

	viewer := tp.PlayerID.String()

	switch p := msg.Payload.(type) {
	case *deal_no_mercy_schema.Action_ActionStartTurn:
		if msg.PlayerId != viewer {
			return &deal_no_mercy_schema.Action{
				PlayerId:       msg.PlayerId,
				Kind:           msg.Kind,
				SeqNum:         msg.SeqNum,
				TurnDeadlineMs: msg.TurnDeadlineMs,
				Payload: &deal_no_mercy_schema.Action_MaskedActionStartTurn{
					MaskedActionStartTurn: &deal_no_mercy_schema.MaskedActionStartTurn{
						NumCards:  int32(len(p.ActionStartTurn.GetCards())),
						MovesLeft: p.ActionStartTurn.GetMovesLeft(),
						Debts:     p.ActionStartTurn.GetDebts(),
						GoAgain:   p.ActionStartTurn.GetGoAgain(),
					},
				},
			}
		}

	case *deal_no_mercy_schema.Action_ActionPlayBigPayday:
		if msg.PlayerId != viewer {
			return &deal_no_mercy_schema.Action{
				PlayerId:       msg.PlayerId,
				Kind:           msg.Kind,
				SeqNum:         msg.SeqNum,
				TurnDeadlineMs: msg.TurnDeadlineMs,
				Payload: &deal_no_mercy_schema.Action_MaskedActionPlayBigPayday{
					MaskedActionPlayBigPayday: &deal_no_mercy_schema.MaskedActionPlayBigPayday{
						LastPlayedCard: p.ActionPlayBigPayday.GetLastPlayedCard(),
						NumCards:       int32(len(p.ActionPlayBigPayday.GetCards())),
					},
				},
			}
		}

	case *deal_no_mercy_schema.Action_ActionDiscardCards:
		if msg.PlayerId != viewer {
			return &deal_no_mercy_schema.Action{
				PlayerId:       msg.PlayerId,
				Kind:           msg.Kind,
				SeqNum:         msg.SeqNum,
				TurnDeadlineMs: msg.TurnDeadlineMs,
				Payload: &deal_no_mercy_schema.Action_MaskedActionDiscardCards{
					MaskedActionDiscardCards: &deal_no_mercy_schema.MaskedActionDiscardCards{
						NumCards: int32(len(p.ActionDiscardCards.GetCards())),
					},
				},
			}
		}

	case *deal_no_mercy_schema.Action_ActionDemandComplied:
		return maskDemandComplied(viewer, msg, p.ActionDemandComplied)
	}

	return msg
}

// maskDemandComplied rebuilds a demand-complied action, swapping a
// pickpocket transfer for its masked variant when the viewer is neither of
// the two participants. All other transfer kinds are public and pass through
// unchanged.
func maskDemandComplied(viewer string, msg *deal_no_mercy_schema.Action, complied *deal_no_mercy_schema.ActionDemandComplied) *deal_no_mercy_schema.Action {
	pickpocket, ok := complied.GetTransfer().(*deal_no_mercy_schema.ActionDemandComplied_TransferPickpocket)
	if !ok {
		return msg
	}

	tp := pickpocket.TransferPickpocket
	// The victim is the action's PlayerId (the complier / transfer source);
	// the attacker is the transfer's TargetId. Both see the real cards.
	if viewer == msg.PlayerId || viewer == tp.GetTargetId() {
		return msg
	}

	masked := &deal_no_mercy_schema.ActionDemandComplied{
		DemandId: complied.GetDemandId(),
		Transfer: &deal_no_mercy_schema.ActionDemandComplied_MaskedTransferPickpocket{
			MaskedTransferPickpocket: &deal_no_mercy_schema.MaskedTransferPickpocket{
				SourceId: tp.GetSourceId(),
				TargetId: tp.GetTargetId(),
				Category: tp.GetCategory(),
				NumCards: int32(len(tp.GetCards())),
			},
		},
	}

	return &deal_no_mercy_schema.Action{
		PlayerId:       msg.PlayerId,
		Kind:           msg.Kind,
		SeqNum:         msg.SeqNum,
		TurnDeadlineMs: msg.TurnDeadlineMs,
		Payload: &deal_no_mercy_schema.Action_ActionDemandComplied{
			ActionDemandComplied: masked,
		},
	}
}
