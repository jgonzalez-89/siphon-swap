package models

// Quote representa una cotización de un exchange
type Quote struct {
	From   NetworkPair `json:"from"`
	To     NetworkPair `json:"to"`
	Amount float64     `json:"amount"`
}

// QuoteRequest representa una solicitud de cotización
type QuoteRequest struct {
	From   string  `json:"from" validate:"required"`
	To     string  `json:"to" validate:"required"`
	Amount float64 `json:"amount" validate:"required,gt=0"`
}
