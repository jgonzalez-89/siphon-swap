package currencies

import (
	"context"
	"fmt"
	"strings"

	"github.com/samber/lo"

	"cryptoswap/internal/lib/apierrors"
	"cryptoswap/internal/lib/logger"
	"cryptoswap/internal/services/interfaces"
	"cryptoswap/internal/services/models"
)

type CurrencyService interface {
	GetCurrencies(ctx context.Context, filters models.Filters) ([]models.Currency, *apierrors.ApiError)
	GetQuote(ctx context.Context, from, to models.NetworkPair, amount float64) (models.Quote, *apierrors.ApiError)
}

func NewCurrencyService(logger logger.Logger, db interfaces.CurrencyRepository) *currencyService {
	return &currencyService{
		logger: logger,
		db:     db,
	}
}

type currencyService struct {
	logger    logger.Logger
	db        interfaces.CurrencyRepository
	exchanges map[string]interfaces.CurrencyFetcher
}

func (s *currencyService) GetCurrencies(ctx context.Context, filters models.Filters,
) ([]models.Currency, *apierrors.ApiError) {
	s.logger.Infof(ctx, "Getting currencies with filters: %+v", filters)

	currencies, err := s.db.GetCurrencies(ctx, filters)
	if err != nil {
		s.logger.Errorf(ctx, "Error getting currencies: %+v", err)
		return nil, err
	}

	return currencies, nil
}

func (s *currencyService) GetQuote(ctx context.Context, from, to models.NetworkPair,
	amount float64) (models.Quote, *apierrors.ApiError) {
	s.logger.Infof(ctx, "Getting quote for %s to %s with amount %f", from, to, amount)

	err := s.doPairsExist(ctx, from, to)
	if err != nil {
		return models.Quote{}, err
	}

	return models.Quote{}, nil
}

func (cs *currencyService) doPairsExist(ctx context.Context, from, to models.NetworkPair) *apierrors.ApiError {
	currencies, err := cs.db.GetCurrenciesByPairs(ctx, from, to)
	if err != nil {
		cs.logger.Errorf(ctx, "Error getting currencies: %+v", err)
		return err
	}

	// Grab all networks:
	fetchedNetworkLookup := map[models.NetworkPair]bool{}
	for _, currency := range currencies {
		networks := currency.GetNetworks()
		for _, network := range networks {
			fetchedNetworkLookup[network] = true
		}
	}

	// Check if any of the symbols don't exist and get the not found string symbols
	notFoundPairs := []models.NetworkPair{}
	for _, network := range []models.NetworkPair{from, to} {
		if _, ok := fetchedNetworkLookup[network]; !ok {
			notFoundPairs = append(notFoundPairs, network)
		}
	}

	if len(notFoundPairs) > 0 {
		notFoundPairsStrings := lo.Map(notFoundPairs, func(pair models.NetworkPair, _ int) string {
			return pair.String()
		})
		return apierrors.NewApiError(apierrors.BadRequestError,
			fmt.Errorf("network %s not found", strings.Join(notFoundPairsStrings, ", ")))
	}

	return nil
}
