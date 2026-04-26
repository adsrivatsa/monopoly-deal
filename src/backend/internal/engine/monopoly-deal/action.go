package monopoly_deal

import (
	"the-deal/internal/schema/monopoly_deal_schema"

	"github.com/google/uuid"
)

type ActionKind string

const (
	ActionKindUnspecified  ActionKind = "unspecified"
	ActionKindPlayMoney    ActionKind = "play_money"
	ActionKindPlayProperty ActionKind = "play_property"
	ActionKindPlayHouse    ActionKind = "play_house"
	ActionKindPlayHotel    ActionKind = "play_hotel"
	ActionKindPlayPassGo   ActionKind = "play_pass_go"
)

var ActionKindProtoMap = map[ActionKind]monopoly_deal_schema.ActionKind{
	ActionKindUnspecified:  monopoly_deal_schema.ActionKind_ACTION_KIND_UNSPECIFIED,
	ActionKindPlayMoney:    monopoly_deal_schema.ActionKind_ACTION_KIND_PLAY_MONEY,
	ActionKindPlayProperty: monopoly_deal_schema.ActionKind_ACTION_KIND_PLAY_PROPERTY,
	ActionKindPlayHouse:    monopoly_deal_schema.ActionKind_ACTION_KIND_PLAY_HOUSE,
	ActionKindPlayHotel:    monopoly_deal_schema.ActionKind_ACTION_KIND_PLAY_HOTEL,
	ActionKindPlayPassGo:   monopoly_deal_schema.ActionKind_ACTION_KIND_PLAY_PASS_GO,
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
	Cards          Cards      `msgpack:"e"`
	LastPlayedCard Card       `msgpack:"f"`
}

func NewActionPlayPassGo(seqNum int, playerID uuid.UUID, cards Cards, lastPlayedCard Card) *ActionPlayPassGo {
	return &ActionPlayPassGo{
		Kind:           ActionKindPlayPassGo,
		Version:        1,
		SeqNum:         seqNum,
		PlayerID:       playerID,
		Cards:          cards,
		LastPlayedCard: lastPlayedCard,
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
