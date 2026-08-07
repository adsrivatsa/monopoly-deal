package deal_no_mercy

// Engine <-> protobuf conversion layer for Monopoly Deal: No Mercy. Mirrors
// the classic engine's Proto()/FromProto idioms. Every engine value is
// representable; conversions are total (nil-safe where the engine uses
// pointers). Masking (hiding hidden information from other players) is NOT
// done here — it belongs to the Phase 3 service layer; this file only
// produces the full, unmasked wire records plus the masked variants the
// service will select between.

import (
	"the-deal/internal/schema/deal_no_mercy_schema"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Enums
// ---------------------------------------------------------------------------

var AssetKeyProtoMap = map[AssetKey]deal_no_mercy_schema.AssetKey{
	AssetKeyUnspecified: deal_no_mercy_schema.AssetKey_ASSET_KEY_UNSPECIFIED,

	AssetKeyBrown1:  deal_no_mercy_schema.AssetKey_ASSET_KEY_BROWN1,
	AssetKeyBrown2:  deal_no_mercy_schema.AssetKey_ASSET_KEY_BROWN2,
	AssetKeySky1:    deal_no_mercy_schema.AssetKey_ASSET_KEY_SKY1,
	AssetKeySky2:    deal_no_mercy_schema.AssetKey_ASSET_KEY_SKY2,
	AssetKeySky3:    deal_no_mercy_schema.AssetKey_ASSET_KEY_SKY3,
	AssetKeyPink1:   deal_no_mercy_schema.AssetKey_ASSET_KEY_PINK1,
	AssetKeyPink2:   deal_no_mercy_schema.AssetKey_ASSET_KEY_PINK2,
	AssetKeyPink3:   deal_no_mercy_schema.AssetKey_ASSET_KEY_PINK3,
	AssetKeyOrange1: deal_no_mercy_schema.AssetKey_ASSET_KEY_ORANGE1,
	AssetKeyOrange2: deal_no_mercy_schema.AssetKey_ASSET_KEY_ORANGE2,
	AssetKeyOrange3: deal_no_mercy_schema.AssetKey_ASSET_KEY_ORANGE3,
	AssetKeyRed1:    deal_no_mercy_schema.AssetKey_ASSET_KEY_RED1,
	AssetKeyRed2:    deal_no_mercy_schema.AssetKey_ASSET_KEY_RED2,
	AssetKeyRed3:    deal_no_mercy_schema.AssetKey_ASSET_KEY_RED3,
	AssetKeyYellow1: deal_no_mercy_schema.AssetKey_ASSET_KEY_YELLOW1,
	AssetKeyYellow2: deal_no_mercy_schema.AssetKey_ASSET_KEY_YELLOW2,
	AssetKeyYellow3: deal_no_mercy_schema.AssetKey_ASSET_KEY_YELLOW3,
	AssetKeyGreen1:  deal_no_mercy_schema.AssetKey_ASSET_KEY_GREEN1,
	AssetKeyGreen2:  deal_no_mercy_schema.AssetKey_ASSET_KEY_GREEN2,
	AssetKeyGreen3:  deal_no_mercy_schema.AssetKey_ASSET_KEY_GREEN3,
	AssetKeyBlue1:   deal_no_mercy_schema.AssetKey_ASSET_KEY_BLUE1,
	AssetKeyBlue2:   deal_no_mercy_schema.AssetKey_ASSET_KEY_BLUE2,
	AssetKeyUtil1:   deal_no_mercy_schema.AssetKey_ASSET_KEY_UTIL1,
	AssetKeyUtil2:   deal_no_mercy_schema.AssetKey_ASSET_KEY_UTIL2,
	AssetKeyRail1:   deal_no_mercy_schema.AssetKey_ASSET_KEY_RAIL1,
	AssetKeyRail2:   deal_no_mercy_schema.AssetKey_ASSET_KEY_RAIL2,
	AssetKeyRail3:   deal_no_mercy_schema.AssetKey_ASSET_KEY_RAIL3,
	AssetKeyRail4:   deal_no_mercy_schema.AssetKey_ASSET_KEY_RAIL4,

	AssetKeyWildBrownSky:        deal_no_mercy_schema.AssetKey_ASSET_KEY_WILD_BROWN_SKY,
	AssetKeyWildSkyRailroad:     deal_no_mercy_schema.AssetKey_ASSET_KEY_WILD_SKY_RAILROAD,
	AssetKeyWildPinkOrange:      deal_no_mercy_schema.AssetKey_ASSET_KEY_WILD_PINK_ORANGE,
	AssetKeyWildRedYellow:       deal_no_mercy_schema.AssetKey_ASSET_KEY_WILD_RED_YELLOW,
	AssetKeyWildGreenBlue:       deal_no_mercy_schema.AssetKey_ASSET_KEY_WILD_GREEN_BLUE,
	AssetKeyWildGreenRailroad:   deal_no_mercy_schema.AssetKey_ASSET_KEY_WILD_GREEN_RAILROAD,
	AssetKeyWildUtilityRailroad: deal_no_mercy_schema.AssetKey_ASSET_KEY_WILD_UTILITY_RAILROAD,
	AssetKeyWildWild:            deal_no_mercy_schema.AssetKey_ASSET_KEY_WILD_WILD,

	AssetKeyMoney1:  deal_no_mercy_schema.AssetKey_ASSET_KEY_MONEY_1,
	AssetKeyMoney2:  deal_no_mercy_schema.AssetKey_ASSET_KEY_MONEY_2,
	AssetKeyMoney3:  deal_no_mercy_schema.AssetKey_ASSET_KEY_MONEY_3,
	AssetKeyMoney4:  deal_no_mercy_schema.AssetKey_ASSET_KEY_MONEY_4,
	AssetKeyMoney5:  deal_no_mercy_schema.AssetKey_ASSET_KEY_MONEY_5,
	AssetKeyMoney10: deal_no_mercy_schema.AssetKey_ASSET_KEY_MONEY_10,
	AssetKeyMoney15: deal_no_mercy_schema.AssetKey_ASSET_KEY_MONEY_15,

	AssetKeySetSnatcher:  deal_no_mercy_schema.AssetKey_ASSET_KEY_SET_SNATCHER,
	AssetKeyDebtTrap:     deal_no_mercy_schema.AssetKey_ASSET_KEY_DEBT_TRAP,
	AssetKeyGoAgain:      deal_no_mercy_schema.AssetKey_ASSET_KEY_GO_AGAIN,
	AssetKeyHeist:        deal_no_mercy_schema.AssetKey_ASSET_KEY_HEIST,
	AssetKeyMarketCrash:  deal_no_mercy_schema.AssetKey_ASSET_KEY_MARKET_CRASH,
	AssetKeyBigPayday:    deal_no_mercy_schema.AssetKey_ASSET_KEY_BIG_PAYDAY,
	AssetKeyRepoMan:      deal_no_mercy_schema.AssetKey_ASSET_KEY_REPO_MAN,
	AssetKeyShack:        deal_no_mercy_schema.AssetKey_ASSET_KEY_SHACK,
	AssetKeyPropertyRaid: deal_no_mercy_schema.AssetKey_ASSET_KEY_PROPERTY_RAID,
	AssetKeyTaxDay:       deal_no_mercy_schema.AssetKey_ASSET_KEY_TAX_DAY,
	AssetKeyPickpocket:   deal_no_mercy_schema.AssetKey_ASSET_KEY_PICKPOCKET,
	AssetKeyBankSwap:     deal_no_mercy_schema.AssetKey_ASSET_KEY_BANK_SWAP,
	AssetKeyYoink:        deal_no_mercy_schema.AssetKey_ASSET_KEY_YOINK,
	AssetKeyNah:          deal_no_mercy_schema.AssetKey_ASSET_KEY_NAH,

	AssetKeyRentWild:            deal_no_mercy_schema.AssetKey_ASSET_KEY_RENT_WILD,
	AssetKeyRentBrownSky:        deal_no_mercy_schema.AssetKey_ASSET_KEY_RENT_BROWN_SKY,
	AssetKeyRentPinkOrange:      deal_no_mercy_schema.AssetKey_ASSET_KEY_RENT_PINK_ORANGE,
	AssetKeyRentRedYellow:       deal_no_mercy_schema.AssetKey_ASSET_KEY_RENT_RED_YELLOW,
	AssetKeyRentGreenBlue:       deal_no_mercy_schema.AssetKey_ASSET_KEY_RENT_GREEN_BLUE,
	AssetKeyRentUtilityRailroad: deal_no_mercy_schema.AssetKey_ASSET_KEY_RENT_UTILITY_RAILROAD,

	AssetKeyDoubleRentWild:            deal_no_mercy_schema.AssetKey_ASSET_KEY_DOUBLE_RENT_WILD,
	AssetKeyDoubleRentBrownSky:        deal_no_mercy_schema.AssetKey_ASSET_KEY_DOUBLE_RENT_BROWN_SKY,
	AssetKeyDoubleRentPinkOrange:      deal_no_mercy_schema.AssetKey_ASSET_KEY_DOUBLE_RENT_PINK_ORANGE,
	AssetKeyDoubleRentRedYellow:       deal_no_mercy_schema.AssetKey_ASSET_KEY_DOUBLE_RENT_RED_YELLOW,
	AssetKeyDoubleRentGreenBlue:       deal_no_mercy_schema.AssetKey_ASSET_KEY_DOUBLE_RENT_GREEN_BLUE,
	AssetKeyDoubleRentUtilityRailroad: deal_no_mercy_schema.AssetKey_ASSET_KEY_DOUBLE_RENT_UTILITY_RAILROAD,
}

func (a AssetKey) Proto() deal_no_mercy_schema.AssetKey {
	return AssetKeyProtoMap[a]
}

var CategoryProtoMap = map[Category]deal_no_mercy_schema.Category{
	CategoryUnspecified:  deal_no_mercy_schema.Category_CATEGORY_UNSPECIFIED,
	CategoryPureProperty: deal_no_mercy_schema.Category_CATEGORY_PURE_PROPERTY,
	CategoryWildProperty: deal_no_mercy_schema.Category_CATEGORY_WILD_PROPERTY,
	CategoryMoney:        deal_no_mercy_schema.Category_CATEGORY_MONEY,
	CategoryAction:       deal_no_mercy_schema.Category_CATEGORY_ACTION,
}

func (c Category) Proto() deal_no_mercy_schema.Category {
	return CategoryProtoMap[c]
}

var ColorProtoMap = map[Color]deal_no_mercy_schema.Color{
	ColorUnspecified: deal_no_mercy_schema.Color_COLOR_UNSPECIFIED,
	ColorBrown:       deal_no_mercy_schema.Color_COLOR_BROWN,
	ColorSky:         deal_no_mercy_schema.Color_COLOR_SKY,
	ColorPink:        deal_no_mercy_schema.Color_COLOR_PINK,
	ColorOrange:      deal_no_mercy_schema.Color_COLOR_ORANGE,
	ColorRed:         deal_no_mercy_schema.Color_COLOR_RED,
	ColorYellow:      deal_no_mercy_schema.Color_COLOR_YELLOW,
	ColorGreen:       deal_no_mercy_schema.Color_COLOR_GREEN,
	ColorBlue:        deal_no_mercy_schema.Color_COLOR_BLUE,
	ColorUtility:     deal_no_mercy_schema.Color_COLOR_UTILITY,
	ColorRailroad:    deal_no_mercy_schema.Color_COLOR_RAILROAD,
}

func (c Color) Proto() deal_no_mercy_schema.Color {
	return ColorProtoMap[c]
}

var ProtoColorMap = map[deal_no_mercy_schema.Color]Color{
	deal_no_mercy_schema.Color_COLOR_UNSPECIFIED: ColorUnspecified,
	deal_no_mercy_schema.Color_COLOR_BROWN:       ColorBrown,
	deal_no_mercy_schema.Color_COLOR_SKY:         ColorSky,
	deal_no_mercy_schema.Color_COLOR_PINK:        ColorPink,
	deal_no_mercy_schema.Color_COLOR_ORANGE:      ColorOrange,
	deal_no_mercy_schema.Color_COLOR_RED:         ColorRed,
	deal_no_mercy_schema.Color_COLOR_YELLOW:      ColorYellow,
	deal_no_mercy_schema.Color_COLOR_GREEN:       ColorGreen,
	deal_no_mercy_schema.Color_COLOR_BLUE:        ColorBlue,
	deal_no_mercy_schema.Color_COLOR_UTILITY:     ColorUtility,
	deal_no_mercy_schema.Color_COLOR_RAILROAD:    ColorRailroad,
}

func ColorFromProto(c deal_no_mercy_schema.Color) Color {
	return ProtoColorMap[c]
}

var StealCategoryProtoMap = map[StealCategory]deal_no_mercy_schema.StealCategory{
	StealCategoryUnspecified: deal_no_mercy_schema.StealCategory_STEAL_CATEGORY_UNSPECIFIED,
	StealCategoryProperty:    deal_no_mercy_schema.StealCategory_STEAL_CATEGORY_PROPERTY,
	StealCategoryMoney:       deal_no_mercy_schema.StealCategory_STEAL_CATEGORY_MONEY,
	StealCategoryAction:      deal_no_mercy_schema.StealCategory_STEAL_CATEGORY_ACTION,
}

func (sc StealCategory) Proto() deal_no_mercy_schema.StealCategory {
	return StealCategoryProtoMap[sc]
}

var ProtoStealCategoryMap = map[deal_no_mercy_schema.StealCategory]StealCategory{
	deal_no_mercy_schema.StealCategory_STEAL_CATEGORY_UNSPECIFIED: StealCategoryUnspecified,
	deal_no_mercy_schema.StealCategory_STEAL_CATEGORY_PROPERTY:    StealCategoryProperty,
	deal_no_mercy_schema.StealCategory_STEAL_CATEGORY_MONEY:       StealCategoryMoney,
	deal_no_mercy_schema.StealCategory_STEAL_CATEGORY_ACTION:      StealCategoryAction,
}

func StealCategoryFromProto(sc deal_no_mercy_schema.StealCategory) StealCategory {
	return ProtoStealCategoryMap[sc]
}

var SpeedProtoMap = map[Speed]deal_no_mercy_schema.Speed{
	SpeedUnidentified: deal_no_mercy_schema.Speed_SPEED_UNSPECIFIED,
	SpeedSlow:         deal_no_mercy_schema.Speed_SPEED_SLOW,
	SpeedMedium:       deal_no_mercy_schema.Speed_SPEED_MEDIUM,
	SpeedFast:         deal_no_mercy_schema.Speed_SPEED_FAST,
}

func (s Speed) Proto() deal_no_mercy_schema.Speed {
	return SpeedProtoMap[s]
}

var DemandKindProtoMap = map[DemandKind]deal_no_mercy_schema.DemandKind{
	DemandKindUnspecified:     deal_no_mercy_schema.DemandKind_DEMAND_KIND_UNSPECIFIED,
	DemandKindPayment:         deal_no_mercy_schema.DemandKind_DEMAND_KIND_PAYMENT,
	DemandKindProperty:        deal_no_mercy_schema.DemandKind_DEMAND_KIND_PROPERTY,
	DemandKindPropertySet:     deal_no_mercy_schema.DemandKind_DEMAND_KIND_PROPERTY_SET,
	DemandKindColorProperties: deal_no_mercy_schema.DemandKind_DEMAND_KIND_COLOR_PROPERTIES,
	DemandKindBankCard:        deal_no_mercy_schema.DemandKind_DEMAND_KIND_BANK_CARD,
	DemandKindBankSwap:        deal_no_mercy_schema.DemandKind_DEMAND_KIND_BANK_SWAP,
	DemandKindDebtTrap:        deal_no_mercy_schema.DemandKind_DEMAND_KIND_DEBT_TRAP,
	DemandKindRepoMan:         deal_no_mercy_schema.DemandKind_DEMAND_KIND_REPO_MAN,
	DemandKindTaxDay:          deal_no_mercy_schema.DemandKind_DEMAND_KIND_TAX_DAY,
	DemandKindPickpocket:      deal_no_mercy_schema.DemandKind_DEMAND_KIND_PICKPOCKET,
}

func (dk DemandKind) Proto() deal_no_mercy_schema.DemandKind {
	return DemandKindProtoMap[dk]
}

var DemandSourceProtoMap = map[DemandSource]deal_no_mercy_schema.DemandSource{
	DemandSourceUnspecified:  deal_no_mercy_schema.DemandSource_DEMAND_SOURCE_UNSPECIFIED,
	DemandSourceRent:         deal_no_mercy_schema.DemandSource_DEMAND_SOURCE_RENT,
	DemandSourceYoink:        deal_no_mercy_schema.DemandSource_DEMAND_SOURCE_YOINK,
	DemandSourceSetSnatcher:  deal_no_mercy_schema.DemandSource_DEMAND_SOURCE_SET_SNATCHER,
	DemandSourceMarketCrash:  deal_no_mercy_schema.DemandSource_DEMAND_SOURCE_MARKET_CRASH,
	DemandSourcePropertyRaid: deal_no_mercy_schema.DemandSource_DEMAND_SOURCE_PROPERTY_RAID,
	DemandSourceHeist:        deal_no_mercy_schema.DemandSource_DEMAND_SOURCE_HEIST,
	DemandSourceBankSwap:     deal_no_mercy_schema.DemandSource_DEMAND_SOURCE_BANK_SWAP,
	DemandSourceDebtTrap:     deal_no_mercy_schema.DemandSource_DEMAND_SOURCE_DEBT_TRAP,
	DemandSourceRepoMan:      deal_no_mercy_schema.DemandSource_DEMAND_SOURCE_REPO_MAN,
	DemandSourceTaxDay:       deal_no_mercy_schema.DemandSource_DEMAND_SOURCE_TAX_DAY,
	DemandSourcePickpocket:   deal_no_mercy_schema.DemandSource_DEMAND_SOURCE_PICKPOCKET,
	DemandSourceNah:          deal_no_mercy_schema.DemandSource_DEMAND_SOURCE_NAH,
}

func (ds DemandSource) Proto() deal_no_mercy_schema.DemandSource {
	return DemandSourceProtoMap[ds]
}

var ActionKindProtoMap = map[ActionKind]deal_no_mercy_schema.ActionKind{
	ActionKindUnspecified:    deal_no_mercy_schema.ActionKind_ACTION_KIND_UNSPECIFIED,
	ActionKindPlayMoney:      deal_no_mercy_schema.ActionKind_ACTION_KIND_PLAY_MONEY,
	ActionKindPlayProperty:   deal_no_mercy_schema.ActionKind_ACTION_KIND_PLAY_PROPERTY,
	ActionKindPlayShack:      deal_no_mercy_schema.ActionKind_ACTION_KIND_PLAY_SHACK,
	ActionKindPlayBigPayday:  deal_no_mercy_schema.ActionKind_ACTION_KIND_PLAY_BIG_PAYDAY,
	ActionKindPlayGoAgain:    deal_no_mercy_schema.ActionKind_ACTION_KIND_PLAY_GO_AGAIN,
	ActionKindDemandCreated:  deal_no_mercy_schema.ActionKind_ACTION_KIND_DEMAND_CREATED,
	ActionKindDemandComplied: deal_no_mercy_schema.ActionKind_ACTION_KIND_DEMAND_COMPLIED,
	ActionKindDebtSettled:    deal_no_mercy_schema.ActionKind_ACTION_KIND_DEBT_SETTLED,
	ActionKindDiscardCards:   deal_no_mercy_schema.ActionKind_ACTION_KIND_DISCARD_CARDS,
	ActionKindStartTurn:      deal_no_mercy_schema.ActionKind_ACTION_KIND_START_TURN,
	ActionKindRearrangeCard:  deal_no_mercy_schema.ActionKind_ACTION_KIND_REARRANGE_CARD,
}

func (a ActionKind) Proto() deal_no_mercy_schema.ActionKind {
	return ActionKindProtoMap[a]
}

// ---------------------------------------------------------------------------
// Cards / property sets
// ---------------------------------------------------------------------------

func (c Card) Proto() *deal_no_mercy_schema.Card {
	colors := make([]deal_no_mercy_schema.Color, len(c.Colors))
	for i, color := range c.Colors {
		colors[i] = color.Proto()
	}

	return &deal_no_mercy_schema.Card{
		CardId:      string(c.ID),
		AssetKey:    c.AssetKey.Proto(),
		Category:    c.Category.Proto(),
		ActiveColor: c.ActiveColor.Proto(),
		Colors:      colors,
		Value:       int32(c.Value),
	}
}

// cardPtrProto is a nil-safe helper for the many optional *Card fields.
func cardPtrProto(c *Card) *deal_no_mercy_schema.Card {
	if c == nil {
		return nil
	}
	return c.Proto()
}

func (c *Cards) Proto() []*deal_no_mercy_schema.Card {
	if c == nil {
		return nil
	}

	cards := make([]*deal_no_mercy_schema.Card, len(*c))
	for i, card := range *c {
		cards[i] = card.Proto()
	}
	return cards
}

func (ps *PropertySet) Proto(playerID uuid.UUID) *deal_no_mercy_schema.PropertySet {
	return &deal_no_mercy_schema.PropertySet{
		PlayerId:      playerID.String(),
		PropertySetId: string(ps.ID),
		Color:         ps.Color.Proto(),
		Cards:         ps.Cards.Proto(),
		Shack:         cardPtrProto(ps.Shack),
	}
}

func (ps *PropertySets) Proto(playerID uuid.UUID) []*deal_no_mercy_schema.PropertySet {
	if ps == nil {
		return nil
	}

	sets := make([]*deal_no_mercy_schema.PropertySet, len(*ps))
	for i, p := range *ps {
		sets[i] = p.Proto(playerID)
	}
	return sets
}

// ---------------------------------------------------------------------------
// Debt
// ---------------------------------------------------------------------------

func (d DebtObligation) Proto() *deal_no_mercy_schema.DebtChip {
	return &deal_no_mercy_schema.DebtChip{
		Id:         string(d.ID),
		DebtorId:   d.DebtorID.String(),
		CreditorId: d.CreditorID.String(),
	}
}

func debtsProto(debts []DebtObligation) []*deal_no_mercy_schema.DebtChip {
	if debts == nil {
		return nil
	}
	out := make([]*deal_no_mercy_schema.DebtChip, len(debts))
	for i, d := range debts {
		out[i] = d.Proto()
	}
	return out
}

// ---------------------------------------------------------------------------
// Demand
// ---------------------------------------------------------------------------

func (d *Demand) Proto() *deal_no_mercy_schema.Demand {
	out := &deal_no_mercy_schema.Demand{
		Id:           string(d.ID),
		PlayerId:     d.TargetID.String(),
		SourceId:     d.SourceID.String(),
		DemandKind:   d.Kind.Proto(),
		IsActive:     d.IsActive,
		DemandSource: d.Source.Proto(),
	}

	switch d.Kind {
	case DemandKindPayment:
		out.Demand = &deal_no_mercy_schema.Demand_PaymentDemand{
			PaymentDemand: &deal_no_mercy_schema.PaymentDemand{
				Amount: int32(d.Payment.Amount),
			},
		}
	case DemandKindProperty:
		out.Demand = &deal_no_mercy_schema.Demand_PropertyDemand{
			PropertyDemand: &deal_no_mercy_schema.PropertyDemand{
				TargetCardId: string(d.Property.TargetCardID),
			},
		}
	case DemandKindPropertySet:
		out.Demand = &deal_no_mercy_schema.Demand_PropertySetDemand{
			PropertySetDemand: &deal_no_mercy_schema.PropertySetDemand{
				PropertySetId: string(d.PropertySet.PropertySetID),
			},
		}
	case DemandKindColorProperties:
		out.Demand = &deal_no_mercy_schema.Demand_ColorPropertiesDemand{
			ColorPropertiesDemand: &deal_no_mercy_schema.ColorPropertiesDemand{
				Color: d.ColorProperties.Color.Proto(),
			},
		}
	case DemandKindBankCard:
		out.Demand = &deal_no_mercy_schema.Demand_BankCardDemand{
			BankCardDemand: &deal_no_mercy_schema.BankCardDemand{
				TargetCardId: string(d.BankCard.TargetCardID),
			},
		}
	case DemandKindBankSwap:
		out.Demand = &deal_no_mercy_schema.Demand_BankSwapDemand{
			BankSwapDemand: &deal_no_mercy_schema.BankSwapDemand{},
		}
	case DemandKindDebtTrap:
		out.Demand = &deal_no_mercy_schema.Demand_DebtTrapDemand{
			DebtTrapDemand: &deal_no_mercy_schema.DebtTrapDemand{},
		}
	case DemandKindRepoMan:
		out.Demand = &deal_no_mercy_schema.Demand_RepoManDemand{
			RepoManDemand: &deal_no_mercy_schema.RepoManDemand{},
		}
	case DemandKindTaxDay:
		out.Demand = &deal_no_mercy_schema.Demand_TaxDayDemand{
			TaxDayDemand: &deal_no_mercy_schema.TaxDayDemand{},
		}
	case DemandKindPickpocket:
		out.Demand = &deal_no_mercy_schema.Demand_PickpocketDemand{
			PickpocketDemand: &deal_no_mercy_schema.PickpocketDemand{
				Category: d.Pickpocket.Category.Proto(),
			},
		}
	}

	return out
}

// ---------------------------------------------------------------------------
// Settings
// ---------------------------------------------------------------------------

func (s *Settings) Proto() *deal_no_mercy_schema.Settings {
	return &deal_no_mercy_schema.Settings{
		Speed:               s.Speed.Proto(),
		MaxHandSize:         int32(s.MaxHandSize),
		MovesPerTurn:        int32(s.MovesPerTurn),
		WinSetAmount:        int32(s.WinSetAmount),
		DebtChipsPerPlayer:  int32(s.DebtChipsPerPlayer),
		YoinkPayment:        int32(s.YoinkPayment),
		ShackRentBonus:      int32(s.ShackRentBonus),
		BigPaydayHandTarget: int32(s.BigPaydayHandTarget),
		NahConsumesMove:     s.NahConsumesMove,
	}
}

// ---------------------------------------------------------------------------
// GameState
// ---------------------------------------------------------------------------

// Proto renders the game from playerID's perspective: their own hand is
// revealed, all money/properties/demands/debts are public. Players,
// Deadlines and AssetImages are populated by the service-layer caller
// (mirrors the classic engine).
func (g *Game) Proto(playerID uuid.UUID) *deal_no_mercy_schema.GameState {
	currPlayerID := g.Players[g.CurrPlayerIdx]

	hand := g.Hands[playerID]
	handProto := &deal_no_mercy_schema.Hand{
		Cards: hand.Proto(),
	}

	monies := make([]*deal_no_mercy_schema.Money, 0, len(g.Players))
	var properties []*deal_no_mercy_schema.PropertySet
	for _, id := range g.Players {
		money := g.Money[id]
		monies = append(monies, &deal_no_mercy_schema.Money{
			PlayerId: id.String(),
			Cards:    money.Proto(),
		})

		property := g.Properties[id]
		properties = append(properties, property.Proto(id)...)
	}

	demandsProto := make([]*deal_no_mercy_schema.Demand, 0, len(g.Demands))
	for _, demand := range g.Demands {
		demandsProto = append(demandsProto, demand.Proto())
	}

	return &deal_no_mercy_schema.GameState{
		SeqNum:          int32(g.SequenceNum),
		Players:         nil, // populated by caller
		CurrentPlayerId: currPlayerID.String(),
		MovesLeft:       int32(g.MovesLeft),
		YourHand:        handProto,
		Money:           monies,
		Properties:      properties,
		Demands:         demandsProto,
		DebtChips:       debtsProto(g.Debts),
		LastAction:      g.LastAction.Proto(),
		AssetImages:     nil, // populated by caller
		MaxHandSize:     int32(g.Config.MaxHandSize),
		Deadlines:       nil, // populated by caller
		Settings:        g.Config.Proto(),
	}
}
