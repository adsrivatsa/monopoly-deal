package main

import (
	"encoding/json"
	stderrors "errors"
	"io"
	"monopoly-deal/internal/errors"
	"net/http"
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

func Write(w http.ResponseWriter, status int, data any, headers ...http.Header) error {
	out, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(out)
	return err
}

func Error(w http.ResponseWriter, err error) {
	var intErr errors.Error
	if !stderrors.As(err, &intErr) {
		stderrors.As(errors.Internal(err), &intErr)
	}
	Write(w, intErr.Status, intErr.Render())
}
