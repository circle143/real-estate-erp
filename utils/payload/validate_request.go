package payload

import (
	"circledigital.in/real-state-erp/utils/custom"
	"log"
	"net/http"
)

// ValidateAndDecodeRequest validates and decodes incoming http request body
func ValidateAndDecodeRequest[T any](w http.ResponseWriter, r *http.Request) *T {
	payload, err := decodeJSON[T](w, r)
	if err != nil {
		HandleError(w, err)
		return nil
	}

	if err := validatorObj.Struct(payload); err != nil {
		log.Println(err)
		HandleError(w, &custom.RequestError{
			Status:  http.StatusBadRequest,
			Message: "Invalid / missing fields in request body",
		})
		return nil
	}

	return &payload
}
