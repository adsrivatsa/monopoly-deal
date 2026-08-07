package deal_no_mercy

import (
	"testing"

	deal_no_mercy "the-deal/internal/engine/deal-no-mercy"
	"the-deal/internal/schema/deal_no_mercy_schema"
	"the-deal/internal/token"

	"github.com/google/uuid"
)

func tp(id uuid.UUID) token.Payload {
	return token.Payload{PlayerID: id}
}

func sampleCard(id string) deal_no_mercy.Card {
	return deal_no_mercy.NewCard(deal_no_mercy.Identifier(id), deal_no_mercy.CategoryMoney, deal_no_mercy.AssetKeyMoney1, 1)
}

func TestMaskActionStartTurn(t *testing.T) {
	c := &Controller{}
	owner := uuid.New()
	other := uuid.New()

	engineAction := deal_no_mercy.NewActionStartTurn(0, owner, deal_no_mercy.Cards{sampleCard("001"), sampleCard("002")}, 3, nil, false)
	action := engineAction.Proto()

	// Owner sees the real drawn cards.
	ownerView := c.MaskAction(tp(owner), action)
	if ownerView.GetActionStartTurn() == nil {
		t.Fatalf("owner should see unmasked ActionStartTurn, got %T", ownerView.Payload)
	}
	if len(ownerView.GetActionStartTurn().GetCards()) != 2 {
		t.Fatalf("owner should see 2 cards, got %d", len(ownerView.GetActionStartTurn().GetCards()))
	}

	// A different viewer only sees the count.
	otherView := c.MaskAction(tp(other), action)
	masked := otherView.GetMaskedActionStartTurn()
	if masked == nil {
		t.Fatalf("other should see MaskedActionStartTurn, got %T", otherView.Payload)
	}
	if masked.GetNumCards() != 2 {
		t.Fatalf("masked NumCards = %d, want 2", masked.GetNumCards())
	}
	if masked.GetMovesLeft() != 3 {
		t.Fatalf("masked MovesLeft = %d, want 3", masked.GetMovesLeft())
	}
}

func TestMaskActionPlayBigPayday(t *testing.T) {
	c := &Controller{}
	owner := uuid.New()
	other := uuid.New()

	engineAction := deal_no_mercy.NewActionPlayBigPayday(0, owner, sampleCard("bp0"),
		deal_no_mercy.Cards{sampleCard("001"), sampleCard("002"), sampleCard("003")})
	action := engineAction.Proto()

	if c.MaskAction(tp(owner), action).GetActionPlayBigPayday() == nil {
		t.Fatalf("owner should see unmasked big payday")
	}

	masked := c.MaskAction(tp(other), action).GetMaskedActionPlayBigPayday()
	if masked == nil {
		t.Fatalf("other should see masked big payday")
	}
	if masked.GetNumCards() != 3 {
		t.Fatalf("masked NumCards = %d, want 3", masked.GetNumCards())
	}
}

func TestMaskActionDiscardCards(t *testing.T) {
	c := &Controller{}
	owner := uuid.New()
	other := uuid.New()

	engineAction := deal_no_mercy.NewActionDiscardCards(0, owner, sampleCard("001"), sampleCard("002"))
	action := engineAction.Proto()

	if c.MaskAction(tp(owner), action).GetActionDiscardCards() == nil {
		t.Fatalf("owner should see unmasked discard")
	}

	masked := c.MaskAction(tp(other), action).GetMaskedActionDiscardCards()
	if masked == nil {
		t.Fatalf("other should see masked discard")
	}
	if masked.GetNumCards() != 2 {
		t.Fatalf("masked NumCards = %d, want 2", masked.GetNumCards())
	}
}

// buildPickpocketComplied constructs the demand-complied proto action for a
// pickpocket. Per the engine, the complier (action.PlayerId) is the victim
// (TransferPickpocket.SourceID) and the attacker is TransferPickpocket.TargetID.
func buildPickpocketComplied(victim, attacker uuid.UUID) *deal_no_mercy_schema.Action {
	engineAction := deal_no_mercy.NewActionDemandComplied(0, victim, deal_no_mercy.Identifier("d1"))
	engineAction.TransferPickpocket = &deal_no_mercy.TransferPickpocket{
		SourceID: victim,
		TargetID: attacker,
		Category: deal_no_mercy.StealCategoryAction,
		Cards:    deal_no_mercy.Cards{sampleCard("001"), sampleCard("002")},
	}
	return engineAction.Proto()
}

func TestMaskPickpocketVictimAndAttackerSeeReal(t *testing.T) {
	c := &Controller{}
	victim := uuid.New()
	attacker := uuid.New()

	action := buildPickpocketComplied(victim, attacker)

	// Victim (the complier / action PlayerId) sees the real stolen cards.
	victimView := c.MaskAction(tp(victim), action).GetActionDemandComplied()
	if victimView == nil {
		t.Fatalf("victim should see ActionDemandComplied")
	}
	realTransfer, ok := victimView.GetTransfer().(*deal_no_mercy_schema.ActionDemandComplied_TransferPickpocket)
	if !ok {
		t.Fatalf("victim should see real TransferPickpocket, got %T", victimView.GetTransfer())
	}
	if len(realTransfer.TransferPickpocket.GetCards()) != 2 {
		t.Fatalf("victim should see 2 cards, got %d", len(realTransfer.TransferPickpocket.GetCards()))
	}

	// Attacker (TransferPickpocket.TargetId) also sees the real cards.
	attackerView := c.MaskAction(tp(attacker), action).GetActionDemandComplied()
	if _, ok := attackerView.GetTransfer().(*deal_no_mercy_schema.ActionDemandComplied_TransferPickpocket); !ok {
		t.Fatalf("attacker should see real TransferPickpocket, got %T", attackerView.GetTransfer())
	}
}

func TestMaskPickpocketBystanderMasked(t *testing.T) {
	c := &Controller{}
	victim := uuid.New()
	attacker := uuid.New()
	bystander := uuid.New()

	action := buildPickpocketComplied(victim, attacker)

	view := c.MaskAction(tp(bystander), action).GetActionDemandComplied()
	if view == nil {
		t.Fatalf("bystander should see ActionDemandComplied")
	}
	masked, ok := view.GetTransfer().(*deal_no_mercy_schema.ActionDemandComplied_MaskedTransferPickpocket)
	if !ok {
		t.Fatalf("bystander should see MaskedTransferPickpocket, got %T", view.GetTransfer())
	}
	mtp := masked.MaskedTransferPickpocket
	if mtp.GetNumCards() != 2 {
		t.Fatalf("masked NumCards = %d, want 2", mtp.GetNumCards())
	}
	if mtp.GetCategory() != deal_no_mercy_schema.StealCategory_STEAL_CATEGORY_ACTION {
		t.Fatalf("masked category = %v, want ACTION", mtp.GetCategory())
	}
	if mtp.GetSourceId() != victim.String() {
		t.Fatalf("masked SourceId = %s, want %s", mtp.GetSourceId(), victim.String())
	}
	if mtp.GetTargetId() != attacker.String() {
		t.Fatalf("masked TargetId = %s, want %s", mtp.GetTargetId(), attacker.String())
	}
}

// A non-pickpocket demand-complied (e.g. a payment) is public and must pass
// through unchanged for every viewer.
func TestMaskDemandCompliedNonPickpocketUnchanged(t *testing.T) {
	c := &Controller{}
	payer := uuid.New()
	creditor := uuid.New()
	bystander := uuid.New()

	engineAction := deal_no_mercy.NewActionDemandComplied(0, payer, deal_no_mercy.Identifier("d1"))
	engineAction.TransferCards = &deal_no_mercy.TransferCards{
		SourceID: payer,
		TargetID: creditor,
		Cards:    deal_no_mercy.Cards{sampleCard("001")},
	}
	action := engineAction.Proto()

	view := c.MaskAction(tp(bystander), action).GetActionDemandComplied()
	if _, ok := view.GetTransfer().(*deal_no_mercy_schema.ActionDemandComplied_TransferCards); !ok {
		t.Fatalf("bystander should see real TransferCards, got %T", view.GetTransfer())
	}
}
