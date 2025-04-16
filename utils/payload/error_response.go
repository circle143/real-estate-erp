package payload

import (
	"circledigital.in/real-state-erp/utils/custom"
	"errors"
	"net/http"
)

// errorResponse creates a json error response for client
func errorResponse(w http.ResponseWriter, err error, status ...int) {
	statusCode := http.StatusBadRequest
	if len(status) >= 1 {
		statusCode = status[0]
	}

	var payload custom.JSONResponse

	payload.Error = true
	payload.Message = err.Error()

	EncodeJSON(w, statusCode, payload)
}

// HandleError is used by services to send error to clients
func HandleError(w http.ResponseWriter, err error) {
	var reqErr *custom.RequestError

	if errors.As(err, &reqErr) {
		errorResponse(w, reqErr, reqErr.Status)
	} else {
		err = errors.New(http.StatusText(http.StatusInternalServerError))

		errorResponse(w, err, http.StatusInternalServerError)
	}
}