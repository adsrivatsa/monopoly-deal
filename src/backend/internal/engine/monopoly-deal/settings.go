package monopoly_deal

import "github.com/vmihailenco/msgpack/v5"

type Settings struct {
	NumDecks            int `msgpack:"num_decks" validate:"required,min=1,max=3"`
	StartNumCards       int `msgpack:"start_num_cards" validate:"required,min=5,max=8"`
	MaxHandSize         int `msgpack:"max_hand_size" validate:"required,min=5,max=10"`
	MovesPerTurn        int `msgpack:"moves_per_turn" validate:"required,min=3,max=5"`
	PassGoDraw          int `msgpack:"pass_go_draw" validate:"required,min=2,max=5"`
	ItsMyBirthdayAmount int `msgpack:"its_my_birthday_amount" validate:"required,min=2,max=5"`
	DebtCollectorAmount int `msgpack:"debt_collector_amount" validate:"required,min=5,max=8"`
}

func (s Settings) Raw() any {
	return s
}

func (s Settings) Encode() ([]byte, error) {
	return msgpack.Marshal(s)
}
