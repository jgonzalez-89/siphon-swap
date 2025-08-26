package models

import "time"

// Quote representa una cotización de un exchange
type Quote struct {
	Exchange   string    `json:"exchange"`
	From       string    `json:"from"`
	To         string    `json:"to"`
	FromAmount float64   `json:"from_amount"`
	ToAmount   float64   `json:"to_amount"`
	Rate       float64   `json:"rate"`
	MinAmount  float64   `json:"min_amount,omitempty"`
	MaxAmount  float64   `json:"max_amount,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
}

// QuoteRequest representa una solicitud de cotización
type QuoteRequest struct {
	From   string  `json:"from" validate:"required"`
	To     string  `json:"to" validate:"required"`
	Amount float64 `json:"amount" validate:"required,gt=0"`
}
