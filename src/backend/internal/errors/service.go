package errors

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Entity string

const (
	EntityPlayer     Entity = "player"
	EntityRoom       Entity = "room"
	EntityRoomPlayer Entity = "room_player"
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

var RoomIsFull = NewError("room is full", http.StatusBadRequest, "SER003")

var RoomNotInLobby = NewError("room is not in lobby", http.StatusBadRequest, "SER004")

var PlayerIsNotHost = NewError("player is not host", http.StatusBadRequest, "SER005")

var AllPlayersNotReady = NewError("all players are not ready", http.StatusBadRequest, "SER006")
