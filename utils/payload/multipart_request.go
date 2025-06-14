package payload

import (
	"circledigital.in/real-state-erp/utils/custom"
	"log"
	"net/http"
	"strings"
)

func ParseMultipartForm(w http.ResponseWriter, r *http.Request) error {
	maxSize := 10 << 20 // 10mb
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxSize))
	err := r.ParseMultipartForm(int64(maxSize))
	if err != nil {
		log.Println(err)
		if strings.Contains(err.Error(), "http: request body too large") {
			message := "request body too large. Limit 10MB"
			HandleError(w, &custom.RequestError{Status: http.StatusRequestEntityTooLarge, Message: message})
			return err
		}

		if strings.Contains(err.Error(), "mime: no media type") {
			HandleError(w, &custom.RequestError{
				Status:  http.StatusUnsupportedMediaType,
				Message: http.StatusText(http.StatusUnsupportedMediaType),
			})
			return err
		}

		if strings.Contains(err.Error(), "request Content-Type isn't multipart/form-data") {
			HandleError(w, &custom.RequestError{
				Status:  http.StatusBadRequest,
				Message: "Request Content-Type isn't multipart/form-data",
			})
			return err
		}

		log.Printf("Error parsing multipart form data: %v\n", err)
		HandleError(w, err)
		return err
	}

	return nil
}
