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

func toPairFromRequest(pair NetworkPair) models.NetworkPair {
	return models.NetworkPair{
		Symbol:  pair.Symbol,
		Network: pair.Network,
	}
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

func toQuotes(quotes []models.Quote) []Quote {
	return lo.Map(quotes, func(quote models.Quote, _ int) Quote {
		return Quote{
			From:       fromPair(quote.From),
			To:         fromPair(quote.To),
			Amount:     quote.Amount,
			Exchange:   quote.Exchange,
			Difference: quote.Difference,
		}
	})
}

func fromPair(pair models.NetworkPair) NetworkPair {
	return NetworkPair{
		Symbol:  pair.Symbol,
		Network: pair.Network,
	}
}

func toSwap(swap models.Swap) Swap {
	return Swap{
		Id:            swap.Id,
		From:          fromPair(swap.From),
		To:            fromPair(swap.To),
		PayinAmount:   swap.PayinAmount,
		Exchange:      swap.Exchange,
		Status:        swap.Status,
		CreatedAt:     swap.CreatedAt,
		UpdatedAt:     swap.UpdatedAt,
		Reason:        swap.Reason,
		PayoutAddress: swap.PayoutAddress,
		PayoutAmount:  swap.PayoutAmount,
		ToAddress:     swap.ToAddress,
		RefundAddress: swap.RefundAddress,
	}
}
