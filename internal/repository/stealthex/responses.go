package stealthex

import (
	"cryptoswap/models"
	"time"
)

type CurrencyResponse struct {
	Symbol     string `json:"symbol"`
	Name       string `json:"name"`
	Image      string `json:"image"`
	Network    string `json:"network"`
	HasExtraId bool   `json:"has_extra_id"`
	IsStable   bool   `json:"is_stable"`
}

func (c *CurrencyResponse) ToCurrency() models.Currency {
	return models.Currency{
		Symbol:    c.Symbol,
		Name:      c.Name,
		Image:     c.Image,
		Network:   c.Network,
		Available: true,
	}
}

type QuoteResponse struct {
	EstimatedAmount float64 `json:"estimated_amount"`
	MinAmountResponse
}

type MinAmountResponse struct {
	MinAmount float64 `json:"min_amount"`
	MaxAmount float64 `json:"max_amount"`
}

func (q *QuoteResponse) ToQuote(from, to string, amount float64) *models.Quote {
	return &models.Quote{
		Exchange:   stealthEx,
		From:       from,
		To:         to,
		FromAmount: amount,
		ToAmount:   q.EstimatedAmount,
		Rate:       q.EstimatedAmount / amount,
		MinAmount:  q.MinAmountResponse.MinAmount,
		MaxAmount:  q.MinAmountResponse.MaxAmount,
		Timestamp:  time.Now(),
	}
}

func NewExchangePayload(req models.SwapRequest) *ExchangePayload {
	return &ExchangePayload{
		ExtraIdTo:     "",
		RefundAddress: req.RefundAddress,
		RateId:        "",
		Exchange: Exchange{
			CurrencyFrom: req.From,
			CurrencyTo:   req.To,
			AmountFrom:   req.Amount,
			AddressTo:    req.ToAddress,
		},
	}
}

type Exchange struct {
	AmountFrom   float64 `json:"amount_from"`
	AddressTo    string  `json:"address_to"`
	CurrencyFrom string  `json:"currency_from"`
	CurrencyTo   string  `json:"currency_to"`
}

type ExchangePayload struct {
	ExtraIdTo     string `json:"extra_id_to"`
	RefundAddress string `json:"refund_address"`
	RateId        string `json:"rate_id"`
	Exchange
}

type ExchangeResponse struct {
	Id          string  `json:"id"`
	Status      string  `json:"status"`
	CreatedAt   string  `json:"created_at"`
	AddressFrom string  `json:"address_from"`
	AmountFrom  float64 `json:"amount_from"`
	AmountTo    float64 `json:"amount_to"`
	Exchange
}

func (e *ExchangeResponse) ToSwapResponse(from, to string) *models.SwapResponse {
	createdAt, _ := time.Parse(time.RFC3339, e.CreatedAt)
	return &models.SwapResponse{
		ID:            e.Id,
		Status:        e.Status,
		From:          from,
		To:            to,
		PayinAddress:  e.AddressFrom,
		PayinAmount:   e.AmountFrom,
		PayoutAmount:  e.AmountTo,
		PayoutAddress: e.AddressTo,
	}
}
}
