package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
)

type PriceBreakdownDetail struct {
	Type      string          `json:"type"`
	Price     decimal.Decimal `json:"price"`
	Summary   string          `json:"summary"`
	Total     decimal.Decimal `json:"total"`
	SuperArea decimal.Decimal `json:"salableArea"`
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
