package models

// Quote representa una cotización de un exchange
type Quote struct {
	From       NetworkPair `json:"from"`
	To         NetworkPair `json:"to"`
	Amount     float64     `json:"amount"`
	Exchange   string      `json:"exchange"`
	Difference float64     `json:"difference"`
}

func (q Quote) IsEmpty() bool {
	return q.Amount == 0
}

func (q Quote) UpdateFromPrice(input float64, currs map[NetworkPair]Currency) Quote {
	fromPrice := currs[q.From].Price
	toPrice := currs[q.To].Price
	theoreticalPrice := input * fromPrice / toPrice
	q.Difference = (q.Amount - theoreticalPrice) / theoreticalPrice * 100
	return q
}

// QuoteRequest representa una solicitud de cotización
type QuoteRequest struct {
	From   string  `json:"from" validate:"required"`
	To     string  `json:"to" validate:"required"`
	Amount float64 `json:"amount" validate:"required,gt=0"`
}
