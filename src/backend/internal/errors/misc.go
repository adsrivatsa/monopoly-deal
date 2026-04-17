package errors

import "net/http"

func Internal(err error) Error {
	return NewError("internal error", http.StatusInternalServerError, "INT001", err)
}
