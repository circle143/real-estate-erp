package payload

import (
	"encoding/json"
	"log"
	"net/http"
)

// EncodeJSON encodes the given data and sends response to the client
// this method is also called directly by services when sending success response
func EncodeJSON[T any](w http.ResponseWriter, status int, data T) {
	jsonRes, err := json.MarshalIndent(data, "", "\t")

	if err != nil {
		log.Println(err)
		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_, err = w.Write(jsonRes)
	if err != nil {
		log.Println(err)
		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
		return
	}
}