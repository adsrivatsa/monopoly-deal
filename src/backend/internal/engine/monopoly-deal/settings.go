package monopoly_deal

import (
	"time"

	"github.com/vmihailenco/msgpack/v5"
)

type Speed int

const (
	SpeedUnidentified Speed = iota
	SpeedSlow
	SpeedMedium
	SpeedFast
)

type Settings struct {
	NumDecks int `msgpack:"num_decks" validate:"required,min=1,max=3"`

	// deal rules
	StartNumCards int `msgpack:"start_num_cards" validate:"required,min=5,max=8"`
	MaxHandSize   int `msgpack:"max_hand_size" validate:"required,min=5,max=10"`
	MovesPerTurn  int `msgpack:"moves_per_turn" validate:"required,min=3,max=5"`

	// card rules
	PaydayDraw       int `msgpack:"payday_draw" validate:"required,min=2,max=5"`
	PartyBillPayment int `msgpack:"party_bill_payment" validate:"required,min=2,max=5"`
	SettleUpPayment  int `msgpack:"settle_up_payment" validate:"required,min=5,max=8"`

	// win conditions
	WinSetAmount   int `msgpack:"win_set_amount" validate:"required,min=3,max=6"`
	WinMoneyAmount int `msgpack:"win_money_amount" validate:"min=0,max=40"`

	// game speed
	Speed         Speed         `msgpack:"speed" validate:"required,min=1,max=3"`
	MoveTimeout   time.Duration `msgpack:"move_timeout"`
	DemandTimeout time.Duration `msgpack:"demand_timeout"`

	// deck rules
	PaydayAmount        int `msgpack:"payday_amount" validate:"required,min=10,max=20"`
	DoubleTheRentAmount int `msgpack:"double_the_rent_amount" validate:"required,min=2,max=4"`
	PartyBillAmount     int `msgpack:"party_bill_amount" validate:"required,min=3,max=6"`
	HouseAmount         int `msgpack:"house_amount" validate:"required,min=3,max=6"`
	PropertyStealAmount int `msgpack:"property_steal_amount" validate:"required,min=3,max=6"`
	PropertySwapAmount  int `msgpack:"property_swap_amount" validate:"required,min=3,max=6"`
	SettleUpAmount      int `msgpack:"settle_up_amount" validate:"required,min=3,max=6"`
	HotelAmount         int `msgpack:"hotel_amount" validate:"required,min=2,max=4"`
	NahAmount           int `msgpack:"nah_amount" validate:"required,min=3,max=6"`
	SetSnatcherAmount   int `msgpack:"set_snatcher_amount" validate:"required,min=2,max=4"`
	RentAmount          int `msgpack:"rent_amount" validate:"required,min=2,max=4"`
	WildRentAmount      int `msgpack:"wild_rent_amount" validate:"required,min=3,max=6"`
}

func DefaultSettings() Settings {
	return Settings{
		NumDecks:            1,
		StartNumCards:       5,
		MaxHandSize:         7,
		MovesPerTurn:        3,
		PaydayDraw:          2,
		PartyBillPayment:    2,
		SettleUpPayment:     5,
		WinSetAmount:        3,
		WinMoneyAmount:      0,
		Speed:               SpeedMedium,
		PaydayAmount:        10,
		DoubleTheRentAmount: 2,
		PartyBillAmount:     3,
		HouseAmount:         3,
		PropertyStealAmount: 3,
		PropertySwapAmount:  3,
		SettleUpAmount:      3,
		HotelAmount:         2,
		NahAmount:           3,
		SetSnatcherAmount:   2,
		RentAmount:          2,
		WildRentAmount:      3,
	}
}

func (s *Settings) Encode() ([]byte, error) {
	if s == nil {
		return nil, nil
	}
	return msgpack.Marshal(*s)
}

func (s *Settings) Decode(data []byte) error {
	*s = DefaultSettings()
	err := msgpack.Unmarshal(data, s)
	if err != nil {
		return err
	}

	switch s.Speed {
	case SpeedFast:
		s.MoveTimeout = time.Second * 15
	case SpeedMedium:
		s.MoveTimeout = time.Second * 30
	default:
		s.MoveTimeout = time.Second * 45
	}
	s.DemandTimeout = s.MoveTimeout * 2 / 3

	return nil
}
