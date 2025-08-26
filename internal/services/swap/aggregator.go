// services/aggregator.go
package swap

import (
	"context"
	"cryptoswap/internal/lib/cache"
	"cryptoswap/internal/lib/logger"
	"cryptoswap/internal/services/models"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/icamacho1/Primitives/pkg/maps"
	"github.com/samber/lo"
)

const (
	restCurrenciesCache    = "rest_currencies"
	popularCurrenciesCache = "popular_currencies"
)

var DefaultCacheDuration = 5 * time.Minute

var (
	popularSymbolLookup = map[string]bool{
		"btc": true, "eth": true, "usdt": true, "usdc": true,
		"bnb": true, "sol": true, "ada": true, "dot": true,
		"matic": true, "avax": true, "link": true, "uni": true,
		"xrp": true, "ltc": true, "atom": true, "near": true,
	}
)

// Exchange es la interfaz que deben implementar todos los exchanges
type Exchange interface {
	GetName() string
	GetCurrencies() ([]models.Currency, error)
	GetQuote(from, to string, amount float64) (*models.Quote, error)
	GetMinAmount(from, to string) (float64, error)
	CreateExchange(req models.SwapRequest) (*models.SwapResponse, error)
}

type ExchangeManager interface {
	GetName() string
	GetCurrencies(ctx context.Context) ([]models.Currency, error)
	GetQuote(ctx context.Context, from, to string, amount float64) (*models.Quote, error)
	GetMinAmount(ctx context.Context, from, to string) (float64, error)
	CreateExchange(ctx context.Context, req models.SwapRequest) (*models.SwapResponse, error)
}

// Aggregator coordina múltiples exchanges
type Aggregator struct {
	logger    logger.Logger
	exchanges map[string]Exchange
	cache     *cache.Cache
	mu        sync.RWMutex
}

// NewAggregator crea una nueva instancia del agregador
func NewAggregator(logger logger.Logger) *Aggregator {
	return &Aggregator{
		logger:    logger,
		exchanges: make(map[string]Exchange),
		cache:     cache.NewCache(5 * time.Minute), // Cache de 5 minutos
	}
}

// AddExchange añade un exchange al agregador
func (a *Aggregator) AddExchange(exchange Exchange) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, ok := a.exchanges[exchange.GetName()]; !ok {
		a.exchanges[exchange.GetName()] = exchange
	}
}

// GetExchanges retorna la lista de exchanges
func (a *Aggregator) GetExchanges(ctx context.Context) []string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return lo.Keys(a.exchanges)
}

func (a *Aggregator) getCurrenciesFromCache(ctx context.Context) ([]models.Currency,
	[]models.Currency, bool) {
	popular, ok := a.cache.Get(popularCurrenciesCache)
	if !ok {
		return nil, nil, false
	}

	others, ok := a.cache.Get(restCurrenciesCache)
	if !ok {
		return nil, nil, false
	}

	a.logger.Infof(ctx, "Loaded %d popular currencies and %d rest currencies from cache",
		len(popular.([]models.Currency)), len(others.([]models.Currency)))
	return popular.([]models.Currency), others.([]models.Currency), true
}

// GetAllCurrencies obtiene todas las monedas únicas de todos los exchanges
func (a *Aggregator) GetAllCurrencies(ctx context.Context) (popular []models.Currency,
	others []models.Currency, err error) {

	if popular, others, ok := a.getCurrenciesFromCache(ctx); ok {
		return popular, others, nil
	}

	a.mu.RLock()
	exchanges := a.exchanges
	a.mu.RUnlock()

	var wg sync.WaitGroup
	var mapMu sync.Mutex

	// Obtener currencies de cada exchange en paralelo
	popularCurrenciesLookup := maps.New[models.CurrencyKey, models.Currency]()
	restCurrenciesLookup := maps.New[models.CurrencyKey, models.Currency]()
	for _, exchange := range exchanges {
		wg.Add(1)
		go func(ex Exchange) {
			defer wg.Done()

			currencies, err := ex.GetCurrencies()
			if err != nil {
				a.logger.Errorf(ctx, "Error getting currencies from %s: %v", ex.GetName(), err)
				return
			}

			mapMu.Lock()
			for _, curr := range currencies {
				if _, ok := popularSymbolLookup[curr.GetLowerSymbol()]; ok {
					popularCurrenciesLookup.Add(curr.GetKey(), curr)
					continue
				}
				restCurrenciesLookup.Add(curr.GetKey(), curr)
			}
			mapMu.Unlock()
		}(exchange)
	}

	wg.Wait()

	a.sortCurrenciesAndSetToCache(popularCurrenciesLookup, popularCurrenciesCache)
	a.sortCurrenciesAndSetToCache(restCurrenciesLookup, restCurrenciesCache)

	a.logger.Infof(ctx, "Loaded %d unique currencies from %d exchanges",
		len(popularCurrenciesLookup)+len(restCurrenciesLookup), len(exchanges))
	return popularCurrenciesLookup.Values(), restCurrenciesLookup.Values(), nil
}

func (a *Aggregator) sortCurrenciesAndSetToCache(
	lookup maps.Map[models.CurrencyKey, models.Currency], cacheKey string) {

	currencies := lookup.Values()
	sort.Slice(currencies, func(i, j int) bool {
		return currencies[i].Symbol < currencies[j].Symbol
	})

	a.cache.Set(cacheKey, currencies, DefaultCacheDuration)
}

// GetBestQuote obtiene la mejor cotización de todos los exchanges
func (a *Aggregator) GetBestQuote(ctx context.Context, from, to string, amount float64) (*models.Quote, error) {
	quotes := a.GetAllQuotes(ctx, from, to, amount)

	if len(quotes) == 0 {
		a.logger.Errorf(ctx, "no quotes available for %s -> %s", from, to)
		return nil, fmt.Errorf("no quotes available for %s -> %s", from, to)
	}

	// La primera es la mejor (ya están ordenadas)
	return quotes[0], nil
}

// GetAllQuotes obtiene cotizaciones de todos los exchanges
func (a *Aggregator) GetAllQuotes(ctx context.Context, from, to string, amount float64) []*models.Quote {
	// No need to compute shit
	if from == to {
		return []*models.Quote{}
	}

	// Cache key para este par y cantidad
	cacheKey := fmt.Sprintf("quotes_%s_%s_%.8f", from, to, amount)
	if cached, ok := a.cache.Get(cacheKey); ok {
		return cached.([]*models.Quote)
	}

	a.mu.RLock()
	exchanges := a.exchanges
	a.mu.RUnlock()

	quotes := make([]*models.Quote, 0, len(exchanges))
	var wg sync.WaitGroup
	var quotesMu sync.Mutex

	// Obtener quotes de cada exchange en paralelo
	for _, exchange := range exchanges {
		wg.Add(1)
		go func(ex Exchange) {
			defer wg.Done()

			// Timeout individual por exchange
			quoteChan := make(chan *models.Quote, 1)
			go func() {
				quote, err := ex.GetQuote(from, to, amount)
				if err != nil {
					a.logger.Errorf(ctx, "Error getting quote from %s: %v", ex.GetName(), err)
					quoteChan <- nil
					return
				}
				quoteChan <- quote
			}()

			select {
			case quote := <-quoteChan:
				if quote != nil && quote.ToAmount > 0 {
					quotesMu.Lock()
					quotes = append(quotes, quote)
					quotesMu.Unlock()
				}
			case <-time.After(5 * time.Second):
				a.logger.Errorf(ctx, "Timeout getting quote from %s", ex.GetName())
			}
		}(exchange)
	}
	wg.Wait()

	// Ordenar por mejor precio (mayor ToAmount)
	sort.Slice(quotes, func(i, j int) bool {
		return quotes[i].ToAmount > quotes[j].ToAmount
	})

	// Guardar en cache por 30 segundos
	if len(quotes) > 0 {
		a.cache.Set(cacheKey, quotes, 30*time.Second)
	}

	a.logger.Infof(ctx, "Got %d quotes for %s -> %s", len(quotes), from, to)
	return quotes
}

// GetMinAmounts obtiene los montos mínimos de todos los exchanges
func (a *Aggregator) GetMinAmounts(ctx context.Context, from, to string) map[string]float64 {
	a.mu.RLock()
	exchanges := a.exchanges
	a.mu.RUnlock()

	minAmounts := make(map[string]float64)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, exchange := range exchanges {
		wg.Add(1)
		go func(ex Exchange) {
			defer wg.Done()

			minAmount, err := ex.GetMinAmount(from, to)
			if err != nil {
				a.logger.Errorf(ctx, "Error getting min amount from %s: %v", ex.GetName(), err)
				return
			}

			mu.Lock()
			minAmounts[ex.GetName()] = minAmount
			mu.Unlock()
		}(exchange)
	}

	wg.Wait()
	return minAmounts
}

// CreateExchange crea un intercambio en el exchange especificado
func (a *Aggregator) CreateExchange(req models.SwapRequest) (*models.SwapResponse, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Buscar el exchange especificado
	var selectedExchange Exchange
	for _, ex := range a.exchanges {
		if ex.GetName() == req.Exchange {
			selectedExchange = ex
			break
		}
	}

	if selectedExchange == nil {
		return nil, fmt.Errorf("exchange %s not found", req.Exchange)
	}

	// Crear el intercambio
	response, err := selectedExchange.CreateExchange(req)
	if err != nil {
		return nil, fmt.Errorf("error creating exchange: %w", err)
	}

	log.Printf("Created exchange %s on %s", response.ID, req.Exchange)
	return response, nil
}

// GetExchangeByName obtiene un exchange por nombre
func (a *Aggregator) GetExchangeByName(name string) Exchange {
	a.mu.RLock()
	defer a.mu.RUnlock()

	for _, ex := range a.exchanges {
		if ex.GetName() == name {
			return ex
		}
	}
	return nil
}
