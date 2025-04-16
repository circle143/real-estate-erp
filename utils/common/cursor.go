package common

import (
	"circledigital.in/real-state-erp/utils/custom"
	"encoding/base64"
	"net/http"
	"time"
)

// encodeCursor encodes the given time in base64 encoding to form a valid cursor for pagination
func encodeCursor(t time.Time) string {
	return base64.URLEncoding.EncodeToString([]byte(t.Format(time.RFC3339Nano)))
}

// DecodeCursor decodes the given cursor to time.Time
func DecodeCursor(cursor string) (time.Time, error) {
	decodedBytes, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, &custom.RequestError{
			Status:  http.StatusBadRequest,
			Message: "Invalid cursor provided.",
		}
	}
	val, err := time.Parse(time.RFC3339Nano, string(decodedBytes))
	if err != nil {
		return time.Time{}, &custom.RequestError{
			Status:  http.StatusBadRequest,
			Message: "Invalid cursor provided.",
		}
	}

	return val, nil
}