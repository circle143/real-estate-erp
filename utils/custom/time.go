package custom

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// DateOnly field define date with format year-month-day
type DateOnly struct {
	time.Time
}

func (d *DateOnly) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		// Leave a pointer as nil by not modifying the object
		*d = DateOnly{}
		return nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		log.Println(err)
		return &RequestError{
			Status:  http.StatusBadRequest,
			Message: "Date in invalid format",
		}
	}
	d.Time = t
	return nil
}

func (d *DateOnly) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, d.Time.Format("2006-01-02"))), nil
}
