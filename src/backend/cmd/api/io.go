package main

import (
	"encoding/json"
	stderrors "errors"
	"fmt"
	"io"
	"monopoly-deal/internal/errors"
	"net/http"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

var maxBytes = 10 << 20

func SetMaxBytes(mb int) {
	maxBytes = mb
}

func Read[I any](w http.ResponseWriter, r *http.Request) (I, error) {
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	var out I
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&out); err != nil {
		return out, err
	}

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return out, err
	}

	return out, nil
}

func WriteHTTP(w http.ResponseWriter, status int, data any, headers ...http.Header) {
	out, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(out)
	if err != nil {
		fmt.Println(err)
	}
}

func ErrorHTTP(w http.ResponseWriter, err error) {
	var intErr errors.Error
	if !stderrors.As(err, &intErr) {
		stderrors.As(errors.Internal(err), &intErr)
	}
	WriteHTTP(w, intErr.Status, intErr)
}

func WriteWS(conn *websocket.Conn, message proto.Message) {
	data, err := proto.Marshal(message)
	if err != nil {
		fmt.Println(err)
	}

	err = conn.WriteMessage(websocket.BinaryMessage, data)
	if err != nil {
		fmt.Println(err)
	}
}

func LobbyError(conn *websocket.Conn, err error) {
	var intErr errors.Error
	if !stderrors.As(err, &intErr) {
		stderrors.As(errors.Internal(err), &intErr)
	}
	WriteWS(conn, intErr.Lobby())
}

func GameError(conn *websocket.Conn, err error) {
	var intErr errors.Error
	if !stderrors.As(err, &intErr) {
		stderrors.As(errors.Internal(err), &intErr)
	}
	WriteWS(conn, intErr.Game())
}
