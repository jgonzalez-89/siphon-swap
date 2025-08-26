package models

import "strings"

type CurrencyKey struct {
	Symbol  string
	Network string
}

// Currency representa una moneda disponible
type Currency struct {
	Symbol    string `json:"symbol"`
	Name      string `json:"name"`
	Image     string `json:"image,omitempty"`
	Network   string `json:"network,omitempty"`
	Available bool   `json:"available"`
	provider  string
}

func (c Currency) WithProvider(provider string) Currency {
	c.provider = provider
	return c
}

func (c *Currency) GetKey() CurrencyKey {
	return CurrencyKey{
		Symbol:  strings.ToUpper(c.Symbol),
		Network: strings.ToUpper(c.Network),
	}
}

func (c *Currency) GetLowerSymbol() string {
	return strings.ToLower(c.Symbol)
}

func (c *Currency) GetUpperSymbol() string {
	return strings.ToUpper(c.Symbol)
}
