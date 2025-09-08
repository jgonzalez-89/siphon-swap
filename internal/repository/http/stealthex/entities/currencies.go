package entities

import "cryptoswap/internal/services/models"

const (
	stealthEx = "StealthEX"
)

type CurrencyResponse struct {
	Symbol            string   `json:"symbol"`
	Name              string   `json:"name"`
	Image             string   `json:"icon_url"`
	Network           string   `json:"network"`
	HasExtraId        bool     `json:"has_extra_id"`
	IsStable          bool     `json:"is_stable"`
	ValidationAddress string   `json:"address_regex"`
	ValidationExtra   string   `json:"validation_extra"`
	WarningsFrom      []string `json:"warnings_from"`
	WarningsTo        []string `json:"warnings_to"`
	AddressExplorer   string   `json:"address_explorer"`
	TxExplorer        string   `json:"tx_explorer"`
	ExtraId           string   `json:"extra_id"`
}

func (c *CurrencyResponse) ToNetworkPair() models.NetworkPair {
	return models.NetworkPair{
		Symbol:  c.Symbol,
		Network: c.Network,
	}
}

func (c *CurrencyResponse) ToCurrency() models.Currency {
	return models.NewCurrency(stealthEx, c.Network, c.Symbol, c.Name, c.ValidationAddress, c.Image, true)
}
