package entities

import "cryptoswap/internal/services/models"

func NewQuotePayload(from, to models.NetworkPair, amount float64) QuotePayload {
	return QuotePayload{
		Route: Route{
			From: from,
			To:   to,
		},
		Estimation: "direct",
		Rate:       "floating",
		Amount:     amount,
	}
}

type Route struct {
	From models.NetworkPair `json:"from"`
	To   models.NetworkPair `json:"to"`
}

type QuotePayload struct {
	Route      Route   `json:"route"`
	Estimation string  `json:"estimation"`
	Rate       string  `json:"rate"`
	Amount     float64 `json:"amount"`
}

type QuoteResponse struct {
	EstimatedAmount float64 `json:"estimated_amount"`
}

func (q *QuoteResponse) ToQuote(from, to models.NetworkPair) models.Quote {
	return models.Quote{
		From:     from,
		To:       to,
		Amount:   q.EstimatedAmount,
		Exchange: stealthEx,
	}
}
