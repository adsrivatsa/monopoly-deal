package main

import (
	"encoding/json"
	stderrors "errors"
	"fmt"
	"io"
	"monopoly-deal/internal/errors"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

var maxBytes = 10 << 20

func SetMaxBytes(mb int) {
	maxBytes = mb
}

func Read[I any](w http.ResponseWriter, r *http.Request) (I, error) {
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	var out I
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&out); err != nil {
		return out, errors.Read(err)
	}

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return out, errors.Read(err)
	}

	return out, nil
}

func ReadAndValidate[I any](w http.ResponseWriter, r *http.Request) (I, error) {
	data, err := Read[I](w, r)
	if err != nil {
		return data, err
	}

	return data, ValidateRequest(data)
}

func ValidateRequest(requestPayload any) error {
	err := validate.Struct(requestPayload)
	if err != nil {
		var invalidValidationError *validator.InvalidValidationError
		if stderrors.As(err, &invalidValidationError) {
			return errors.Validation(err)
		}

		var errs errors.ValidationErrors
		for _, err := range err.(validator.ValidationErrors) {
			param := fmt.Sprintf("%s: %s", err.Tag(), err.Param())
			if err.Param() == "" {
				param = err.Tag()
			}
			jsonTag := getJsonTag(requestPayload, err.Field())
			errs = append(errs, errors.NewValidationError(jsonTag, fmt.Sprintf("expected: %s", param)))
		}
		return errors.Validation(errs.Merge())
	}

	return nil
}

func getJsonTag(structure any, fieldName string) string {
	val := reflect.ValueOf(structure)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	typ := val.Type()
	field, found := typ.FieldByName(fieldName)
	if !found {
		return fieldName
	}

	tag := field.Tag.Get("json")
	if tag == "" {
		return fieldName
	}

	tagParts := strings.Split(tag, ",")
	return tagParts[0]
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
	intErr.Message = err.Error()
	WriteHTTP(w, intErr.Status, intErr)
}
