package currencies

import (
	"context"
	"cryptoswap/internal/lib/apierrors"
	"cryptoswap/internal/lib/logger"
	"cryptoswap/internal/services/interfaces"
	"cryptoswap/internal/services/models"
	"fmt"
	"strings"
)

type CurrencyService interface {
	GetCurrencies(ctx context.Context, filters models.Filters) ([]models.Currency, *apierrors.ApiError)
	GetQuote(ctx context.Context, from, to string, amount float64) (models.Quote, *apierrors.ApiError)
}

func NewCurrencyService(logger logger.Logger, db interfaces.CurrencyRepository) *currencyService {
	return &currencyService{
		logger: logger,
		db:     db,
	}
}

type currencyService struct {
	logger logger.Logger
	db     interfaces.CurrencyRepository
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

func (s *currencyService) GetQuote(ctx context.Context, from, to string,
	amount float64) (models.Quote, *apierrors.ApiError) {
	s.logger.Infof(ctx, "Getting quote for %s to %s with amount %f", from, to, amount)

	err := s.doSymbolsExist(ctx, from, to)
	if err != nil {
		return models.Quote{}, err
	}

	return models.Quote{}, nil
}

func (cs *currencyService) doSymbolsExist(ctx context.Context, from, to string) *apierrors.ApiError {
	symbols := []string{from, to}
	currencies, err := cs.db.GetCurrencies(ctx, models.Filters{Symbols: &symbols})
	if err != nil {
		cs.logger.Errorf(ctx, "Error getting currencies: %+v", err)
		return err
	}

	// Check if any of the symbols don't exist and get the not found string symbols
	notFoundSymbols := []string{}
	for _, currency := range currencies {
		if currency.Symbol == from || currency.Symbol == to {
			continue
		}
		notFoundSymbols = append(notFoundSymbols, currency.Symbol)
	}

	if len(notFoundSymbols) > 0 {
		return apierrors.NewApiError(apierrors.NotFoundError,
			fmt.Errorf("symbols %s not found", strings.Join(notFoundSymbols, ", ")))
	}

	return nil
}
