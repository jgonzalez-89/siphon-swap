package coingecko

import (
	"context"
	"cryptoswap/internal/lib/httpclient"
	"cryptoswap/internal/lib/logger"
	"cryptoswap/internal/services/models"
	"net/http"
	"strconv"

	"github.com/samber/lo"
)

func NewCoinGecko(logger logger.Logger, factory httpclient.Factory) *coinGecko {
	return &coinGecko{
		logger:  logger,
		factory: factory,
	}
}

type CoinGecko interface {
	TopTickers(ctx context.Context, vs string, n int) ([]models.Ticker, error)
}

type coinGecko struct {
	logger  logger.Logger
	factory httpclient.Factory
}

func (s *coinGecko) TopTickers(ctx context.Context, vs string, n int) ([]models.Ticker, error) {
	req := s.factory.NewClient(ctx).
		WithQueryParams("vs_currency", vs).
		WithQueryParams("order", "market_cap_desc").
		WithQueryParams("per_page", strconv.Itoa(n)).
		WithQueryParams("page", "1").
		WithQueryParams("price_change_percentage", "24h").
		Get

	response, err := httpclient.HandleRequest[[]marketsCoin](req, "/coins/markets", http.StatusOK)
	if err != nil {
		return nil, err
	}

	return lo.Map(response, func(item marketsCoin, _ int) models.Ticker {
		return item.ToTicker()
	}), nil
}
