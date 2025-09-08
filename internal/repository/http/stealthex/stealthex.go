package stealthex

import (
	"context"
	"cryptoswap/internal/lib/apierrors"
	"cryptoswap/internal/lib/httpclient"
	"cryptoswap/internal/lib/logger"
	"cryptoswap/internal/repository/http/stealthex/entities"
	"cryptoswap/internal/services/interfaces"
	"cryptoswap/internal/services/models"
	"net/http"
	"time"

	"github.com/samber/lo"
)

const (
	stealthEx = "StealthEX"
	limit     = 250
)

var _ interfaces.CurrencyFetcher = &stealthexClientImpl{}

type stealthexClientImpl struct {
	factory httpclient.Factory
	logger  logger.Logger
}

func NewStealthExRepository(logger logger.Logger,
	factory httpclient.Factory) *stealthexClientImpl {
	return &stealthexClientImpl{
		logger:  logger,
		factory: factory,
	}
}

func (s *stealthexClientImpl) GetName() string {
	return stealthEx
}

func (s *stealthexClientImpl) GetCurrencies(ctx context.Context,
) ([]models.Currency, *apierrors.ApiError) {

	count := 0
	currs := make([]entities.CurrencyResponse, 0)
	for {
		offset := getOffset(count)
		request := s.factory.NewClient(ctx).
			WithQueryParams("limit", limit).
			WithQueryParams("offset", offset).Get
		apiCurrencies, err := httpclient.HandleRequest[[]entities.CurrencyResponse](
			request, "/currencies", http.StatusOK)

		if len(apiCurrencies) == 0 {
			break
		}

		if err != nil {
			s.logger.Error(ctx, "error fetching currencies", "error", err)
			break
		}

		count++
		time.Sleep(100 * time.Millisecond)
		currs = append(currs, apiCurrencies...)
	}

	return lo.Map(currs, func(curr entities.CurrencyResponse, _ int) models.Currency {
		return curr.ToCurrency()
	}), nil
}

func getOffset(count int) int {
	if count == 0 {
		return 0
	}
	return count*limit + 1
}

func (s *stealthexClientImpl) GetQuote(ctx context.Context, from, to models.NetworkPair,
	amount float64) (models.Quote, *apierrors.ApiError) {

	payload := entities.NewQuotePayload(from, to, amount)
	request := s.factory.NewClient(ctx).
		WithBody(payload).
		Post

	quote, err := httpclient.HandleRequest[entities.QuoteResponse](
		request, "/rates/estimated-amount", http.StatusOK)
	if err != nil {
		return models.Quote{}, apierrors.NewApiError(apierrors.InternalServerError, err)
	}

	return quote.ToQuote(from, to), nil
}
