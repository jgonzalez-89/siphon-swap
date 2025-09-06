package models

import (
	"strings"

	"github.com/samber/lo"
)

func NewCurrencies(currencies ...Currency) Currencies {
	currencyLookup := make(map[string]Currency)
	for _, currency := range currencies {
		symbol := currency.GetLowerSymbol()
		network := currency.GetFirstNetwork().Network
		if currency, ok := currencyLookup[symbol]; ok {
			currencyLookup[symbol] = currency.WithNetworks(network).
				WithAddressValidation(currency.AddressValidation)
			continue
		}
		currencyLookup[symbol] = currency
	}

	return Currencies{
		currencies:    currencyLookup,
		updatedPrices: make(map[string]Currency),
	}
}

type Currencies struct {
	currencies    map[string]Currency
	updatedPrices map[string]Currency
}

func (c *Currencies) GetCurrencies() []Currency {
	return lo.Values(c.currencies)
}

func (c *Currencies) ExtractPricesToUpdate() []Currency {
	defer c.clearUpdatedPrices()
	return lo.Values(c.updatedPrices)
}

func (c *Currencies) clearUpdatedPrices() {
	c.updatedPrices = make(map[string]Currency)
}

func (c Currencies) HasMorePricesToUpdate() bool {
	return len(c.currencies) != 0
}

func (c Currencies) Has(symbol string) bool {
	_, ok := c.currencies[strings.ToLower(symbol)]
	return ok
}

func (c Currencies) UpdatePrice(symbol string, price float64) bool {
	symbol = strings.ToLower(symbol)
	if !c.Has(symbol) {
		return false
	}

	c.updatedPrices[symbol] = c.currencies[symbol].WithPrice(price)
	delete(c.currencies, symbol)
	return true
}
