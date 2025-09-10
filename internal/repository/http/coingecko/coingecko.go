package coingecko

import (
	"context"
	"cryptoswap/internal/lib/httpclient"
	"cryptoswap/internal/lib/logger"
	"cryptoswap/internal/services/models"
	"net/http"

	"github.com/samber/lo"
)

func NewCoinGecko(logger logger.Logger, factory httpclient.Factory) *coinGecko {
	return &coinGecko{
		logger:  logger,
		factory: factory,
	}
}

type coinGecko struct {
	logger  logger.Logger
	factory httpclient.Factory
}

func (s *coinGecko) TopTickers(ctx context.Context, targetCurrency string, results, page int) ([]models.Ticker, error) {
	req := s.factory.NewClient(ctx).
		WithQueryParams("vs_currency", targetCurrency).
		WithQueryParams("order", "market_cap_desc").
		WithQueryParams("per_page", results).
		WithQueryParams("page", page).
		WithQueryParams("price_change_percentage", "24h").
		Get

	response, err := httpclient.HandleRequest[[]Coin](req, "/coins/markets", http.StatusOK)
	if err != nil {
		return nil, err
	}

	return lo.Map(response, func(item Coin, _ int) models.Ticker {
		return item.ToTicker()
	}), nil
}
