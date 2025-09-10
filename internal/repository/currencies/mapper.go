package currencies

import (
	"cryptoswap/internal/services/models"

	"github.com/samber/lo"
)

func toCurrenciesEntity(currencies []models.Currency) Currencies {
	return lo.Map(currencies, func(currency models.Currency, _ int) Currency {
		return Currency{
			Symbol:            currency.Symbol,
			Name:              currency.Name,
			Image:             currency.Image,
			Available:         currency.Available,
			Price:             currency.Price,
			AddressValidation: currency.AddressValidation,
			Popular:           currency.IsPopular(),
			Networks: lo.Map(currency.GetNetworks(), func(network models.NetworkPair, _ int) CurrencyNetwork {
				return CurrencyNetwork{
					Symbol:  network.Symbol,
					Network: network.Network,
				}
			}),
		}
	})
}

func toPairSlice(currencies []models.NetworkPair) [][]string {
	return lo.Map(currencies, func(currency models.NetworkPair, _ int) []string {
		return []string{currency.Symbol, currency.Network}
	})
}

func toSwapEntity(swap models.Swap) Swap {
	return Swap{
		Id:            swap.Id,
		FromSymbol:    swap.From.Symbol,
		FromNetwork:   swap.From.Network,
		ToSymbol:      swap.To.Symbol,
		ToNetwork:     swap.To.Network,
		PayinAmount:   swap.PayinAmount,
		PayoutAmount:  swap.PayoutAmount,
		PayoutAddress: swap.PayoutAddress,
		ToAddress:     swap.ToAddress,
		RefundAddress: swap.RefundAddress,
		Exchange:      swap.Exchange,
		Status:        swap.Status,
		CreatedAt:     swap.CreatedAt,
		UpdatedAt:     swap.UpdatedAt,
		ExchangeId:    swap.ExchangeId,
		Reason:        swap.Reason,
	}
}
