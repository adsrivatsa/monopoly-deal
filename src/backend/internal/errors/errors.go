package errors

import (
	"fmt"
)

type Error struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Inner   error  `json:"-"`
}

func NewError(message string, status int, code string, inner ...error) Error {
	var i error
	if len(inner) > 0 {
		i = inner[0]
	}
	return Error{
		Message: message,
		Status:  status,
		Code:    code,
		Inner:   i,
	}
}

func (e Error) Error() string {
	f := fmt.Sprintf("(%s) %s", e.Code, e.Message)
	if e.Inner != nil {
		f = fmt.Sprintf("%s: %s", f, e.Inner.Error())
	}
	return f
}

func (e Error) Unwrap() error {
	return e.Inner
}
