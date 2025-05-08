package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type PriceBreakdownDetail struct {
	Type      string  `json:"type"`
	Price     float64 `json:"price"`
	Summary   string  `json:"summary"`
	Total     float64 `json:"total"`
	SuperArea float64 `json:"superArea"`
}

type PriceBreakdownDetails []PriceBreakdownDetail

func (p PriceBreakdownDetails) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *PriceBreakdownDetails) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal PriceBreakdownDetails: %v", value)
	}
	return json.Unmarshal(bytes, p)
}
