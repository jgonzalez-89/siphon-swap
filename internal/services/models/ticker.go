// models/types.go
package models

// Ticker representa el precio actual de una moneda
type Ticker struct {
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price"`
	Change float64 `json:"change_24h"`
}
