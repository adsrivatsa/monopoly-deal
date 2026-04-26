package monopoly_deal

import (
	"the-deal/internal/schema/monopoly_deal_schema"

	"github.com/google/uuid"
)

type ActionKind string

const (
	ActionKindUnspecified        ActionKind = "unspecified"
	ActionKindPlayMoney          ActionKind = "play_money"
	ActionKindPlayProperty       ActionKind = "play_property"
	ActionKindPlayHouse          ActionKind = "play_house"
	ActionKindPlayHotel          ActionKind = "play_hotel"
	ActionKindPlayPassGo         ActionKind = "play_pass_go"
	ActionKindDemandCreated      ActionKind = "demand_created"
	ActionKindPendingRentCreated ActionKind = "pending_rent_created"
	ActionKindDemandComplied     ActionKind = "demand_complied"
	ActionKindDiscardCards       ActionKind = "discard_cards"
	ActionKindStartTurn          ActionKind = "start_turn"
	ActionKindRearrangeCard      ActionKind = "rearrange_card"
)

var ActionKindProtoMap = map[ActionKind]monopoly_deal_schema.ActionKind{
	ActionKindUnspecified:        monopoly_deal_schema.ActionKind_ACTION_KIND_UNSPECIFIED,
	ActionKindPlayMoney:          monopoly_deal_schema.ActionKind_ACTION_KIND_PLAY_MONEY,
	ActionKindPlayProperty:       monopoly_deal_schema.ActionKind_ACTION_KIND_PLAY_PROPERTY,
	ActionKindPlayHouse:          monopoly_deal_schema.ActionKind_ACTION_KIND_PLAY_HOUSE,
	ActionKindPlayHotel:          monopoly_deal_schema.ActionKind_ACTION_KIND_PLAY_HOTEL,
	ActionKindPlayPassGo:         monopoly_deal_schema.ActionKind_ACTION_KIND_PLAY_PASS_GO,
	ActionKindDemandCreated:      monopoly_deal_schema.ActionKind_ACTION_KIND_DEMANDS_CREATED,
	ActionKindPendingRentCreated: monopoly_deal_schema.ActionKind_ACTION_KIND_PENDING_RENT_CREATED,
	ActionKindDemandComplied:     monopoly_deal_schema.ActionKind_ACTION_KIND_DEMAND_COMPLIED,
	ActionKindDiscardCards:       monopoly_deal_schema.ActionKind_ACTION_KIND_DISCARD_CARDS,
	ActionKindStartTurn:          monopoly_deal_schema.ActionKind_ACTION_KIND_START_TURN,
	ActionKindRearrangeCard:      monopoly_deal_schema.ActionKind_ACTION_KIND_REARRANGE_CARD,
}

func (a ActionKind) Proto() monopoly_deal_schema.ActionKind {
	return ActionKindProtoMap[a]
}

type Action interface {
	GetKind() ActionKind
	GetVersion() int
	GetSeqNum() int
	Proto() *monopoly_deal_schema.Action
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

func (a *ActionPlayMoney) GetKind() ActionKind {
	return a.Kind
}

func (a *ActionPlayMoney) GetVersion() int {
	return a.Version
}

func (a *ActionPlayMoney) GetSeqNum() int {
	return a.SeqNum
}

func (a *ActionPlayMoney) Proto() *monopoly_deal_schema.Action {
	return &monopoly_deal_schema.Action{
		PlayerId:       a.PlayerID.String(),
		Kind:           a.Kind.Proto(),
		SeqNum:         int32(a.SeqNum),
		TurnDeadlineMs: 0, // Added by caller
		Payload: &monopoly_deal_schema.Action_ActionPlayMoney{
			ActionPlayMoney: &monopoly_deal_schema.ActionPlayMoney{
				Card: a.Card.Proto(),
			},
		},
	}
}

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

func (a *ActionPlayProperty) GetKind() ActionKind {
	return a.Kind
}

func (a *ActionPlayProperty) GetVersion() int {
	return a.Version
}

func (a *ActionPlayProperty) GetSeqNum() int {
	return a.SeqNum
}

func (a *ActionPlayProperty) Proto() *monopoly_deal_schema.Action {
	return &monopoly_deal_schema.Action{
		PlayerId:       a.PlayerID.String(),
		Kind:           a.Kind.Proto(),
		SeqNum:         int32(a.SeqNum),
		TurnDeadlineMs: 0, // Added by caller
		Payload: &monopoly_deal_schema.Action_ActionPlayProperty{
			ActionPlayProperty: &monopoly_deal_schema.ActionPlayProperty{
				PropertySet: a.PropertySet.Proto(a.PlayerID),
			},
		},
	}
}

type ActionPlayHouse struct {
	Kind        ActionKind  `msgpack:"a"`
	Version     int         `msgpack:"b"`
	SeqNum      int         `msgpack:"c"`
	PlayerID    uuid.UUID   `msgpack:"d"`
	Card        Card        `msgpack:"e"`
	PropertySet PropertySet `msgpack:"f"`
}

func NewActionPlayHouse(seqNum int, playerID uuid.UUID, card Card, propertySet PropertySet) *ActionPlayHouse {
	return &ActionPlayHouse{
		Kind:        ActionKindPlayHouse,
		Version:     1,
		SeqNum:      seqNum,
		PlayerID:    playerID,
		Card:        card,
		PropertySet: propertySet,
	}
}

func (a *ActionPlayHouse) GetKind() ActionKind {
	return a.Kind
}

func (a *ActionPlayHouse) GetVersion() int {
	return a.Version
}

func (a *ActionPlayHouse) GetSeqNum() int {
	return a.SeqNum
}

func (a *ActionPlayHouse) Proto() *monopoly_deal_schema.Action {
	return &monopoly_deal_schema.Action{
		PlayerId:       a.PlayerID.String(),
		Kind:           a.Kind.Proto(),
		SeqNum:         int32(a.SeqNum),
		TurnDeadlineMs: 0, // Added by caller
		Payload: &monopoly_deal_schema.Action_ActionPlayHouse{
			ActionPlayHouse: &monopoly_deal_schema.ActionPlayHouse{
				Card:        a.Card.Proto(),
				PropertySet: a.PropertySet.Proto(a.PlayerID),
			},
		},
	}
}

type ActionPlayHotel struct {
	Kind        ActionKind  `msgpack:"a"`
	Version     int         `msgpack:"b"`
	SeqNum      int         `msgpack:"c"`
	PlayerID    uuid.UUID   `msgpack:"d"`
	Card        Card        `msgpack:"e"`
	PropertySet PropertySet `msgpack:"f"`
}

func NewActionPlayHotel(seqNum int, playerID uuid.UUID, card Card, propertySet PropertySet) *ActionPlayHotel {
	return &ActionPlayHotel{
		Kind:        ActionKindPlayHotel,
		Version:     1,
		SeqNum:      seqNum,
		PlayerID:    playerID,
		Card:        card,
		PropertySet: propertySet,
	}
}

func (a *ActionPlayHotel) GetKind() ActionKind {
	return a.Kind
}

func (a *ActionPlayHotel) GetVersion() int {
	return a.Version
}

func (a *ActionPlayHotel) GetSeqNum() int {
	return a.SeqNum
}

func (a *ActionPlayHotel) Proto() *monopoly_deal_schema.Action {
	return &monopoly_deal_schema.Action{
		PlayerId:       a.PlayerID.String(),
		Kind:           a.Kind.Proto(),
		SeqNum:         int32(a.SeqNum),
		TurnDeadlineMs: 0, // Added by caller
		Payload: &monopoly_deal_schema.Action_ActionPlayHotel{
			ActionPlayHotel: &monopoly_deal_schema.ActionPlayHotel{
				Card:        a.Card.Proto(),
				PropertySet: a.PropertySet.Proto(a.PlayerID),
			},
		},
	}
}

type ActionPlayPassGo struct {
	Kind           ActionKind `msgpack:"a"`
	Version        int        `msgpack:"b"`
	SeqNum         int        `msgpack:"c"`
	PlayerID       uuid.UUID  `msgpack:"d"`
	LastPlayedCard Card       `msgpack:"e"`
	Cards          Cards      `msgpack:"f"`
}

func NewActionPlayPassGo(seqNum int, playerID uuid.UUID, lastPlayedCard Card, cards Cards) *ActionPlayPassGo {
	return &ActionPlayPassGo{
		Kind:           ActionKindPlayPassGo,
		Version:        1,
		SeqNum:         seqNum,
		PlayerID:       playerID,
		LastPlayedCard: lastPlayedCard,
		Cards:          cards,
	}
}

func (a *ActionPlayPassGo) GetKind() ActionKind {
	return a.Kind
}

func (a *ActionPlayPassGo) GetVersion() int {
	return a.Version
}

func (a *ActionPlayPassGo) GetSeqNum() int {
	return a.SeqNum
}

func (a *ActionPlayPassGo) Proto() *monopoly_deal_schema.Action {
	return &monopoly_deal_schema.Action{
		PlayerId:       a.PlayerID.String(),
		Kind:           a.Kind.Proto(),
		SeqNum:         int32(a.SeqNum),
		TurnDeadlineMs: 0, // Added by caller
		Payload: &monopoly_deal_schema.Action_ActionPlayPassGo{
			ActionPlayPassGo: &monopoly_deal_schema.ActionPlayPassGo{
				LastPlayedCard: a.LastPlayedCard.Proto(),
				Cards:          a.Cards.Proto(),
			},
		},
	}
}

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

func (a *ActionDemandsCreated) GetKind() ActionKind {
	return a.Kind
}

func (a *ActionDemandsCreated) GetVersion() int {
	return a.Version
}

func (a *ActionDemandsCreated) GetSeqNum() int {
	return a.SeqNum
}

func (a *ActionDemandsCreated) Proto() *monopoly_deal_schema.Action {
	lastPlayedCard := new(monopoly_deal_schema.Card)
	if a.LastPlayedCard != nil {
		lastPlayedCard = a.LastPlayedCard.Proto()
	}

	demands := make([]*monopoly_deal_schema.Demand, len(a.Demands))
	for i, demand := range a.Demands {
		demands[i] = demand.Proto()
	}

	var deniedDemand *monopoly_deal_schema.Demand
	if a.DeniedDemand != nil {
		deniedDemand = a.DeniedDemand.Proto()
	}

	return &monopoly_deal_schema.Action{
		PlayerId:       a.PlayerID.String(),
		Kind:           a.Kind.Proto(),
		SeqNum:         int32(a.SeqNum),
		TurnDeadlineMs: 0, // Added by caller
		Payload: &monopoly_deal_schema.Action_ActionDemandsCreated{
			ActionDemandsCreated: &monopoly_deal_schema.ActionDemandsCreated{
				LastPlayedCard: lastPlayedCard,
				DeniedDemand:   deniedDemand,
				Demands:        demands,
			},
		},
	}
}

type ActionPendingRentCreated struct {
	Kind           ActionKind  `msgpack:"a"`
	Version        int         `msgpack:"b"`
	SeqNum         int         `msgpack:"c"`
	PlayerID       uuid.UUID   `msgpack:"d"`
	LastPlayedCard Card        `msgpack:"e"`
	PendingRent    PendingRent `msgpack:"f"`
}

func NewActionPendingRentCreated(seqNum int, playerID uuid.UUID, lastPlayerCard Card, pendingRent PendingRent) *ActionPendingRentCreated {
	return &ActionPendingRentCreated{
		Kind:           ActionKindPendingRentCreated,
		Version:        1,
		SeqNum:         seqNum,
		PlayerID:       playerID,
		LastPlayedCard: lastPlayerCard,
		PendingRent:    pendingRent,
	}
}

func (a *ActionPendingRentCreated) GetKind() ActionKind {
	return a.Kind
}

func (a *ActionPendingRentCreated) GetVersion() int {
	return a.Version
}

func (a *ActionPendingRentCreated) GetSeqNum() int {
	return a.SeqNum
}

func (a *ActionPendingRentCreated) Proto() *monopoly_deal_schema.Action {
	return &monopoly_deal_schema.Action{
		PlayerId:       a.PlayerID.String(),
		Kind:           a.Kind.Proto(),
		SeqNum:         int32(a.SeqNum),
		TurnDeadlineMs: 0, // Added by caller
		Payload: &monopoly_deal_schema.Action_ActionPendingRentCreated{
			ActionPendingRentCreated: &monopoly_deal_schema.ActionPendingRentCreated{
				LastCardPlayed: a.LastPlayedCard.Proto(),
				PendingRent:    a.PendingRent.Proto(),
			},
		},
	}
}

type TransferCards struct {
	SourceID     uuid.UUID    `msgpack:"a"`
	TargetID     uuid.UUID    `msgpack:"b"`
	Cards        Cards        `msgpack:"c"`
	PropertySets PropertySets `msgpack:"d"`
	SourceSets   int          `msgpack:"e"`
	SourceMoney  int          `msgpack:"f"`
	TargetSets   int          `msgpack:"g"`
	TargetMoney  int          `msgpack:"h"`
}

type TransferProperty struct {
	SourceID           uuid.UUID    `msgpack:"a"`
	TargetID           uuid.UUID    `msgpack:"b"`
	SourcePropertySets PropertySets `msgpack:"c"`
	TargetPropertySets PropertySets `msgpack:"d"`
}

type TransferPropertySet struct {
	SourceID    uuid.UUID   `msgpack:"a"`
	TargetID    uuid.UUID   `msgpack:"b"`
	PropertySet PropertySet `msgpack:"c"`
}

type ActionDemandComplied struct {
	Kind                ActionKind           `msgpack:"a"`
	Version             int                  `msgpack:"b"`
	SeqNum              int                  `msgpack:"c"`
	PlayerID            uuid.UUID            `msgpack:"d"`
	DemandID            Identifier           `msgpack:"e"`
	TransferCards       *TransferCards       `msgpack:"f"`
	TransferProperty    *TransferProperty    `msgpack:"g"`
	TransferPropertySet *TransferPropertySet `msgpack:"h"`
}

func NewActionDemandComplied(seqNum int, playerID uuid.UUID, demandID Identifier, transferCards *TransferCards, transferProperty *TransferProperty, transferPropertySet *TransferPropertySet) *ActionDemandComplied {
	return &ActionDemandComplied{
		Kind:                ActionKindDemandComplied,
		Version:             1,
		SeqNum:              seqNum,
		PlayerID:            playerID,
		DemandID:            demandID,
		TransferCards:       transferCards,
		TransferProperty:    transferProperty,
		TransferPropertySet: transferPropertySet,
	}
}

func (a *ActionDemandComplied) GetKind() ActionKind {
	return a.Kind
}

func (a *ActionDemandComplied) GetVersion() int {
	return a.Version
}

func (a *ActionDemandComplied) GetSeqNum() int {
	return a.SeqNum
}

func (a *ActionDemandComplied) Proto() *monopoly_deal_schema.Action {
	var demandComplied *monopoly_deal_schema.ActionDemandComplied

	if a.TransferCards != nil {
		demandComplied = &monopoly_deal_schema.ActionDemandComplied{
			DemandId: string(a.DemandID),
			Transfer: &monopoly_deal_schema.ActionDemandComplied_TransferCards{
				TransferCards: &monopoly_deal_schema.TransferCards{
					SourceId:     a.TransferCards.SourceID.String(),
					TargetId:     a.TransferCards.TargetID.String(),
					Cards:        a.TransferCards.Cards.Proto(),
					PropertySets: a.TransferCards.PropertySets.Proto(a.TransferCards.TargetID),
					SourceSets:   int32(a.TransferCards.SourceSets),
					SourceMoney:  int32(a.TransferCards.SourceMoney),
					TargetSets:   int32(a.TransferCards.TargetSets),
					TargetMoney:  int32(a.TransferCards.TargetMoney),
				},
			},
		}
	}

	if a.TransferProperty != nil {
		demandComplied = &monopoly_deal_schema.ActionDemandComplied{
			DemandId: string(a.DemandID),
			Transfer: &monopoly_deal_schema.ActionDemandComplied_TransferProperty{
				TransferProperty: &monopoly_deal_schema.TransferProperty{
					SourceId:           a.TransferProperty.SourceID.String(),
					TargetId:           a.TransferProperty.TargetID.String(),
					SourcePropertySets: a.TransferProperty.SourcePropertySets.Proto(a.TransferProperty.SourceID),
					TargetPropertySets: a.TransferProperty.TargetPropertySets.Proto(a.TransferProperty.TargetID),
				},
			},
		}
	}

	if a.TransferPropertySet != nil {
		demandComplied = &monopoly_deal_schema.ActionDemandComplied{
			DemandId: string(a.DemandID),
			Transfer: &monopoly_deal_schema.ActionDemandComplied_TransferPropertySet{
				TransferPropertySet: &monopoly_deal_schema.TransferPropertySet{
					SourceId:    a.TransferPropertySet.SourceID.String(),
					TargetId:    a.TransferPropertySet.TargetID.String(),
					PropertySet: a.TransferPropertySet.PropertySet.Proto(a.TransferPropertySet.TargetID),
				},
			},
		}
	}

	return &monopoly_deal_schema.Action{
		PlayerId:       a.PlayerID.String(),
		Kind:           a.Kind.Proto(),
		SeqNum:         int32(a.SeqNum),
		TurnDeadlineMs: 0, // Added by caller
		Payload: &monopoly_deal_schema.Action_ActionDemandComplied{
			ActionDemandComplied: demandComplied,
		},
	}
}

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

func (a *ActionDiscardCards) GetKind() ActionKind {
	return a.Kind
}

func (a *ActionDiscardCards) GetVersion() int {
	return a.Version
}

func (a *ActionDiscardCards) GetSeqNum() int {
	return a.SeqNum
}

func (a *ActionDiscardCards) Proto() *monopoly_deal_schema.Action {
	return &monopoly_deal_schema.Action{
		PlayerId:       a.PlayerID.String(),
		Kind:           a.Kind.Proto(),
		SeqNum:         int32(a.SeqNum),
		TurnDeadlineMs: 0, // Added by caller
		Payload: &monopoly_deal_schema.Action_ActionDiscardCards{
			ActionDiscardCards: &monopoly_deal_schema.ActionDiscardCards{
				Cards: a.Cards.Proto(),
			},
		},
	}
}

type ActionStartTurn struct {
	Kind      ActionKind `msgpack:"a"`
	Version   int        `msgpack:"b"`
	SeqNum    int        `msgpack:"c"`
	PlayerID  uuid.UUID  `msgpack:"d"`
	Cards     Cards      `msgpack:"e"`
	MovesLeft int        `msgpack:"f"`
}

func NewActionStartTurn(seqNum int, playerID uuid.UUID, cards Cards, movesLeft int) *ActionStartTurn {
	return &ActionStartTurn{
		Kind:      ActionKindStartTurn,
		Version:   1,
		SeqNum:    seqNum,
		PlayerID:  playerID,
		Cards:     cards,
		MovesLeft: movesLeft,
	}
}

func (a *ActionStartTurn) GetKind() ActionKind {
	return a.Kind
}

func (a *ActionStartTurn) GetVersion() int {
	return a.Version
}

func (a *ActionStartTurn) GetSeqNum() int {
	return a.SeqNum
}

func (a *ActionStartTurn) Proto() *monopoly_deal_schema.Action {
	return &monopoly_deal_schema.Action{
		PlayerId:       a.PlayerID.String(),
		Kind:           a.Kind.Proto(),
		SeqNum:         int32(a.SeqNum),
		TurnDeadlineMs: 0, // Added by caller
		Payload: &monopoly_deal_schema.Action_ActionStartTurn{
			ActionStartTurn: &monopoly_deal_schema.ActionStartTurn{
				Cards:     a.Cards.Proto(),
				MovesLeft: int32(a.MovesLeft),
			},
		},
	}
}

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

func (a *ActionRearrangeCard) GetKind() ActionKind {
	return a.Kind
}

func (a *ActionRearrangeCard) GetVersion() int {
	return a.Version
}

func (a *ActionRearrangeCard) GetSeqNum() int {
	return a.SeqNum
}

func (a *ActionRearrangeCard) Proto() *monopoly_deal_schema.Action {
	var card *monopoly_deal_schema.Card
	if a.Card != nil {
		card = a.Card.Proto()
	}

	return &monopoly_deal_schema.Action{
		PlayerId:       a.PlayerID.String(),
		Kind:           a.Kind.Proto(),
		SeqNum:         int32(a.SeqNum),
		TurnDeadlineMs: 0,
		Payload: &monopoly_deal_schema.Action_ActionRearrangeCard{
			ActionRearrangeCard: &monopoly_deal_schema.ActionRearrangeCard{
				Card:        card,
				PropertySet: a.PropertySet.Proto(a.PlayerID),
				Sets:        int32(a.Sets),
			},
		},
	}
}
