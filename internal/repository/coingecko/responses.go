package coingecko

import (
	"cryptoswap/internal/services/models"
	"strings"
)

type marketsCoin struct {
	Symbol                   string  `json:"symbol"`
	CurrentPrice             float64 `json:"current_price"`
	PriceChangePercentage24h float64 `json:"price_change_percentage_24h"`
}

func (m marketsCoin) ToTicker() models.Ticker {
	return models.Ticker{
		Symbol: strings.ToUpper(m.Symbol),
		Price:  m.CurrentPrice,
		Change: m.PriceChangePercentage24h,
	}
}
