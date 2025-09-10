package changenow

import (
	"context"
	"cryptoswap/internal/lib/apierrors"
	"cryptoswap/internal/lib/httpclient"
	"cryptoswap/internal/lib/logger"
	"cryptoswap/internal/services/interfaces"
	"cryptoswap/internal/services/models"
	"net/http"

	"github.com/samber/lo"
)

const (
	apiName = "ChangeNOW"
)

var _ interfaces.CurrencyFetcher = &changeNowRepository{}

func NewChangeNowRepository(logger logger.Logger,
	factory httpclient.Factory) *changeNowRepository {
	return &changeNowRepository{
		logger:  logger,
		factory: factory,
	}
}

type changeNowRepository struct {
	logger  logger.Logger
	factory httpclient.Factory
}

func (cn *changeNowRepository) GetExchangeName() string {
	return apiName
}

func (cn *changeNowRepository) GetCurrencies(ctx context.Context) ([]models.Currency, *apierrors.ApiError) {
	request := cn.factory.NewClient(ctx).WithQueryParams("active", "true").Get

	currencies, err := httpclient.HandleRequest[[]Currency](request, "/exchange/currencies", http.StatusOK)
	if err != nil {
		return []models.Currency{}, apierrors.NewApiError(apierrors.InternalServer, err)
	}

	return lo.Map(currencies, func(curr Currency, _ int) models.Currency {
		return curr.ToModel(cn.GetExchangeName())
	}), nil
}

func (cn *changeNowRepository) GetQuote(ctx context.Context, from, to models.NetworkPair,
	amount float64) (models.Quote, *apierrors.ApiError) {
	return models.Quote{}, nil
}
