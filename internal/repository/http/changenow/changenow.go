package changenow

import (
	"context"
	"cryptoswap/internal/lib/apierrors"
	"cryptoswap/internal/lib/httpclient"
	"cryptoswap/internal/lib/logger"
	"cryptoswap/internal/services/models"
	"net/http"

	"github.com/samber/lo"
)

const (
	apiName = "ChangeNOW"
)

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

func (cn *changeNowRepository) GetName() string {
	return apiName
}

func (cn *changeNowRepository) GetCurrencies(ctx context.Context) ([]models.Currency, *apierrors.ApiError) {
	request := cn.factory.NewClient(ctx).WithQueryParams("active", "true").Get

	currencies, err := httpclient.HandleRequest[[]Currency](request, "/exchange/currencies", http.StatusOK)
	if err != nil {
		return []models.Currency{}, apierrors.NewApiError(apierrors.InternalServerError, err)
	}

	return lo.Map(currencies, func(curr Currency, _ int) models.Currency {
		return curr.ToModel(cn.GetName())
	}), nil
}
