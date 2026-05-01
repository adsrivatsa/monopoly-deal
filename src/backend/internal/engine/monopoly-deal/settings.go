package monopoly_deal

import (
	"time"

	"github.com/vmihailenco/msgpack/v5"
)

type Speed int

const (
	SpeedSlow Speed = iota
	SpeedMedium
	SpeedFast
)

type Settings struct {
	NumDecks            int           `msgpack:"num_decks" validate:"required,min=1,max=3"`
	StartNumCards       int           `msgpack:"start_num_cards" validate:"required,min=5,max=8"`
	MaxHandSize         int           `msgpack:"max_hand_size" validate:"required,min=5,max=10"`
	MovesPerTurn        int           `msgpack:"moves_per_turn" validate:"required,min=3,max=5"`
	PassGoDraw          int           `msgpack:"pass_go_draw" validate:"required,min=2,max=5"`
	ItsMyBirthdayAmount int           `msgpack:"its_my_birthday_amount" validate:"required,min=2,max=5"`
	DebtCollectorAmount int           `msgpack:"debt_collector_amount" validate:"required,min=5,max=8"`
	WinSetAmount        int           `msgpack:"win_set_amount" validate:"required,min=3,max=6"`
	WinMoneyAmount      int           `msgpack:"win_money_amount" validate:"min=0,max=40"`
	Speed               Speed         `msgpack:"speed" validate:"required,min=0,max=2"`
	MoveTimeout         time.Duration `msgpack:"move_timeout"`
	DemandTimeout       time.Duration `msgpack:"demand_timeout"`
}

func (s *Settings) Encode() ([]byte, error) {
	if s == nil {
		return nil, nil
	}
	return msgpack.Marshal(*s)
}

func (s *Settings) Decode(data []byte) error {
	err := msgpack.Unmarshal(data, s)
	if err != nil {
		return err
	}

	switch s.Speed {
	case SpeedSlow:
		s.MoveTimeout = time.Second * 45
		s.DemandTimeout = time.Second * 30
	case SpeedMedium:
		s.MoveTimeout = time.Second * 30
		s.DemandTimeout = time.Second * 20
	case SpeedFast:
		s.MoveTimeout = time.Second * 15
		s.DemandTimeout = time.Second * 10
	}

	return nil
}
