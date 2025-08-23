package stealthex

import (
	"context"
	"cryptoswap/internal/lib/httpclient"
	"cryptoswap/internal/lib/logger"
	"cryptoswap/models"
	"cryptoswap/services"
	"fmt"
	"net/http"

	"github.com/samber/lo"
)

const (
	stealthEx = "StealthEX"
)

type stealthexClientImpl struct {
	factory httpclient.Factory
	logger  logger.Logger
	apiKey  string
}

func NewStealthClient(logger logger.Logger,
	factory httpclient.Factory) services.ExchangeManager {
	return &stealthexClientImpl{
		logger:  logger,
		factory: factory,
	}
}

func (s *stealthexClientImpl) GetName() string {
	return stealthEx
}

func (s *stealthexClientImpl) GetCurrencies(ctx context.Context,
) ([]models.Currency, error) {
	request := s.factory.NewClient(ctx).WithAuthHeader(s.apiKey).Get
	apiCurrencies, err := httpclient.HandleRequest[[]CurrencyResponse](
		request, "/currency", http.StatusOK)
	if err != nil {
		return nil, err
	}

	return lo.Map(apiCurrencies, func(curr CurrencyResponse, _ int) models.Currency {
		return curr.ToCurrency()
	}), nil
}

// GetQuote obtiene una cotización
func (s *stealthexClientImpl) GetQuote(ctx context.Context, from, to string,
	amount float64) (*models.Quote, error) {

	request := s.factory.NewClient(ctx).
		WithAuthHeader(s.apiKey).
		WithQueryParams("amount", amount).
		WithQueryParams("fixed", false).
		Get
	quote, err := httpclient.HandleRequest[QuoteResponse](
		request, "/estimate/"+from+"/"+to, http.StatusOK)
	if err != nil {
		return nil, fmt.Errorf("error fetching quote: %w", err)
	}

	return quote.ToQuote(from, to, amount), nil
}

// GetMinAmount obtiene el monto mínimo para un par
func (s *stealthexClientImpl) GetMinAmount(ctx context.Context, from, to string) (float64, error) {
	request := s.factory.NewClient(ctx).WithAuthHeader(s.apiKey).Get
	minAmount, err := httpclient.HandleRequest[MinAmountResponse](
		request, "/range/"+from+"/"+to, http.StatusOK)
	if err != nil {
		return 0, err
	}

	return minAmount.MinAmount, nil
}

// CreateExchange crea un intercambio real
func (s *stealthexClientImpl) CreateExchange(ctx context.Context,
	req models.SwapRequest) (*models.SwapResponse, error) {

	request := s.factory.NewClient(ctx).
		WithAuthHeader(s.apiKey).
		WithBody(NewExchangePayload(req)).
		Post

	exchange, err := httpclient.HandleRequest[ExchangeResponse](
		request, "/exchange", http.StatusCreated)
	if err != nil {
		return nil, err
	}

	return exchange.ToSwapResponse(req.From, req.To), nil
}

// GetExchangeStatus obtiene el estado de un intercambio
func (s *stealthexClientImpl) GetExchangeStatus(ctx context.Context, id string) (string, error) {
	request := s.factory.NewClient(ctx).
		WithAuthHeader(s.apiKey).Get
	exchange, err := httpclient.HandleRequest[ExchangeResponse](
		request, "/exchange/"+id, http.StatusOK)
	if err != nil {
		return "", err
	}

	return exchange.Status, nil
}
