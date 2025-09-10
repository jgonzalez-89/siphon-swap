package currencies

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/samber/lo"

	"cryptoswap/internal/lib/apierrors"
	"cryptoswap/internal/lib/logger"
	"cryptoswap/internal/services/interfaces"
	"cryptoswap/internal/services/models"
)

type CurrencyService interface {
	GetCurrencies(ctx context.Context, filters models.Filters) ([]models.Currency, *apierrors.ApiError)
	GetQuotes(ctx context.Context, from, to models.NetworkPair, amount float64) ([]models.Quote, *apierrors.ApiError)
	GetSwap(ctx context.Context, id string) (models.Swap, *apierrors.ApiError)
	InsertSwap(ctx context.Context, swap models.Swap) (models.Swap, *apierrors.ApiError)
	ProcessSwap(ctx context.Context, swap models.Swap) *apierrors.ApiError
}

func NewCurrencyService(logger logger.Logger, db interfaces.CurrencyRepository,
	notifier interfaces.SwapNotifier,
	exchanges ...interfaces.CurrencyFetcher) *currencyService {
	return &currencyService{
		logger:   logger,
		db:       db,
		notifier: notifier,
		exchanges: lo.SliceToMap(exchanges, func(exchange interfaces.CurrencyFetcher) (string, interfaces.CurrencyFetcher) {
			return exchange.GetExchangeName(), exchange
		}),
	}
}

type currencyService struct {
	logger    logger.Logger
	db        interfaces.CurrencyRepository
	exchanges map[string]interfaces.CurrencyFetcher
	notifier  interfaces.SwapNotifier
}

func (cs *currencyService) GetCurrencies(ctx context.Context, filters models.Filters,
) ([]models.Currency, *apierrors.ApiError) {
	cs.logger.Infof(ctx, "Getting currencies with filters: %+v", filters)

	currencies, err := cs.db.GetCurrencies(ctx, filters)
	if err != nil {
		cs.logger.Errorf(ctx, "Error getting currencies: %+v", err)
		return nil, err
	}

	return currencies, nil
}

func (cs *currencyService) GetQuotes(ctx context.Context, from, to models.NetworkPair,
	amount float64) ([]models.Quote, *apierrors.ApiError) {
	cs.logger.Infof(ctx, "Getting quote for %s to %s with amount %f", from, to, amount)

	currLookup, err := cs.getPairs(ctx, from, to)
	if err != nil {
		return []models.Quote{}, err
	}

	return cs.getQuotesFromAllExchanges(ctx, from, to, amount, currLookup), nil
}

func (cs *currencyService) getQuotesFromAllExchanges(ctx context.Context, from, to models.NetworkPair, amount float64,
	lookup map[models.NetworkPair]models.Currency) []models.Quote {

	quotes := []models.Quote{}
	wg := sync.WaitGroup{}
	for _, exchange := range cs.exchanges {
		wg.Add(1)
		go func(quotes *[]models.Quote) {
			defer wg.Done()
			quote, err := exchange.GetQuote(ctx, from, to, amount)
			if err != nil {
				cs.logger.Error(ctx, err)
				return
			}
			if !quote.IsEmpty() {
				*quotes = append(*quotes, quote.UpdateFromPrice(amount, lookup))
			}
		}(&quotes)
	}
	wg.Wait()

	return quotes
}

func (cs *currencyService) getPairs(ctx context.Context, from, to models.NetworkPair,
) (map[models.NetworkPair]models.Currency, *apierrors.ApiError) {
	currencies, err := cs.db.GetCurrenciesByPairs(ctx, from, to)
	if err != nil {
		cs.logger.Errorf(ctx, "Error getting currencies: %+v", err)
		return nil, err
	}

	// Grab all networks:
	fetchedNetworkLookup := map[models.NetworkPair]models.Currency{}
	for _, currency := range currencies {
		networks := currency.GetNetworks()
		for _, network := range networks {
			fetchedNetworkLookup[network] = currency
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
		return nil, apierrors.NewApiError(apierrors.BadRequest,
			fmt.Errorf("networks %s not found", strings.Join(notFoundPairsStrings, ", ")))
	}

	return fetchedNetworkLookup, nil
}

func (cs *currencyService) GetSwap(ctx context.Context, id string) (models.Swap, *apierrors.ApiError) {
	cs.logger.Infof(ctx, "Getting swap with id: %s", id)

	swap, err := cs.db.GetSwap(ctx, id)
	if err != nil {
		cs.logger.Errorf(ctx, "Error getting swap: %+v", err)
		return models.Swap{}, err
	}

	return swap, nil
}

func (cs *currencyService) InsertSwap(ctx context.Context, swap models.Swap) (models.Swap, *apierrors.ApiError) {
	cs.logger.Infof(ctx, "Inserting swap", swap)

	pairs, err := cs.getPairs(ctx, swap.From, swap.To)
	if err != nil {
		cs.logger.Errorf(ctx, "Error getting swap: %+v", err)
		return models.Swap{}, err
	}

	if err := swap.HasValidAddress(pairs[swap.To]); err != nil {
		cs.logger.Errorf(ctx, "Error validating address: %+v", err)
		return models.Swap{}, err
	}

	// TODO: Go to the requested exchange and create the swap, inform the current swap
	// TODO: Add transactioner, the billing should be somewhere else
	newSwap, err := cs.db.InsertSwap(ctx, *swap.WithBillingConditions("XYZ-address", "xdsq2324adgs", 10))
	if err != nil {
		cs.logger.Errorf(ctx, "Error inserting swap: %+v", err)
		return models.Swap{}, err
	}

	if err := cs.notifier.NotifySwap(ctx, swap); err != nil {
		cs.logger.Errorf(ctx, "Error notifying swap: %+v", err)
		return models.Swap{}, err
	}

	return newSwap, nil
}

func (cs *currencyService) ProcessSwap(ctx context.Context, swap models.Swap) *apierrors.ApiError {
	cs.logger.Infof(ctx, "Updating swap")

	err := cs.db.UpdateSwap(ctx, *swap.Complete())
	if err != nil {
		cs.logger.Errorf(ctx, "Error updating swap: %+v", err)
		return err
	}

	return nil
}
