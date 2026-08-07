package deal_no_mercy

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
	TurnDraw      int `msgpack:"turn_draw" validate:"required,min=2,max=5"`

	// win conditions
	WinSetAmount   int `msgpack:"win_set_amount" validate:"required,min=3,max=6"`
	WinMoneyAmount int `msgpack:"win_money_amount" validate:"min=0,max=40"`

	// no mercy rules
	DebtChipsPerPlayer  int `msgpack:"debt_chips_per_player" validate:"required,min=1,max=5"`
	YoinkPayment        int `msgpack:"yoink_payment" validate:"required,min=5,max=15"`
	ShackRentBonus      int `msgpack:"shack_rent_bonus" validate:"required,min=3,max=8"`
	BigPaydayHandTarget int `msgpack:"big_payday_hand_target" validate:"required,min=5,max=10"`

	// game speed
	Speed         Speed         `msgpack:"speed" validate:"required,min=1,max=3"`
	MoveTimeout   time.Duration `msgpack:"move_timeout"`
	DemandTimeout time.Duration `msgpack:"demand_timeout"`

	// deck rules
	SetSnatcherAmount    int `msgpack:"set_snatcher_amount" validate:"required,min=2,max=5"`
	DebtTrapAmount       int `msgpack:"debt_trap_amount" validate:"required,min=2,max=5"`
	GoAgainAmount        int `msgpack:"go_again_amount" validate:"required,min=2,max=5"`
	HeistAmount          int `msgpack:"heist_amount" validate:"required,min=2,max=5"`
	MarketCrashAmount    int `msgpack:"market_crash_amount" validate:"required,min=2,max=5"`
	BigPaydayAmount      int `msgpack:"big_payday_amount" validate:"required,min=2,max=6"`
	RepoManAmount        int `msgpack:"repo_man_amount" validate:"required,min=1,max=4"`
	ShackAmount          int `msgpack:"shack_amount" validate:"required,min=2,max=5"`
	PropertyRaidAmount   int `msgpack:"property_raid_amount" validate:"required,min=2,max=5"`
	TaxDayAmount         int `msgpack:"tax_day_amount" validate:"required,min=1,max=4"`
	PickpocketAmount     int `msgpack:"pickpocket_amount" validate:"required,min=2,max=5"`
	BankSwapAmount       int `msgpack:"bank_swap_amount" validate:"required,min=1,max=4"`
	YoinkAmount          int `msgpack:"yoink_amount" validate:"required,min=2,max=5"`
	NahAmount            int `msgpack:"nah_amount" validate:"required,min=3,max=7"`
	RentAmount           int `msgpack:"rent_amount" validate:"required,min=1,max=3"`
	DoubleRentAmount     int `msgpack:"double_rent_amount" validate:"required,min=1,max=3"`
	WildRentAmount       int `msgpack:"wild_rent_amount" validate:"required,min=2,max=5"`
	DoubleRentWildAmount int `msgpack:"double_rent_wild_amount" validate:"required,min=2,max=5"`

	// nah! rules
	NahConsumesMove bool `msgpack:"nah_consumes_move"`
}

// DefaultSettings mirrors the official Monopoly Deal: No Mercy rules and
// card counts (see docs/deal-no-mercy/card-list.md). With NumDecks=1 the
// deck composition sums to exactly 120 cards.
func DefaultSettings() Settings {
	return Settings{
		NumDecks:            1,
		StartNumCards:       5,
		MaxHandSize:         7,
		MovesPerTurn:        3,
		TurnDraw:            2,
		WinSetAmount:        3,
		WinMoneyAmount:      0,
		DebtChipsPerPlayer:  3,
		YoinkPayment:        10,
		ShackRentBonus:      5,
		BigPaydayHandTarget: 7,
		Speed:               SpeedMedium,

		SetSnatcherAmount:    3,
		DebtTrapAmount:       3,
		GoAgainAmount:        3,
		HeistAmount:          3,
		MarketCrashAmount:    3,
		BigPaydayAmount:      4,
		RepoManAmount:        2,
		ShackAmount:          3,
		PropertyRaidAmount:   3,
		TaxDayAmount:         2,
		PickpocketAmount:     3,
		BankSwapAmount:       2,
		YoinkAmount:          3,
		NahAmount:            5,
		RentAmount:           1,
		DoubleRentAmount:     1,
		WildRentAmount:       3,
		DoubleRentWildAmount: 3,

		NahConsumesMove: true,
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
