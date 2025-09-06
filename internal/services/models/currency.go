package models

import (
	"strings"
)

type Filters struct {
	Name    *string   `json:"name,omitempty"`
	Popular *bool     `json:"popular,omitempty"`
	Active  *bool     `json:"active,omitempty"`
	Symbols *[]string `json:"symbols,omitempty"`
}

func (f *Filters) ToMap() map[string]any {
	cast := map[string]any{}
	if f.Name != nil {
		cast["name"] = *f.Name
	}
	if f.Popular != nil {
		cast["popular"] = *f.Popular
	}
	if f.Active != nil {
		cast["active"] = *f.Active
	}
	if f.Symbols != nil {
		cast["symbol"] = *f.Symbols
	}
	return cast
}

var popularCurrencies = map[string]bool{
	"btc": true, "eth": true, "usdt": true, "usdc": true,
	"bnb": true, "sol": true, "ada": true, "dot": true,
	"matic": true, "avax": true, "link": true, "uni": true,
	"xrp": true, "ltc": true, "atom": true, "near": true,
}

func NewCurrency(provider, network, symbol, name, addressValidation, image string, available bool) Currency {
	symbol = strings.ToLower(symbol)
	network = strings.ToLower(network)
	nw := newNetworks()
	return Currency{
		Symbol:            symbol,
		Name:              name,
		Image:             image,
		Available:         available,
		AddressValidation: addressValidation,
		Networks:          *nw.Add(symbol, network),
		provider:          provider,
	}
}

type Currency struct {
	Symbol            string   `json:"symbol"`
	Name              string   `json:"name"`
	Image             string   `json:"image,omitempty"`
	Available         bool     `json:"available"`
	AddressValidation string   `json:"addressValidation,omitempty"`
	Price             float64  `json:"price,omitempty"`
	Networks          Networks `json:"networks,omitempty"`
	provider          string
}

func (c Currency) IsPopular() bool {
	return popularCurrencies[c.GetLowerSymbol()]
}

func (c Currency) GetFirstNetwork() NetworkPair {
	return c.Networks.first
}

func (c Currency) WithAddressValidation(addressValidation string) Currency {
	if addressValidation == "" {
		return c
	}

	c.AddressValidation = addressValidation
	return c
}

func (c Currency) WithNetworks(networks ...string) Currency {
	for _, network := range networks {
		c.Networks.Add(c.Symbol, network)
	}
	return c
}

func (c Currency) WithPrice(price float64) Currency {
	c.Price = price
	return c
}

func (c Currency) WithProvider(provider string) Currency {
	c.provider = provider
	return c
}

func (c *Currency) GetNetworks() []NetworkPair {
	return c.Networks.GetAll()
}

func (c *Currency) GetLowerSymbol() string {
	return strings.ToLower(c.Symbol)
}

func (c *Currency) GetUpperSymbol() string {
	return strings.ToUpper(c.Symbol)
}
