package errors

import (
	"database/sql"
	"errors"
	"fmt"
	"monopoly-deal/internal/schema"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Error struct {
	Message string `json:"-"`
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

func (e Error) Proto() *schema.ServerMessage {
	res := &schema.ServerMessage{
		Payload: &schema.ServerMessage_Error{
			Error: &schema.Error{
				Code:    e.Code,
				Message: e.Error(),
				Status:  int32(e.Status),
			},
		},
	}
	return res
}

var (
	InvalidToken        = NewError("invalid token", http.StatusBadRequest, "TOK001")
	ExpiredToken        = NewError("expired token", http.StatusUnauthorized, "TOK002")
	InvalidTokenContent = NewError("invalid token", http.StatusBadRequest, "TOK003")
	InvalidTokenType    = NewError("invalid token", http.StatusBadRequest, "TOK004")
	DuplicateSocket     = NewError("duplicate socket created", http.StatusConflict, "API002")
)

func InvalidUUID(err error) Error {
	return NewError("invalid UUID", http.StatusBadRequest, "VAL001", err)
}

func InvalidMessageType[T any]() Error {
	var expectedType T
	msg := fmt.Sprintf("invalid message type, expected type %t", expectedType)
	return NewError(msg, http.StatusBadRequest, "API003")
}

func Unauthenticated(err error) Error {
	return NewError("unauthenticated", http.StatusBadRequest, "API001", err)
}

func Internal(err error) Error {
	return NewError("internal error", http.StatusInternalServerError, "INT001", err)
}

type Entity string

const (
	EntityPlayer Entity = "player"
	EntityRoom   Entity = "room"
)

type DBViolation string

const (
	ForeignKeyViolation DBViolation = "23503"
	UniqueViolation     DBViolation = "23505"
	NotNullViolation    DBViolation = "23502"
	NoDataFound         DBViolation = "P0002"
)

func DBErrorCode(err error) DBViolation {
	if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
		return NoDataFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		code := DBViolation(pgErr.Code)
		return code
	}

	return ""
}

func EntityNotFound(ent Entity, err ...error) Error {
	f := fmt.Sprintf("%s not found", ent)
	return NewError(f, http.StatusNotFound, "SER001", err...)
}

func EntityAlreadyExists(ent Entity, err ...error) Error {
	f := fmt.Sprintf("%s already exists", ent)
	return NewError(f, http.StatusBadRequest, "SER002", err...)
}
