package changenow

import "cryptoswap/internal/services/models"

type Currency struct {
	Ticker             string `json:"ticker"`
	Name               string `json:"name"`
	Image              string `json:"image"`
	IsAvailable        bool   `json:"isAvailable"`
	IsStable           bool   `json:"isStable"`
	SupportsFixedRate  bool   `json:"supportsFixedRate"`
	Network            string `json:"network"`
	TokenContract      string `json:"tokenContract"`
	Buy                bool   `json:"buy"`
	Sell               bool   `json:"sell"`
	LegacyTicker       string `json:"legacyTicker"`
	IsExtraIdSupported bool   `json:"isExtraIdSupported"`
	IsFiat             bool   `json:"isFiat"`
	Featured           bool   `json:"featured"`
	HasExternalId      bool   `json:"hasExternalId"`
}

func (c Currency) ToModel(provider string) models.Currency {
	return models.NewCurrency(provider, c.Network, c.Ticker, c.Name, "", c.Image, c.IsAvailable)
}
