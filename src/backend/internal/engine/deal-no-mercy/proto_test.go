package deal_no_mercy

import (
	"testing"

	"the-deal/internal/schema/deal_no_mercy_schema"

	"github.com/google/uuid"
)

func TestCardProto(t *testing.T) {
	card := NewCard("00a", CategoryWildProperty, AssetKeyWildPinkOrange, 2, ColorPink, ColorOrange)

	p := card.Proto()
	if p.GetCardId() != "00a" {
		t.Fatalf("card id = %q, want 00a", p.GetCardId())
	}
	if p.GetAssetKey() != deal_no_mercy_schema.AssetKey_ASSET_KEY_WILD_PINK_ORANGE {
		t.Fatalf("asset key = %v", p.GetAssetKey())
	}
	if p.GetCategory() != deal_no_mercy_schema.Category_CATEGORY_WILD_PROPERTY {
		t.Fatalf("category = %v", p.GetCategory())
	}
	if p.GetActiveColor() != deal_no_mercy_schema.Color_COLOR_PINK {
		t.Fatalf("active color = %v", p.GetActiveColor())
	}
	if p.GetValue() != 2 {
		t.Fatalf("value = %d", p.GetValue())
	}
	if len(p.GetColors()) != 2 {
		t.Fatalf("colors len = %d, want 2", len(p.GetColors()))
	}
}

// TestAllAssetKeysProto asserts every engine AssetKey maps to a non-unspecified
// proto value (i.e. the conversion is total). The 15M money card and all No
// Mercy actions are the risky additions.
func TestAllAssetKeysProto(t *testing.T) {
	for _, key := range AllAssetKeys() {
		if key.Proto() == deal_no_mercy_schema.AssetKey_ASSET_KEY_UNSPECIFIED {
			t.Errorf("asset key %q has no proto mapping", key)
		}
	}
	// Spot-check the brand-new denomination.
	if AssetKeyMoney15.Proto() != deal_no_mercy_schema.AssetKey_ASSET_KEY_MONEY_15 {
		t.Errorf("money15 mapped to %v", AssetKeyMoney15.Proto())
	}
}

func TestColorRoundTrip(t *testing.T) {
	for _, c := range append(AllColors(), ColorUnspecified) {
		got := ColorFromProto(c.Proto())
		if got != c {
			t.Errorf("color %v round-tripped to %v", c, got)
		}
	}
}

func TestStealCategoryRoundTrip(t *testing.T) {
	for _, sc := range []StealCategory{StealCategoryUnspecified, StealCategoryProperty, StealCategoryMoney, StealCategoryAction} {
		got := StealCategoryFromProto(sc.Proto())
		if got != sc {
			t.Errorf("steal category %v round-tripped to %v", sc, got)
		}
	}
}

func TestPropertySetProtoWithShack(t *testing.T) {
	shack := NewCard("s1", CategoryAction, AssetKeyShack, 3)
	set := NewPropertySet("set1", ColorBlue)
	set.Cards.Add(NewCard("b1", CategoryPureProperty, AssetKeyBlue1, 4, ColorBlue))
	set.Shack = &shack

	playerID := uuid.New()
	p := set.Proto(playerID)

	if p.GetPlayerId() != playerID.String() {
		t.Fatalf("player id mismatch")
	}
	if p.GetPropertySetId() != "set1" {
		t.Fatalf("set id = %q", p.GetPropertySetId())
	}
	if p.GetColor() != deal_no_mercy_schema.Color_COLOR_BLUE {
		t.Fatalf("color = %v", p.GetColor())
	}
	if len(p.GetCards()) != 1 {
		t.Fatalf("cards len = %d", len(p.GetCards()))
	}
	if p.GetShack() == nil || p.GetShack().GetAssetKey() != deal_no_mercy_schema.AssetKey_ASSET_KEY_SHACK {
		t.Fatalf("shack not preserved: %v", p.GetShack())
	}

	// No shack -> nil in proto.
	set.Shack = nil
	if set.Proto(playerID).GetShack() != nil {
		t.Fatalf("expected nil shack")
	}
}

func TestDemandProtoAllKinds(t *testing.T) {
	src := uuid.New()
	tgt := uuid.New()

	cases := []struct {
		name       string
		demand     Demand
		wantKind   deal_no_mercy_schema.DemandKind
		wantSource deal_no_mercy_schema.DemandSource
		check      func(t *testing.T, p *deal_no_mercy_schema.Demand)
	}{
		{
			name:       "payment",
			demand:     NewPaymentDemand("d1", src, tgt, 7, DemandSourceRent),
			wantKind:   deal_no_mercy_schema.DemandKind_DEMAND_KIND_PAYMENT,
			wantSource: deal_no_mercy_schema.DemandSource_DEMAND_SOURCE_RENT,
			check: func(t *testing.T, p *deal_no_mercy_schema.Demand) {
				if p.GetPaymentDemand().GetAmount() != 7 {
					t.Errorf("amount = %d", p.GetPaymentDemand().GetAmount())
				}
			},
		},
		{
			name:       "property",
			demand:     NewPropertyDemand("d2", src, tgt, "c9", DemandSourceMarketCrash),
			wantKind:   deal_no_mercy_schema.DemandKind_DEMAND_KIND_PROPERTY,
			wantSource: deal_no_mercy_schema.DemandSource_DEMAND_SOURCE_MARKET_CRASH,
			check: func(t *testing.T, p *deal_no_mercy_schema.Demand) {
				if p.GetPropertyDemand().GetTargetCardId() != "c9" {
					t.Errorf("target card = %q", p.GetPropertyDemand().GetTargetCardId())
				}
			},
		},
		{
			name:       "property_set",
			demand:     NewPropertySetDemand("d3", src, tgt, "ps1", DemandSourceSetSnatcher),
			wantKind:   deal_no_mercy_schema.DemandKind_DEMAND_KIND_PROPERTY_SET,
			wantSource: deal_no_mercy_schema.DemandSource_DEMAND_SOURCE_SET_SNATCHER,
			check: func(t *testing.T, p *deal_no_mercy_schema.Demand) {
				if p.GetPropertySetDemand().GetPropertySetId() != "ps1" {
					t.Errorf("set id = %q", p.GetPropertySetDemand().GetPropertySetId())
				}
			},
		},
		{
			name:       "color_properties",
			demand:     NewColorPropertiesDemand("d4", src, tgt, ColorGreen, DemandSourcePropertyRaid),
			wantKind:   deal_no_mercy_schema.DemandKind_DEMAND_KIND_COLOR_PROPERTIES,
			wantSource: deal_no_mercy_schema.DemandSource_DEMAND_SOURCE_PROPERTY_RAID,
			check: func(t *testing.T, p *deal_no_mercy_schema.Demand) {
				if p.GetColorPropertiesDemand().GetColor() != deal_no_mercy_schema.Color_COLOR_GREEN {
					t.Errorf("color = %v", p.GetColorPropertiesDemand().GetColor())
				}
			},
		},
		{
			name:       "bank_card",
			demand:     NewBankCardDemand("d5", src, tgt, "c3", DemandSourceHeist),
			wantKind:   deal_no_mercy_schema.DemandKind_DEMAND_KIND_BANK_CARD,
			wantSource: deal_no_mercy_schema.DemandSource_DEMAND_SOURCE_HEIST,
			check: func(t *testing.T, p *deal_no_mercy_schema.Demand) {
				if p.GetBankCardDemand().GetTargetCardId() != "c3" {
					t.Errorf("target card = %q", p.GetBankCardDemand().GetTargetCardId())
				}
			},
		},
		{
			name:       "bank_swap",
			demand:     NewBankSwapDemand("d6", src, tgt),
			wantKind:   deal_no_mercy_schema.DemandKind_DEMAND_KIND_BANK_SWAP,
			wantSource: deal_no_mercy_schema.DemandSource_DEMAND_SOURCE_BANK_SWAP,
			check: func(t *testing.T, p *deal_no_mercy_schema.Demand) {
				if p.GetBankSwapDemand() == nil {
					t.Errorf("bank swap payload missing")
				}
			},
		},
		{
			name:       "debt_trap",
			demand:     NewDebtTrapDemand("d7", src, tgt),
			wantKind:   deal_no_mercy_schema.DemandKind_DEMAND_KIND_DEBT_TRAP,
			wantSource: deal_no_mercy_schema.DemandSource_DEMAND_SOURCE_DEBT_TRAP,
			check: func(t *testing.T, p *deal_no_mercy_schema.Demand) {
				if p.GetDebtTrapDemand() == nil {
					t.Errorf("debt trap payload missing")
				}
			},
		},
		{
			name:       "repo_man",
			demand:     NewRepoManDemand("d8", src, tgt),
			wantKind:   deal_no_mercy_schema.DemandKind_DEMAND_KIND_REPO_MAN,
			wantSource: deal_no_mercy_schema.DemandSource_DEMAND_SOURCE_REPO_MAN,
			check: func(t *testing.T, p *deal_no_mercy_schema.Demand) {
				if p.GetRepoManDemand() == nil {
					t.Errorf("repo man payload missing")
				}
			},
		},
		{
			name:       "tax_day",
			demand:     NewTaxDayDemand("d9", src, tgt),
			wantKind:   deal_no_mercy_schema.DemandKind_DEMAND_KIND_TAX_DAY,
			wantSource: deal_no_mercy_schema.DemandSource_DEMAND_SOURCE_TAX_DAY,
			check: func(t *testing.T, p *deal_no_mercy_schema.Demand) {
				if p.GetTaxDayDemand() == nil {
					t.Errorf("tax day payload missing")
				}
			},
		},
		{
			name:       "pickpocket",
			demand:     NewPickpocketDemand("d10", src, tgt, StealCategoryAction),
			wantKind:   deal_no_mercy_schema.DemandKind_DEMAND_KIND_PICKPOCKET,
			wantSource: deal_no_mercy_schema.DemandSource_DEMAND_SOURCE_PICKPOCKET,
			check: func(t *testing.T, p *deal_no_mercy_schema.Demand) {
				if p.GetPickpocketDemand().GetCategory() != deal_no_mercy_schema.StealCategory_STEAL_CATEGORY_ACTION {
					t.Errorf("category = %v", p.GetPickpocketDemand().GetCategory())
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := tc.demand
			p := d.Proto()
			if p.GetDemandKind() != tc.wantKind {
				t.Fatalf("kind = %v, want %v", p.GetDemandKind(), tc.wantKind)
			}
			if p.GetDemandSource() != tc.wantSource {
				t.Fatalf("source = %v, want %v", p.GetDemandSource(), tc.wantSource)
			}
			// The demand target becomes the proto PlayerId (who must respond).
			if p.GetPlayerId() != tgt.String() {
				t.Fatalf("player id = %q, want target %q", p.GetPlayerId(), tgt.String())
			}
			if p.GetSourceId() != src.String() {
				t.Fatalf("source id = %q, want %q", p.GetSourceId(), src.String())
			}
			if !p.GetIsActive() {
				t.Fatalf("expected active demand")
			}
			tc.check(t, p)
		})
	}
}

func TestSettingsProto(t *testing.T) {
	s := DefaultSettings()
	p := s.Proto()

	if p.GetSpeed() != deal_no_mercy_schema.Speed_SPEED_MEDIUM {
		t.Errorf("speed = %v", p.GetSpeed())
	}
	if p.GetMaxHandSize() != 7 {
		t.Errorf("max hand = %d", p.GetMaxHandSize())
	}
	if p.GetDebtChipsPerPlayer() != 3 {
		t.Errorf("debt chips = %d", p.GetDebtChipsPerPlayer())
	}
	if p.GetYoinkPayment() != 10 {
		t.Errorf("yoink = %d", p.GetYoinkPayment())
	}
	if p.GetShackRentBonus() != 5 {
		t.Errorf("shack bonus = %d", p.GetShackRentBonus())
	}
	if p.GetBigPaydayHandTarget() != 7 {
		t.Errorf("big payday target = %d", p.GetBigPaydayHandTarget())
	}
	if !p.GetNahConsumesMove() {
		t.Errorf("expected nah consumes move")
	}
}

func TestGameStateProto(t *testing.T) {
	p1 := uuid.New()
	p2 := uuid.New()
	g := NewGame(DefaultSettings(), []uuid.UUID{p1, p2})

	// Seed a debt so DebtChips surfaces in the state.
	g.Debts = append(g.Debts, DebtObligation{ID: "debt1", DebtorID: p2, CreditorID: p1})

	state := g.Proto(p1)

	if state.GetCurrentPlayerId() != p1.String() {
		t.Errorf("current player = %q", state.GetCurrentPlayerId())
	}
	if state.GetMovesLeft() != int32(g.Config.MovesPerTurn) {
		t.Errorf("moves left = %d", state.GetMovesLeft())
	}
	// Own hand is revealed (5 starting cards).
	if len(state.GetYourHand().GetCards()) != g.Config.StartNumCards {
		t.Errorf("hand size = %d, want %d", len(state.GetYourHand().GetCards()), g.Config.StartNumCards)
	}
	if len(state.GetMoney()) != 2 {
		t.Errorf("money piles = %d, want 2", len(state.GetMoney()))
	}
	if len(state.GetDebtChips()) != 1 {
		t.Errorf("debt chips = %d, want 1", len(state.GetDebtChips()))
	}
	if state.GetDebtChips()[0].GetDebtorId() != p2.String() {
		t.Errorf("debtor = %q", state.GetDebtChips()[0].GetDebtorId())
	}
	if state.GetMaxHandSize() != int32(g.Config.MaxHandSize) {
		t.Errorf("max hand = %d", state.GetMaxHandSize())
	}
	if state.GetSettings() == nil {
		t.Errorf("settings missing")
	}
	// Players / Deadlines / AssetImages are the caller's responsibility.
	if state.GetPlayers() != nil {
		t.Errorf("players should be populated by caller, got %d", len(state.GetPlayers()))
	}
}

func TestActionProtoInterface(t *testing.T) {
	playerID := uuid.New()
	card := NewCard("m1", CategoryMoney, AssetKeyMoney5, 5)

	// Every constructed Action must satisfy the interface (compile-time via
	// the slice) and yield a non-nil proto with the right kind and player.
	set := NewPropertySet("ps1", ColorRed)
	set.Cards.Add(NewCard("r1", CategoryPureProperty, AssetKeyRed1, 3, ColorRed))
	debt := DebtObligation{ID: "dbt", DebtorID: playerID, CreditorID: uuid.New()}

	actions := []Action{
		NewActionPlayMoney(1, playerID, card),
		NewActionPlayProperty(2, playerID, set),
		NewActionPlayShack(3, playerID, NewCard("s1", CategoryAction, AssetKeyShack, 3), set),
		NewActionPlayBigPayday(4, playerID, NewCard("bp", CategoryAction, AssetKeyBigPayday, 1), Cards{card}),
		NewActionPlayGoAgain(5, playerID, NewCard("ga", CategoryAction, AssetKeyGoAgain, 1)),
		NewActionDemandsCreated(6, playerID, &card, nil, NewPaymentDemand("d1", playerID, uuid.New(), 3, DemandSourceRent)),
		NewActionDemandComplied(7, playerID, "d1"),
		NewActionDebtSettled(8, playerID, debt, card, nil, false),
		NewActionDiscardCards(9, playerID, card),
		NewActionStartTurn(10, playerID, Cards{card}, 3, []DebtObligation{debt}, true),
		NewActionRearrangeCard(11, playerID, &card, set, 0),
	}

	for _, a := range actions {
		p := a.Proto()
		if p == nil {
			t.Fatalf("action %s produced nil proto", a.GetKind())
		}
		if p.GetPlayerId() != playerID.String() {
			t.Errorf("action %s player id = %q", a.GetKind(), p.GetPlayerId())
		}
		if p.GetKind() != a.GetKind().Proto() {
			t.Errorf("action %s kind mismatch: %v", a.GetKind(), p.GetKind())
		}
		if p.GetPayload() == nil {
			t.Errorf("action %s has nil payload", a.GetKind())
		}
	}
}

func TestActionStartTurnProtoAndMasked(t *testing.T) {
	playerID := uuid.New()
	drawn := Cards{
		NewCard("c1", CategoryMoney, AssetKeyMoney1, 1),
		NewCard("c2", CategoryMoney, AssetKeyMoney2, 2),
	}
	debt := DebtObligation{ID: "d1", DebtorID: playerID, CreditorID: uuid.New()}
	a := NewActionStartTurn(1, playerID, drawn, 3, []DebtObligation{debt}, true)

	full := a.Proto().GetActionStartTurn()
	if len(full.GetCards()) != 2 {
		t.Fatalf("full cards = %d", len(full.GetCards()))
	}
	if !full.GetGoAgain() {
		t.Fatalf("expected go again")
	}
	if len(full.GetDebts()) != 1 {
		t.Fatalf("full debts = %d", len(full.GetDebts()))
	}

	masked := a.MaskedProto().GetMaskedActionStartTurn()
	if masked.GetNumCards() != 2 {
		t.Fatalf("masked num cards = %d", masked.GetNumCards())
	}
	// Debts and go-again survive masking (public info).
	if len(masked.GetDebts()) != 1 || !masked.GetGoAgain() {
		t.Fatalf("masked lost public state: debts=%d goAgain=%v", len(masked.GetDebts()), masked.GetGoAgain())
	}
}

func TestActionDemandCompliedTransfers(t *testing.T) {
	src := uuid.New()
	tgt := uuid.New()
	card := NewCard("c1", CategoryMoney, AssetKeyMoney3, 3)

	// Payment with a debt chip issued on shortfall.
	a := NewActionDemandComplied(1, tgt, "d1")
	a.TransferCards = &TransferCards{
		SourceID:    tgt,
		TargetID:    src,
		Cards:       Cards{card},
		SourceMoney: 0,
		Debt:        &DebtObligation{ID: "chip1", DebtorID: tgt, CreditorID: src},
	}
	tc := a.Proto().GetActionDemandComplied().GetTransferCards()
	if tc == nil {
		t.Fatalf("transfer cards missing")
	}
	if tc.GetDebt() == nil || tc.GetDebt().GetId() != "chip1" {
		t.Fatalf("debt chip not carried on shortfall payment")
	}

	// Pickpocket: full vs masked variant.
	pp := &TransferPickpocket{
		SourceID: src,
		TargetID: tgt,
		Category: StealCategoryProperty,
		Cards:    Cards{card, card},
	}
	if len(pp.Proto().GetCards()) != 2 {
		t.Fatalf("full pickpocket cards = %d", len(pp.Proto().GetCards()))
	}
	if pp.MaskedProto().GetNumCards() != 2 {
		t.Fatalf("masked pickpocket count = %d", pp.MaskedProto().GetNumCards())
	}
}

func TestSnapshotThenProto(t *testing.T) {
	p1 := uuid.New()
	p2 := uuid.New()
	g := NewGame(DefaultSettings(), []uuid.UUID{p1, p2})

	data, err := g.EncodeMsgpack()
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	restored, err := DecodeMsgpack(data)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	// A restored game still converts cleanly.
	state := restored.Proto(p1)
	if state.GetSettings().GetDebtChipsPerPlayer() != 3 {
		t.Fatalf("restored settings lost debt chips knob")
	}
	if len(state.GetMoney()) != 2 {
		t.Fatalf("restored money piles = %d", len(state.GetMoney()))
	}
}
