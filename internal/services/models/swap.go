package models

import "time"

// SwapRequest representa una solicitud de intercambio
type SwapRequest struct {
	From          string  `json:"from" validate:"required"`
	To            string  `json:"to" validate:"required"`
	Amount        float64 `json:"amount" validate:"required,gt=0"`
	ToAddress     string  `json:"to_address" validate:"required"`
	RefundAddress string  `json:"refund_address,omitempty" validate:"required"`
	Exchange      string  `json:"exchange" validate:"required"`
}

// SwapResponse representa la respuesta de un intercambio
type SwapResponse struct {
	ID            string    `json:"id"`
	Status        string    `json:"status"`
	From          string    `json:"from"` // AÑADIDO
	To            string    `json:"to"`   // AÑADIDO
	PayinAddress  string    `json:"payin_address"`
	PayinAmount   float64   `json:"payin_amount"`
	PayoutAmount  float64   `json:"payout_amount"`
	PayoutAddress string    `json:"payout_address"`
	Exchange      string    `json:"exchange"`
	CreatedAt     time.Time `json:"created_at"`
}
