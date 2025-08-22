package models

import "time"

// Currency representa una moneda disponible
type Currency struct {
	Symbol    string `json:"symbol"`
	Name      string `json:"name"`
	Image     string `json:"image,omitempty"`
	Network   string `json:"network,omitempty"`
	Available bool   `json:"available"`
}

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
	From   string  `json:"from"`
	To     string  `json:"to"`
	Amount float64 `json:"amount"`
}

// SwapRequest representa una solicitud de intercambio
type SwapRequest struct {
	From          string  `json:"from"`
	To            string  `json:"to"`
	Amount        float64 `json:"amount"`
	ToAddress     string  `json:"to_address"`
	RefundAddress string  `json:"refund_address,omitempty"`
	Exchange      string  `json:"exchange"`
}

// SwapResponse representa la respuesta de un intercambio
type SwapResponse struct {
	ID            string    `json:"id"`
	Status        string    `json:"status"`
	From          string    `json:"from"`        // AÑADIDO
	To            string    `json:"to"`          // AÑADIDO
	PayinAddress  string    `json:"payin_address"`
	PayinAmount   float64   `json:"payin_amount"`
	PayoutAmount  float64   `json:"payout_amount"`
	PayoutAddress string    `json:"payout_address"`
	Exchange      string    `json:"exchange"`
	CreatedAt     time.Time `json:"created_at"`
}

// Ticker representa el precio actual de una moneda
type Ticker struct {
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price"`
	Change float64 `json:"change_24h"`
}