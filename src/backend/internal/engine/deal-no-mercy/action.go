package deal_no_mercy

import (
	"github.com/google/uuid"
)

// Actions are the engine's outbound event records. Proto() conversions are
// intentionally omitted in Phase 1 — the deal_no_mercy protobuf schema does
// not exist yet; Phase 2/3 adds the wire mapping.

type ActionKind string

const (
	ActionKindUnspecified    ActionKind = "unspecified"
	ActionKindPlayMoney      ActionKind = "play_money"
	ActionKindPlayProperty   ActionKind = "play_property"
	ActionKindPlayShack      ActionKind = "play_shack"
	ActionKindPlayBigPayday  ActionKind = "play_big_payday"
	ActionKindPlayGoAgain    ActionKind = "play_go_again"
	ActionKindDemandCreated  ActionKind = "demand_created"
	ActionKindDemandComplied ActionKind = "demand_complied"
	ActionKindDebtSettled    ActionKind = "debt_settled"
	ActionKindDiscardCards   ActionKind = "discard_cards"
	ActionKindStartTurn      ActionKind = "start_turn"
	ActionKindRearrangeCard  ActionKind = "rearrange_card"
)

type Action interface {
	GetKind() ActionKind
	GetVersion() int
	GetSeqNum() int
}

type ActionPlayMoney struct {
	Kind     ActionKind `msgpack:"a"`
	Version  int        `msgpack:"b"`
	SeqNum   int        `msgpack:"c"`
	PlayerID uuid.UUID  `msgpack:"d"`
	Card     Card       `msgpack:"e"`
}

func NewActionPlayMoney(seqNum int, playerID uuid.UUID, card Card) *ActionPlayMoney {
	return &ActionPlayMoney{
		Kind:     ActionKindPlayMoney,
		Version:  1,
		SeqNum:   seqNum,
		PlayerID: playerID,
		Card:     card,
	}
}

func (a *ActionPlayMoney) GetKind() ActionKind { return a.Kind }
func (a *ActionPlayMoney) GetVersion() int     { return a.Version }
func (a *ActionPlayMoney) GetSeqNum() int      { return a.SeqNum }

type ActionPlayProperty struct {
	Kind        ActionKind  `msgpack:"a"`
	Version     int         `msgpack:"b"`
	SeqNum      int         `msgpack:"c"`
	PlayerID    uuid.UUID   `msgpack:"d"`
	PropertySet PropertySet `msgpack:"e"`
}

func NewActionPlayProperty(seqNum int, playerID uuid.UUID, propertySet PropertySet) *ActionPlayProperty {
	return &ActionPlayProperty{
		Kind:        ActionKindPlayProperty,
		Version:     1,
		SeqNum:      seqNum,
		PlayerID:    playerID,
		PropertySet: propertySet,
	}
}

func (a *ActionPlayProperty) GetKind() ActionKind { return a.Kind }
func (a *ActionPlayProperty) GetVersion() int     { return a.Version }
func (a *ActionPlayProperty) GetSeqNum() int      { return a.SeqNum }

type ActionPlayShack struct {
	Kind        ActionKind  `msgpack:"a"`
	Version     int         `msgpack:"b"`
	SeqNum      int         `msgpack:"c"`
	PlayerID    uuid.UUID   `msgpack:"d"`
	Card        Card        `msgpack:"e"`
	PropertySet PropertySet `msgpack:"f"`
}

func NewActionPlayShack(seqNum int, playerID uuid.UUID, card Card, propertySet PropertySet) *ActionPlayShack {
	return &ActionPlayShack{
		Kind:        ActionKindPlayShack,
		Version:     1,
		SeqNum:      seqNum,
		PlayerID:    playerID,
		Card:        card,
		PropertySet: propertySet,
	}
}

func (a *ActionPlayShack) GetKind() ActionKind { return a.Kind }
func (a *ActionPlayShack) GetVersion() int     { return a.Version }
func (a *ActionPlayShack) GetSeqNum() int      { return a.SeqNum }

type ActionPlayBigPayday struct {
	Kind           ActionKind `msgpack:"a"`
	Version        int        `msgpack:"b"`
	SeqNum         int        `msgpack:"c"`
	PlayerID       uuid.UUID  `msgpack:"d"`
	LastPlayedCard Card       `msgpack:"e"`
	Cards          Cards      `msgpack:"f"`
}

func NewActionPlayBigPayday(seqNum int, playerID uuid.UUID, lastPlayedCard Card, cards Cards) *ActionPlayBigPayday {
	return &ActionPlayBigPayday{
		Kind:           ActionKindPlayBigPayday,
		Version:        1,
		SeqNum:         seqNum,
		PlayerID:       playerID,
		LastPlayedCard: lastPlayedCard,
		Cards:          cards,
	}
}

func (a *ActionPlayBigPayday) GetKind() ActionKind { return a.Kind }
func (a *ActionPlayBigPayday) GetVersion() int     { return a.Version }
func (a *ActionPlayBigPayday) GetSeqNum() int      { return a.SeqNum }

type ActionPlayGoAgain struct {
	Kind           ActionKind `msgpack:"a"`
	Version        int        `msgpack:"b"`
	SeqNum         int        `msgpack:"c"`
	PlayerID       uuid.UUID  `msgpack:"d"`
	LastPlayedCard Card       `msgpack:"e"`
}

func NewActionPlayGoAgain(seqNum int, playerID uuid.UUID, lastPlayedCard Card) *ActionPlayGoAgain {
	return &ActionPlayGoAgain{
		Kind:           ActionKindPlayGoAgain,
		Version:        1,
		SeqNum:         seqNum,
		PlayerID:       playerID,
		LastPlayedCard: lastPlayedCard,
	}
}

func (a *ActionPlayGoAgain) GetKind() ActionKind { return a.Kind }
func (a *ActionPlayGoAgain) GetVersion() int     { return a.Version }
func (a *ActionPlayGoAgain) GetSeqNum() int      { return a.SeqNum }

type ActionDemandsCreated struct {
	Kind           ActionKind `msgpack:"a"`
	Version        int        `msgpack:"b"`
	SeqNum         int        `msgpack:"c"`
	PlayerID       uuid.UUID  `msgpack:"d"`
	LastPlayedCard *Card      `msgpack:"e"`
	DeniedDemand   *Demand    `msgpack:"g"`
	Demands        []Demand   `msgpack:"f"`
}

func NewActionDemandsCreated(seqNum int, playerID uuid.UUID, lastPlayedCard *Card, deniedDemand *Demand, demands ...Demand) *ActionDemandsCreated {
	return &ActionDemandsCreated{
		Kind:           ActionKindDemandCreated,
		Version:        1,
		SeqNum:         seqNum,
		PlayerID:       playerID,
		LastPlayedCard: lastPlayedCard,
		DeniedDemand:   deniedDemand,
		Demands:        demands,
	}
}

func (a *ActionDemandsCreated) GetKind() ActionKind { return a.Kind }
func (a *ActionDemandsCreated) GetVersion() int     { return a.Version }
func (a *ActionDemandsCreated) GetSeqNum() int      { return a.SeqNum }

// TransferCards records a payment: cards moved from SourceID to TargetID,
// where money/action cards land in the bank and property cards form new
// sets. Debt is non-nil when the payer could not cover the amount and a
// debt chip was issued to the creditor.
type TransferCards struct {
	SourceID     uuid.UUID       `msgpack:"a"`
	TargetID     uuid.UUID       `msgpack:"b"`
	Cards        Cards           `msgpack:"c"`
	PropertySets PropertySets    `msgpack:"d"`
	SourceSets   int             `msgpack:"e"`
	SourceMoney  int             `msgpack:"f"`
	TargetSets   int             `msgpack:"g"`
	TargetMoney  int             `msgpack:"h"`
	Debt         *DebtObligation `msgpack:"i"`
}

// TransferProperty records a single property card moved from SourceID to
// TargetID (market crash); PropertySet is the resulting set on the receiver.
type TransferProperty struct {
	SourceID    uuid.UUID   `msgpack:"a"`
	TargetID    uuid.UUID   `msgpack:"b"`
	Card        Card        `msgpack:"c"`
	PropertySet PropertySet `msgpack:"d"`
}

// TransferPropertySets records whole sets (shacks included) moved from
// SourceID to TargetID (set snatcher: one set; property raid: all sets of a
// color).
type TransferPropertySets struct {
	SourceID     uuid.UUID    `msgpack:"a"`
	TargetID     uuid.UUID    `msgpack:"b"`
	PropertySets PropertySets `msgpack:"c"`
}

// TransferBankSwap records the banks after the swap: SourceBank is what
// SourceID (the demand target / payer) now holds.
type TransferBankSwap struct {
	SourceID   uuid.UUID `msgpack:"a"`
	TargetID   uuid.UUID `msgpack:"b"`
	SourceBank Cards     `msgpack:"c"`
	TargetBank Cards     `msgpack:"d"`
}

// DistributionEntry is one card handed to RecipientID during a repo man /
// tax day distribution. PropertySet is set when the card landed as a
// property; otherwise the card went to the recipient's bank.
type DistributionEntry struct {
	RecipientID uuid.UUID    `msgpack:"a"`
	Card        Card         `msgpack:"b"`
	PropertySet *PropertySet `msgpack:"c"`
}

// TransferDistribution records a repo man / tax day compliance: SourceID
// kept KeptCard and distributed the rest.
type TransferDistribution struct {
	SourceID uuid.UUID           `msgpack:"a"`
	KeptCard Card                `msgpack:"b"`
	Entries  []DistributionEntry `msgpack:"c"`
}

// TransferPickpocket records hand cards of Category moved from SourceID's
// hand to TargetID's hand.
type TransferPickpocket struct {
	SourceID uuid.UUID     `msgpack:"a"`
	TargetID uuid.UUID     `msgpack:"b"`
	Category StealCategory `msgpack:"c"`
	Cards    Cards         `msgpack:"d"`
}

// TransferDebtChip records a debt trap compliance: a chip moved from the
// target's pool into an outstanding obligation.
type TransferDebtChip struct {
	Debt DebtObligation `msgpack:"a"`
}

type ActionDemandComplied struct {
	Kind                 ActionKind            `msgpack:"a"`
	Version              int                   `msgpack:"b"`
	SeqNum               int                   `msgpack:"c"`
	PlayerID             uuid.UUID             `msgpack:"d"`
	DemandID             Identifier            `msgpack:"e"`
	TransferCards        *TransferCards        `msgpack:"f"`
	TransferProperty     *TransferProperty     `msgpack:"g"`
	TransferPropertySets *TransferPropertySets `msgpack:"h"`
	TransferBankSwap     *TransferBankSwap     `msgpack:"i"`
	TransferDistribution *TransferDistribution `msgpack:"j"`
	TransferPickpocket   *TransferPickpocket   `msgpack:"k"`
	TransferDebtChip     *TransferDebtChip     `msgpack:"l"`
}

func NewActionDemandComplied(seqNum int, playerID uuid.UUID, demandID Identifier) *ActionDemandComplied {
	return &ActionDemandComplied{
		Kind:     ActionKindDemandComplied,
		Version:  1,
		SeqNum:   seqNum,
		PlayerID: playerID,
		DemandID: demandID,
	}
}

func (a *ActionDemandComplied) GetKind() ActionKind { return a.Kind }
func (a *ActionDemandComplied) GetVersion() int     { return a.Version }
func (a *ActionDemandComplied) GetSeqNum() int      { return a.SeqNum }

// ActionDebtSettled records a mandatory debt settlement: the debtor gave the
// creditor Card, which the creditor received as if they had played it
// themselves (property → PropertySet, anything else → bank).
type ActionDebtSettled struct {
	Kind        ActionKind     `msgpack:"a"`
	Version     int            `msgpack:"b"`
	SeqNum      int            `msgpack:"c"`
	PlayerID    uuid.UUID      `msgpack:"d"`
	Debt        DebtObligation `msgpack:"e"`
	Card        Card           `msgpack:"f"`
	PropertySet *PropertySet   `msgpack:"g"`
	Forgiven    bool           `msgpack:"h"`
}

func NewActionDebtSettled(seqNum int, playerID uuid.UUID, debt DebtObligation, card Card, propertySet *PropertySet, forgiven bool) *ActionDebtSettled {
	return &ActionDebtSettled{
		Kind:        ActionKindDebtSettled,
		Version:     1,
		SeqNum:      seqNum,
		PlayerID:    playerID,
		Debt:        debt,
		Card:        card,
		PropertySet: propertySet,
		Forgiven:    forgiven,
	}
}

func (a *ActionDebtSettled) GetKind() ActionKind { return a.Kind }
func (a *ActionDebtSettled) GetVersion() int     { return a.Version }
func (a *ActionDebtSettled) GetSeqNum() int      { return a.SeqNum }

type ActionDiscardCards struct {
	Kind     ActionKind `msgpack:"a"`
	Version  int        `msgpack:"b"`
	SeqNum   int        `msgpack:"c"`
	PlayerID uuid.UUID  `msgpack:"d"`
	Cards    Cards      `msgpack:"e"`
}

func NewActionDiscardCards(seqNum int, playerID uuid.UUID, cards ...Card) *ActionDiscardCards {
	return &ActionDiscardCards{
		Kind:     ActionKindDiscardCards,
		Version:  1,
		SeqNum:   seqNum,
		PlayerID: playerID,
		Cards:    cards,
	}
}

func (a *ActionDiscardCards) GetKind() ActionKind { return a.Kind }
func (a *ActionDiscardCards) GetVersion() int     { return a.Version }
func (a *ActionDiscardCards) GetSeqNum() int      { return a.SeqNum }

type ActionStartTurn struct {
	Kind      ActionKind       `msgpack:"a"`
	Version   int              `msgpack:"b"`
	SeqNum    int              `msgpack:"c"`
	PlayerID  uuid.UUID        `msgpack:"d"`
	Cards     Cards            `msgpack:"e"`
	MovesLeft int              `msgpack:"f"`
	Debts     []DebtObligation `msgpack:"g"`
	GoAgain   bool             `msgpack:"h"`
}

func NewActionStartTurn(seqNum int, playerID uuid.UUID, cards Cards, movesLeft int, debts []DebtObligation, goAgain bool) *ActionStartTurn {
	return &ActionStartTurn{
		Kind:      ActionKindStartTurn,
		Version:   1,
		SeqNum:    seqNum,
		PlayerID:  playerID,
		Cards:     cards,
		MovesLeft: movesLeft,
		Debts:     debts,
		GoAgain:   goAgain,
	}
}

func (a *ActionStartTurn) GetKind() ActionKind { return a.Kind }
func (a *ActionStartTurn) GetVersion() int     { return a.Version }
func (a *ActionStartTurn) GetSeqNum() int      { return a.SeqNum }

type ActionRearrangeCard struct {
	Kind        ActionKind  `msgpack:"a"`
	Version     int         `msgpack:"b"`
	SeqNum      int         `msgpack:"c"`
	PlayerID    uuid.UUID   `msgpack:"d"`
	Card        *Card       `msgpack:"e"`
	PropertySet PropertySet `msgpack:"f"`
	Sets        int         `msgpack:"g"`
}

func NewActionRearrangeCard(seqNum int, playerID uuid.UUID, card *Card, propertySet PropertySet, numSets int) *ActionRearrangeCard {
	return &ActionRearrangeCard{
		Kind:        ActionKindRearrangeCard,
		Version:     1,
		SeqNum:      seqNum,
		PlayerID:    playerID,
		Card:        card,
		PropertySet: propertySet,
		Sets:        numSets,
	}
}

func (a *ActionRearrangeCard) GetKind() ActionKind { return a.Kind }
func (a *ActionRearrangeCard) GetVersion() int     { return a.Version }
func (a *ActionRearrangeCard) GetSeqNum() int      { return a.SeqNum }
