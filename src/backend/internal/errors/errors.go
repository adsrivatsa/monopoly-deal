package errors

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Error struct {
	Message  string `json:"-"`
	Status   int    `json:"status"`
	Code     string `json:"code"`
	Inner    error  `json:"-"`
	Rendered string `json:"message"`
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

func (e Error) Render() Error {
	e.Rendered = e.Error()
	return e
}

var (
	InvalidToken        = NewError("invalid token", http.StatusBadRequest, "TOK001")
	ExpiredToken        = NewError("expired token", http.StatusUnauthorized, "TOK002")
	InvalidTokenContent = NewError("invalid token", http.StatusBadRequest, "TOK003")
	InvalidTokenType    = NewError("invalid token", http.StatusBadRequest, "TOK004")
)

func Unauthenticated(err error) Error {
	return NewError("unauthenticated", http.StatusBadRequest, "AUTH001", err)
}

func Internal(err error) Error {
	return NewError("internal error", http.StatusInternalServerError, "INT001", err)
}

type Entity string

const (
	EntityPlayer Entity = "player"
)

type DBViolation string

const (
	ForeignKeyViolation DBViolation = "23503"
	UniqueViolation     DBViolation = "23505"
	NotNullViolation    DBViolation = "23502"
	NoDataFound         DBViolation = "P0002"
)

func InvalidDBViolation(code DBViolation, err error) Error {
	f := fmt.Sprintf("db violation function not set for %s", code)
	return NewError(f, http.StatusInternalServerError, "INT002", err)
}

func DBError(cases map[DBViolation]func(err error) Error, err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
		fn, ok := cases[NoDataFound]
		if !ok {
			return InvalidDBViolation(NoDataFound, err)
		}
		return fn(err)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		code := DBViolation(pgErr.Code)
		fn, ok := cases[code]
		if !ok {
			return InvalidDBViolation(code, err)
		}
		return fn(err)
	}

	return Internal(err)
}

func EntityNotFound(ent Entity) func(err error) Error {
	f := fmt.Sprintf("%s not found", ent)
	return func(err error) Error {
		return NewError(f, http.StatusNotFound, "SER001", err)
	}
}

func EntityAlreadyExists(ent Entity) func(err error) Error {
	f := fmt.Sprintf("%s already exists", ent)
	return func(err error) Error {
		return NewError(f, http.StatusBadRequest, "SER002", err)
	}
}
