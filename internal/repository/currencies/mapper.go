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
