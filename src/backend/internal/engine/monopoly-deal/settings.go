package monopoly_deal

type Settings struct {
	NumDecks            int `json:"num_decks" validate:"required,min=1,max=3"`
	StartNumCards       int `json:"start_num_cards" validate:"required,min=5,max=8"`
	MaxHandSize         int `json:"max_hand_size" validate:"required,min=5,max=10"`
	MovesPerTurn        int `json:"moves_per_turn" validate:"required,min=3,max=5"`
	PassGoDraw          int `json:"pass_go_draw" validate:"required,min=2,max=5"`
	ItsMyBirthdayAmount int `json:"its_my_birthday_amount" validate:"required,min=2,max=5"`
}

func (s Settings) Raw() any {
	return s
}
