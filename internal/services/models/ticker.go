// models/types.go
package models

import "strings"

// Ticker representa el precio actual de una moneda
type Ticker struct {
	Name   string  `json:"name"`
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price"`
	Change float64 `json:"change_24h"`
}

func (t Ticker) GetLowerSymbol() string {
	return strings.ToLower(t.Symbol)
}
