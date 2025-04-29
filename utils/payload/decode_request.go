package payload

import (
	"circledigital.in/real-state-erp/utils/custom"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

// decodeJSON decodes the incoming request body
func decodeJSON[T any](w http.ResponseWriter, r *http.Request) (T, error) {
	r.Body = http.MaxBytesReader(w, r.Body, int64(1<<20))

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var payload T
	err := decoder.Decode(&payload)

	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			message := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
			return payload, &custom.RequestError{
				Status:  http.StatusBadRequest,
				Message: message,
			}

		case errors.Is(err, io.ErrUnexpectedEOF):
			message := "Request body contains badly-formed JSON"
			return payload, &custom.RequestError{
				Status:  http.StatusBadRequest,
				Message: message,
			}

		case errors.As(err, &unmarshalTypeError):
			message := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			return payload, &custom.RequestError{
				Status:  http.StatusBadRequest,
				Message: message,
			}

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			message := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			return payload, &custom.RequestError{
				Status:  http.StatusBadRequest,
				Message: message,
			}

		case errors.Is(err, io.EOF):
			message := "Request body must not be empty"
			return payload, &custom.RequestError{
				Status:  http.StatusBadRequest,
				Message: message,
			}

		case err.Error() == "http: request body too large":
			message := "Request body must not be larger than 1MB"
			return payload, &custom.RequestError{
				Status:  http.StatusRequestEntityTooLarge,
				Message: message,
			}

		default:
			log.Printf("Error decoding request body: %v\n", err)
			return payload, err
		}
	}
	err = decoder.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		message := "Request body must only contain a single JSON object"
		return payload, &custom.RequestError{
			Status:  http.StatusBadRequest,
			Message: message,
		}
	}

	return payload, nil
}
