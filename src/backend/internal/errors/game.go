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

var PropertySetIsComplete = NewError("property set is complete", http.StatusBadRequest, "GME0010")

var CannotRespondToDemand = NewError("cannot respond to demand", http.StatusBadRequest, "GME0011")

var CannotRespondNo = NewError("cannot respond no", http.StatusBadRequest, "GME0012")

var PaymentDoesNotCoverAmount = NewError("payment does not cover amount", http.StatusBadRequest, "GME0013")

var ActiveDemandExists = NewError("active demand exists", http.StatusBadRequest, "GME0014")

var DemandDoesNotExist = NewError("demand does not exist", http.StatusBadRequest, "GME0015")

var InvalidAmountOfCards = NewError("invalid amount of cards", http.StatusBadRequest, "GME0016")

var CardCannotBeStolen = NewError("card cannot be stolen", http.StatusBadRequest, "GME0017")

var PropertySetIsNotComplete = NewError("property set is not complete", http.StatusBadRequest, "GME0018")

var CannotStealFromSelf = NewError("cannot steal from self", http.StatusBadRequest, "GME0019")

var PendingRentExists = NewError("pending rent exists", http.StatusBadRequest, "GME0020")

var PendingRentDoesntExist = NewError("pending rent does not exist", http.StatusBadRequest, "GME0021")

var InvalidPropertySets = NewError("invalid property sets", http.StatusBadRequest, "GME0022")

var CannotDiscardYet = NewError("cannot discard yet", http.StatusBadRequest, "GME0023")
