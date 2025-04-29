package custom

import (
	"database/sql/driver"
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

const dateFormat = "2006-01-02"

func (d *DateOnly) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		*d = DateOnly{}
		return nil
	}
	t, err := time.Parse(dateFormat, s)
	if err != nil {
		log.Println(err)
		return &RequestError{
			Status:  http.StatusBadRequest,
			Message: "Date in invalid format (expected YYYY-MM-DD)",
		}
	}
	d.Time = t
	return nil
}

func (d *DateOnly) MarshalJSON() ([]byte, error) {
	if d.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf(`"%s"`, d.Time.Format(dateFormat))), nil
}

func (d DateOnly) Value() (driver.Value, error) {
	if d.Time.IsZero() {
		return nil, nil
	}
	return d.Time.Format(dateFormat), nil
}

func (d *DateOnly) Scan(value interface{}) error {
	if value == nil {
		*d = DateOnly{}
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		d.Time = v
	case string:
		t, err := time.Parse(dateFormat, v)
		if err != nil {
			return err
		}
		d.Time = t
	case []byte:
		t, err := time.Parse(dateFormat, string(v))
		if err != nil {
			return err
		}
		d.Time = t
	default:
		return fmt.Errorf("cannot scan type %T into DateOnly", value)
	}
	return nil
}
