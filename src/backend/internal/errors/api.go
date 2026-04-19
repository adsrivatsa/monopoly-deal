package errors

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func NewValidationError(field string, message string) ValidationError {
	return ValidationError{
		Field:   field,
		Message: message,
	}
}

type ValidationErrors []ValidationError

func (es ValidationErrors) Merge() error {
	msg := ""
	for _, e := range es {
		msg += fmt.Sprintf("%s: %s, ", e.Field, e.Message)
	}
	msg = strings.TrimSuffix(msg, ", ")
	return errors.New(msg)
}

var (
	InvalidToken        = NewError("invalid token", http.StatusBadRequest, "TOK001")
	ExpiredToken        = NewError("expired token", http.StatusUnauthorized, "TOK002")
	InvalidTokenContent = NewError("invalid token", http.StatusBadRequest, "TOK003")
	InvalidTokenType    = NewError("invalid token", http.StatusBadRequest, "TOK004")
)

func InvalidUUID(err error) Error {
	return NewError("invalid UUID", http.StatusBadRequest, "VAL001", err)
}

var GameNotSupported = NewError("game not supported", http.StatusBadRequest, "VAL002")

func Unauthenticated(err error) Error {
	return NewError("unauthenticated", http.StatusBadRequest, "API001", err)
}

var DuplicateSocket = NewError("duplicate socket created", http.StatusConflict, "API002")

func InvalidMessageType[T any]() Error {
	var expectedType T
	msg := fmt.Sprintf("invalid message type, expected type %t", expectedType)
	return NewError(msg, http.StatusBadRequest, "API003")
}

func Validation(err error) Error {
	return NewError("validation error", http.StatusBadRequest, "API004", err)
}

func Read(err error) Error {
	return NewError("read error", http.StatusBadRequest, "API005", err)
}
