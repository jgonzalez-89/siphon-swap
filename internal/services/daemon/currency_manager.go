package daemon

import (
	"context"
	"cryptoswap/internal/lib/constants"
	"cryptoswap/internal/lib/logger"
	"cryptoswap/internal/services/interfaces"
	"cryptoswap/internal/services/models"
	"sync"
	"time"
)

const (
	targetCurrency = "usd"
	defaultResults = 250
	pages          = 100
)

func NewCurrencyManager(logger logger.Logger, repository interfaces.CurrencyRepository,
	cashFetcher interfaces.CashFetcher,
	currencyFetchers ...interfaces.CurrencyFetcher) *currencyManager {
	return &currencyManager{
		logger:           logger,
		repository:       repository,
		cashFetcher:      cashFetcher,
		currencyFetchers: currencyFetchers,
	}
}

type currencyManager struct {
	logger           logger.Logger
	cashFetcher      interfaces.CashFetcher
	currencyFetchers []interfaces.CurrencyFetcher
	repository       interfaces.CurrencyRepository
}

func (cm *currencyManager) Start(ctx context.Context) {
	go cm.runEvery(ctx, time.Minute*5, cm.storeCurrencies)
	time.Sleep(time.Second * 10)
	go cm.runEvery(ctx, time.Minute, cm.updatePrices)
}

func (cm *currencyManager) runEvery(ctx context.Context, interval time.Duration, fn func(context.Context)) {
	fn(constants.AddRequestIdToContext(ctx))
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fn(constants.AddRequestIdToContext(ctx))
		}
	}
}

func (cm *currencyManager) storeCurrencies(ctx context.Context) {
	currencies := &[]models.Currency{}
	wg := &sync.WaitGroup{}
	for _, currencyFetcher := range cm.currencyFetchers {
		wg.Add(1)
		go func(wg *sync.WaitGroup, currencies *[]models.Currency) {
			defer wg.Done()
			currs, err := currencyFetcher.GetCurrencies(ctx)
			if err != nil {
				cm.logger.Errorf(ctx, "Error fetching currencies: %v", err)
				return
			}
			*currencies = append(*currencies, currs...)
		}(wg, currencies)
	}

	wg.Wait()
	cm.logger.Infof(ctx, "Fetched %d currencies", len(*currencies))

	manager := models.NewCurrencies(*currencies...)
	if err := cm.repository.InsertCurrencies(ctx, manager.GetCurrencies()); err != nil {
		cm.logger.Errorf(ctx, "Error inserting currencies: %v", err)
	}
}

func (cm *currencyManager) updatePrices(ctx context.Context) {
	cm.logger.Infof(ctx, "Updating currency prices")
	currencies, err := cm.repository.GetCurrencies(ctx, models.Filters{})
	if err != nil {
		cm.logger.Errorf(ctx, "Error getting currencies: %v", err)
		return
	}

	manager := models.NewCurrencies(currencies...)
	for page := range pages {
		if !manager.HasMorePricesToUpdate() {
			break
		}

		pageTickers, err := cm.cashFetcher.TopTickers(ctx, targetCurrency, defaultResults, page)
		if err != nil {
			cm.logger.Errorf(ctx, "Error getting top tickers: %v", err)
			continue
		}

		if len(pageTickers) == 0 {
			cm.logger.Infof(ctx, "Total tickers exhausted at page %d", page)
			break
		}

		for _, ticker := range pageTickers {
			manager.UpdatePrice(ticker.Symbol, ticker.Price)
		}

		updatedPrices := manager.ExtractPricesToUpdate()
		if err := cm.repository.UpdatePrices(ctx, updatedPrices); err != nil {
			cm.logger.Errorf(ctx, "Error updating prices: %v", err)
		}

		time.Sleep(time.Second * 10)
	}
}
