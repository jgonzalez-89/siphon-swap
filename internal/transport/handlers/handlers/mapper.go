package handlers

import (
	"cryptoswap/internal/services/models"

	"github.com/samber/lo"
)

func toCurrencies(currencies []models.Currency) []Currency {
	return lo.Map(currencies, func(currency models.Currency, _ int) Currency {
		return Currency{
			Name:              currency.Name,
			Symbol:            currency.Symbol,
			Image:             currency.Image,
			Available:         currency.Available,
			AddressValidation: currency.AddressValidation,
			Price:             currency.Price,
			Networks:          toNetworks(currency.GetNetworks()),
		}
	})
}

func toNetworks(networks []models.NetworkPair) []Network {
	return lo.Map(networks, func(network models.NetworkPair, _ int) Network {
		return Network{
			Name: network.Network,
		}
	})
}

func toPair(symbol, network string) models.NetworkPair {
	return models.NetworkPair{
		Symbol:  symbol,
		Network: network,
	}
}

func toFilter(filter GetV1CurrenciesParams) models.Filters {
	return models.Filters{
		Name:    filter.Name,
		Popular: filter.Popular,
		Active:  filter.Active,
		Symbols: filter.Symbols,
	}
}
