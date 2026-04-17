package errors

import "net/http"

var PlayerNotInGame = NewError("player not in game", http.StatusBadRequest, "GME001")

var PlayerDoesNotHaveCard = NewError("player does not have card", http.StatusBadRequest, "GME002")

var NotPlayersTurn = NewError("not players turn", http.StatusBadRequest, "GME003")

var PlayerHandHasTooManyCards = NewError("player hand has too many cards", http.StatusBadRequest, "GME004")

var CardDoesNotExist = NewError("card does not exist", http.StatusBadRequest, "GME005")

var InvalidCardForAction = NewError("invalid card for action", http.StatusBadRequest, "GME006")

var PropertySetDoesntExist = NewError("property set does not exist", http.StatusBadRequest, "GME007")

var CardCannotBeAssignedToSet = NewError("card cannot be assigned to set", http.StatusBadRequest, "GME008")

var NoMovesLeft = NewError("no moves left", http.StatusBadRequest, "GME009")

var PropertySetIsFull = NewError("property set is full", http.StatusBadRequest, "GME0010")
