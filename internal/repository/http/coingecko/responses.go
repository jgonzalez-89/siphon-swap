package coingecko

import (
	"cryptoswap/internal/services/models"
	"strings"
)

type Coin struct {
	Id                                 string  `json:"id"`
	Symbol                             string  `json:"symbol"`
	CurrentPrice                       float64 `json:"current_price"`
	PriceChangePercentage24h           float64 `json:"price_change_percentage_24h"`
	Image                              string  `json:"image"`
	MarketCap                          float64 `json:"market_cap"`
	MarketCapRank                      int     `json:"market_cap_rank"`
	TotalVolume                        float64 `json:"total_volume"`
	High24h                            float64 `json:"high_24h"`
	Low24h                             float64 `json:"low_24h"`
	ATH                                float64 `json:"ath"`
	ATHChangePercentage                float64 `json:"ath_change_percentage"`
	ATHDate                            string  `json:"ath_date"`
	ATL                                float64 `json:"atl"`
	ATLChangePercentage                float64 `json:"atl_change_percentage"`
	ATLDate                            string  `json:"atl_date"`
	ROI                                any     `json:"roi"`
	LastUpdated                        string  `json:"last_updated"`
	PriceChangePercentage24hInCurrency float64 `json:"price_change_percentage_24h_in_currency"`
	MarketCapChange24h                 float64 `json:"market_cap_change_24h"`
	MarketCapChangePercentage24h       float64 `json:"market_cap_change_percentage_24h"`
	CirculatingSupply                  float64 `json:"circulating_supply"`
	TotalSupply                        float64 `json:"total_supply"`
	MaxSupply                          float64 `json:"max_supply"`
	Name                               string  `json:"name"`
	Network                            string  `json:"network"`
	Provider                           string  `json:"provider"`
	Available                          bool    `json:"available"`
}

func (m Coin) ToTicker() models.Ticker {
	return models.Ticker{
		Name:   m.Name,
		Symbol: strings.ToUpper(m.Symbol),
		Price:  m.CurrentPrice,
		Change: m.PriceChangePercentage24h,
	}
}
