package interfaces

import (
	"context"
	"cryptoswap/internal/lib/apierrors"
	"cryptoswap/internal/services/models"
)

type CashFetcher interface {
	TopTickers(ctx context.Context, targetCurrency string, results, page int) ([]models.Ticker, error)
}

type CurrencyFetcher interface {
	GetCurrencies(ctx context.Context) ([]models.Currency, *apierrors.ApiError)
}

type CurrencyRepository interface {
	GetCurrencies(ctx context.Context, filters models.Filters) ([]models.Currency, *apierrors.ApiError)
	GetCurrenciesByPairs(ctx context.Context, pairs ...models.NetworkPair) ([]models.Currency, *apierrors.ApiError)
	InsertCurrencies(ctx context.Context, currencies []models.Currency) *apierrors.ApiError
	UpdatePrices(ctx context.Context, currencies []models.Currency) *apierrors.ApiError
}
