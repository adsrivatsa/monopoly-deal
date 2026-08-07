package deal_no_mercy

import (
	"the-deal/internal/schema/deal_no_mercy_schema"
)

// Each Action type produces its full, unmasked *deal_no_mercy_schema.Action.
// TurnDeadlineMs is left 0 and stamped by the service-layer caller (mirrors
// the classic engine). Masking of hidden information (drawn cards, stolen
// hand cards) is the Phase 3 service layer's job; the masked proto variants
// this file's siblings define are what the service selects between.

func (a *ActionPlayMoney) Proto() *deal_no_mercy_schema.Action {
	return &deal_no_mercy_schema.Action{
		PlayerId:       a.PlayerID.String(),
		Kind:           a.Kind.Proto(),
		SeqNum:         int32(a.SeqNum),
		TurnDeadlineMs: 0,
		Payload: &deal_no_mercy_schema.Action_ActionPlayMoney{
			ActionPlayMoney: &deal_no_mercy_schema.ActionPlayMoney{
				Card: a.Card.Proto(),
			},
		},
	}
}

func (a *ActionPlayProperty) Proto() *deal_no_mercy_schema.Action {
	return &deal_no_mercy_schema.Action{
		PlayerId:       a.PlayerID.String(),
		Kind:           a.Kind.Proto(),
		SeqNum:         int32(a.SeqNum),
		TurnDeadlineMs: 0,
		Payload: &deal_no_mercy_schema.Action_ActionPlayProperty{
			ActionPlayProperty: &deal_no_mercy_schema.ActionPlayProperty{
				PropertySet: a.PropertySet.Proto(a.PlayerID),
			},
		},
	}
}

func (a *ActionPlayShack) Proto() *deal_no_mercy_schema.Action {
	return &deal_no_mercy_schema.Action{
		PlayerId:       a.PlayerID.String(),
		Kind:           a.Kind.Proto(),
		SeqNum:         int32(a.SeqNum),
		TurnDeadlineMs: 0,
		Payload: &deal_no_mercy_schema.Action_ActionPlayShack{
			ActionPlayShack: &deal_no_mercy_schema.ActionPlayShack{
				Card:        a.Card.Proto(),
				PropertySet: a.PropertySet.Proto(a.PlayerID),
			},
		},
	}
}

func (a *ActionPlayBigPayday) Proto() *deal_no_mercy_schema.Action {
	return &deal_no_mercy_schema.Action{
		PlayerId:       a.PlayerID.String(),
		Kind:           a.Kind.Proto(),
		SeqNum:         int32(a.SeqNum),
		TurnDeadlineMs: 0,
		Payload: &deal_no_mercy_schema.Action_ActionPlayBigPayday{
			ActionPlayBigPayday: &deal_no_mercy_schema.ActionPlayBigPayday{
				LastPlayedCard: a.LastPlayedCard.Proto(),
				Cards:          a.Cards.Proto(),
			},
		},
	}
}

// MaskedProto hides the drawn cards for BIG PAYDAY (only the count is public).
func (a *ActionPlayBigPayday) MaskedProto() *deal_no_mercy_schema.Action {
	return &deal_no_mercy_schema.Action{
		PlayerId:       a.PlayerID.String(),
		Kind:           a.Kind.Proto(),
		SeqNum:         int32(a.SeqNum),
		TurnDeadlineMs: 0,
		Payload: &deal_no_mercy_schema.Action_MaskedActionPlayBigPayday{
			MaskedActionPlayBigPayday: &deal_no_mercy_schema.MaskedActionPlayBigPayday{
				LastPlayedCard: a.LastPlayedCard.Proto(),
				NumCards:       int32(a.Cards.Len()),
			},
		},
	}
}

func (a *ActionPlayGoAgain) Proto() *deal_no_mercy_schema.Action {
	return &deal_no_mercy_schema.Action{
		PlayerId:       a.PlayerID.String(),
		Kind:           a.Kind.Proto(),
		SeqNum:         int32(a.SeqNum),
		TurnDeadlineMs: 0,
		Payload: &deal_no_mercy_schema.Action_ActionPlayGoAgain{
			ActionPlayGoAgain: &deal_no_mercy_schema.ActionPlayGoAgain{
				LastPlayedCard: a.LastPlayedCard.Proto(),
			},
		},
	}
}

func (a *ActionDemandsCreated) Proto() *deal_no_mercy_schema.Action {
	var lastPlayedCard *deal_no_mercy_schema.Card
	if a.LastPlayedCard != nil {
		lastPlayedCard = a.LastPlayedCard.Proto()
	}

	var deniedDemand *deal_no_mercy_schema.Demand
	if a.DeniedDemand != nil {
		deniedDemand = a.DeniedDemand.Proto()
	}

	demands := make([]*deal_no_mercy_schema.Demand, len(a.Demands))
	for i := range a.Demands {
		demands[i] = a.Demands[i].Proto()
	}

	return &deal_no_mercy_schema.Action{
		PlayerId:       a.PlayerID.String(),
		Kind:           a.Kind.Proto(),
		SeqNum:         int32(a.SeqNum),
		TurnDeadlineMs: 0,
		Payload: &deal_no_mercy_schema.Action_ActionDemandsCreated{
			ActionDemandsCreated: &deal_no_mercy_schema.ActionDemandsCreated{
				LastPlayedCard: lastPlayedCard,
				DeniedDemand:   deniedDemand,
				Demands:        demands,
			},
		},
	}
}

// ---- Transfer sub-messages ----

func (t *TransferCards) Proto() *deal_no_mercy_schema.TransferCards {
	var debt *deal_no_mercy_schema.DebtChip
	if t.Debt != nil {
		debt = t.Debt.Proto()
	}

	return &deal_no_mercy_schema.TransferCards{
		SourceId:     t.SourceID.String(),
		TargetId:     t.TargetID.String(),
		Cards:        t.Cards.Proto(),
		PropertySets: t.PropertySets.Proto(t.TargetID),
		SourceSets:   int32(t.SourceSets),
		SourceMoney:  int32(t.SourceMoney),
		TargetSets:   int32(t.TargetSets),
		TargetMoney:  int32(t.TargetMoney),
		Debt:         debt,
	}
}

func (t *TransferProperty) Proto() *deal_no_mercy_schema.TransferProperty {
	return &deal_no_mercy_schema.TransferProperty{
		SourceId:    t.SourceID.String(),
		TargetId:    t.TargetID.String(),
		Card:        t.Card.Proto(),
		PropertySet: t.PropertySet.Proto(t.TargetID),
	}
}

func (t *TransferPropertySets) Proto() *deal_no_mercy_schema.TransferPropertySets {
	return &deal_no_mercy_schema.TransferPropertySets{
		SourceId:     t.SourceID.String(),
		TargetId:     t.TargetID.String(),
		PropertySets: t.PropertySets.Proto(t.TargetID),
	}
}

func (t *TransferBankSwap) Proto() *deal_no_mercy_schema.TransferBankSwap {
	return &deal_no_mercy_schema.TransferBankSwap{
		SourceId:   t.SourceID.String(),
		TargetId:   t.TargetID.String(),
		SourceBank: t.SourceBank.Proto(),
		TargetBank: t.TargetBank.Proto(),
	}
}

func (t *TransferDistribution) Proto() *deal_no_mercy_schema.TransferDistribution {
	entries := make([]*deal_no_mercy_schema.DistributionEntry, len(t.Entries))
	for i, e := range t.Entries {
		var setProto *deal_no_mercy_schema.PropertySet
		if e.PropertySet != nil {
			setProto = e.PropertySet.Proto(e.RecipientID)
		}
		entries[i] = &deal_no_mercy_schema.DistributionEntry{
			RecipientId: e.RecipientID.String(),
			Card:        e.Card.Proto(),
			PropertySet: setProto,
		}
	}

	return &deal_no_mercy_schema.TransferDistribution{
		SourceId: t.SourceID.String(),
		KeptCard: t.KeptCard.Proto(),
		Entries:  entries,
	}
}

func (t *TransferPickpocket) Proto() *deal_no_mercy_schema.TransferPickpocket {
	return &deal_no_mercy_schema.TransferPickpocket{
		SourceId: t.SourceID.String(),
		TargetId: t.TargetID.String(),
		Category: t.Category.Proto(),
		Cards:    t.Cards.Proto(),
	}
}

// MaskedProto hides the specific hand cards stolen by PICKPOCKET (only the
// count and named category are public to non-participants).
func (t *TransferPickpocket) MaskedProto() *deal_no_mercy_schema.MaskedTransferPickpocket {
	return &deal_no_mercy_schema.MaskedTransferPickpocket{
		SourceId: t.SourceID.String(),
		TargetId: t.TargetID.String(),
		Category: t.Category.Proto(),
		NumCards: int32(t.Cards.Len()),
	}
}

func (t *TransferDebtChip) Proto() *deal_no_mercy_schema.TransferDebtChip {
	return &deal_no_mercy_schema.TransferDebtChip{
		Debt: t.Debt.Proto(),
	}
}

func (a *ActionDemandComplied) Proto() *deal_no_mercy_schema.Action {
	complied := &deal_no_mercy_schema.ActionDemandComplied{
		DemandId: string(a.DemandID),
	}

	switch {
	case a.TransferCards != nil:
		complied.Transfer = &deal_no_mercy_schema.ActionDemandComplied_TransferCards{
			TransferCards: a.TransferCards.Proto(),
		}
	case a.TransferProperty != nil:
		complied.Transfer = &deal_no_mercy_schema.ActionDemandComplied_TransferProperty{
			TransferProperty: a.TransferProperty.Proto(),
		}
	case a.TransferPropertySets != nil:
		complied.Transfer = &deal_no_mercy_schema.ActionDemandComplied_TransferPropertySets{
			TransferPropertySets: a.TransferPropertySets.Proto(),
		}
	case a.TransferBankSwap != nil:
		complied.Transfer = &deal_no_mercy_schema.ActionDemandComplied_TransferBankSwap{
			TransferBankSwap: a.TransferBankSwap.Proto(),
		}
	case a.TransferDistribution != nil:
		complied.Transfer = &deal_no_mercy_schema.ActionDemandComplied_TransferDistribution{
			TransferDistribution: a.TransferDistribution.Proto(),
		}
	case a.TransferPickpocket != nil:
		complied.Transfer = &deal_no_mercy_schema.ActionDemandComplied_TransferPickpocket{
			TransferPickpocket: a.TransferPickpocket.Proto(),
		}
	case a.TransferDebtChip != nil:
		complied.Transfer = &deal_no_mercy_schema.ActionDemandComplied_TransferDebtChip{
			TransferDebtChip: a.TransferDebtChip.Proto(),
		}
	}

	return &deal_no_mercy_schema.Action{
		PlayerId:       a.PlayerID.String(),
		Kind:           a.Kind.Proto(),
		SeqNum:         int32(a.SeqNum),
		TurnDeadlineMs: 0,
		Payload: &deal_no_mercy_schema.Action_ActionDemandComplied{
			ActionDemandComplied: complied,
		},
	}
}

func (a *ActionDebtSettled) Proto() *deal_no_mercy_schema.Action {
	var setProto *deal_no_mercy_schema.PropertySet
	if a.PropertySet != nil {
		setProto = a.PropertySet.Proto(a.Debt.CreditorID)
	}

	return &deal_no_mercy_schema.Action{
		PlayerId:       a.PlayerID.String(),
		Kind:           a.Kind.Proto(),
		SeqNum:         int32(a.SeqNum),
		TurnDeadlineMs: 0,
		Payload: &deal_no_mercy_schema.Action_ActionDebtSettled{
			ActionDebtSettled: &deal_no_mercy_schema.ActionDebtSettled{
				Debt:        a.Debt.Proto(),
				Card:        a.Card.Proto(),
				PropertySet: setProto,
				Forgiven:    a.Forgiven,
			},
		},
	}
}

func (a *ActionDiscardCards) Proto() *deal_no_mercy_schema.Action {
	return &deal_no_mercy_schema.Action{
		PlayerId:       a.PlayerID.String(),
		Kind:           a.Kind.Proto(),
		SeqNum:         int32(a.SeqNum),
		TurnDeadlineMs: 0,
		Payload: &deal_no_mercy_schema.Action_ActionDiscardCards{
			ActionDiscardCards: &deal_no_mercy_schema.ActionDiscardCards{
				Cards: a.Cards.Proto(),
			},
		},
	}
}

// MaskedProto hides which cards another player discarded (only the count).
func (a *ActionDiscardCards) MaskedProto() *deal_no_mercy_schema.Action {
	return &deal_no_mercy_schema.Action{
		PlayerId:       a.PlayerID.String(),
		Kind:           a.Kind.Proto(),
		SeqNum:         int32(a.SeqNum),
		TurnDeadlineMs: 0,
		Payload: &deal_no_mercy_schema.Action_MaskedActionDiscardCards{
			MaskedActionDiscardCards: &deal_no_mercy_schema.MaskedActionDiscardCards{
				NumCards: int32(a.Cards.Len()),
			},
		},
	}
}

func (a *ActionStartTurn) Proto() *deal_no_mercy_schema.Action {
	return &deal_no_mercy_schema.Action{
		PlayerId:       a.PlayerID.String(),
		Kind:           a.Kind.Proto(),
		SeqNum:         int32(a.SeqNum),
		TurnDeadlineMs: 0,
		Payload: &deal_no_mercy_schema.Action_ActionStartTurn{
			ActionStartTurn: &deal_no_mercy_schema.ActionStartTurn{
				Cards:     a.Cards.Proto(),
				MovesLeft: int32(a.MovesLeft),
				Debts:     debtsProto(a.Debts),
				GoAgain:   a.GoAgain,
			},
		},
	}
}

// MaskedProto hides the drawn cards for another player's turn start; the
// outstanding debts and go-again flag stay public.
func (a *ActionStartTurn) MaskedProto() *deal_no_mercy_schema.Action {
	return &deal_no_mercy_schema.Action{
		PlayerId:       a.PlayerID.String(),
		Kind:           a.Kind.Proto(),
		SeqNum:         int32(a.SeqNum),
		TurnDeadlineMs: 0,
		Payload: &deal_no_mercy_schema.Action_MaskedActionStartTurn{
			MaskedActionStartTurn: &deal_no_mercy_schema.MaskedActionStartTurn{
				NumCards:  int32(a.Cards.Len()),
				MovesLeft: int32(a.MovesLeft),
				Debts:     debtsProto(a.Debts),
				GoAgain:   a.GoAgain,
			},
		},
	}
}

func (a *ActionRearrangeCard) Proto() *deal_no_mercy_schema.Action {
	return &deal_no_mercy_schema.Action{
		PlayerId:       a.PlayerID.String(),
		Kind:           a.Kind.Proto(),
		SeqNum:         int32(a.SeqNum),
		TurnDeadlineMs: 0,
		Payload: &deal_no_mercy_schema.Action_ActionRearrangeCard{
			ActionRearrangeCard: &deal_no_mercy_schema.ActionRearrangeCard{
				Card:        cardPtrProto(a.Card),
				PropertySet: a.PropertySet.Proto(a.PlayerID),
				Sets:        int32(a.Sets),
			},
		},
	}
}
