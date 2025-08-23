package services

import (
	"context"
	"cryptoswap/models"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"
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
	exchanges []Exchange
	cache     *Cache
	mu        sync.RWMutex
}

// NewAggregator crea una nueva instancia del agregador
func NewAggregator() *Aggregator {
	return &Aggregator{
		exchanges: make([]Exchange, 0),
		cache:     NewCache(5 * time.Minute), // Cache de 5 minutos
	}
}

// AddExchange añade un exchange al agregador
func (a *Aggregator) AddExchange(exchange Exchange) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.exchanges = append(a.exchanges, exchange)
	log.Printf("Added exchange: %s", exchange.GetName())
}

// GetExchanges retorna la lista de exchanges
func (a *Aggregator) GetExchanges() []string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	names := make([]string, len(a.exchanges))
	for i, ex := range a.exchanges {
		names[i] = ex.GetName()
	}
	return names
}

// GetAllCurrencies obtiene todas las monedas únicas de todos los exchanges
func (a *Aggregator) GetAllCurrencies() ([]models.Currency, error) {
	// Verificar cache
	cacheKey := "all_currencies"
	if cached := a.cache.Get(cacheKey); cached != nil {
		if currencies, ok := cached.([]models.Currency); ok {
			return currencies, nil
		}
	}

	a.mu.RLock()
	exchanges := a.exchanges
	a.mu.RUnlock()

	// Map para eliminar duplicados
	currencyMap := make(map[string]models.Currency)
	var wg sync.WaitGroup
	var mapMu sync.Mutex

	// Obtener currencies de cada exchange en paralelo
	for _, exchange := range exchanges {
		wg.Add(1)
		go func(ex Exchange) {
			defer wg.Done()

			currencies, err := ex.GetCurrencies()
			if err != nil {
				log.Printf("Error getting currencies from %s: %v", ex.GetName(), err)
				return
			}

			mapMu.Lock()
			for _, curr := range currencies {
				// Si no existe o el nuevo está disponible y el anterior no
				if existing, exists := currencyMap[curr.Symbol]; !exists ||
					(!existing.Available && curr.Available) {
					currencyMap[curr.Symbol] = curr
				}
			}
			mapMu.Unlock()
		}(exchange)
	}

	wg.Wait()

	// Convertir map a slice
	currencies := make([]models.Currency, 0, len(currencyMap))
	for _, curr := range currencyMap {
		currencies = append(currencies, curr)
	}

	// Ordenar alfabéticamente
	sort.Slice(currencies, func(i, j int) bool {
		return currencies[i].Symbol < currencies[j].Symbol
	})

	// Guardar en cache
	a.cache.Set(cacheKey, currencies, 5*time.Minute)

	log.Printf("Loaded %d unique currencies from %d exchanges", len(currencies), len(exchanges))
	return currencies, nil
}

// GetBestQuote obtiene la mejor cotización de todos los exchanges
func (a *Aggregator) GetBestQuote(from, to string, amount float64) (*models.Quote, error) {
	quotes := a.GetAllQuotes(from, to, amount)

	if len(quotes) == 0 {
		return nil, fmt.Errorf("no quotes available for %s -> %s", from, to)
	}

	// La primera es la mejor (ya están ordenadas)
	return quotes[0], nil
}

// GetAllQuotes obtiene cotizaciones de todos los exchanges
func (a *Aggregator) GetAllQuotes(from, to string, amount float64) []*models.Quote {
	// Cache key para este par y cantidad
	cacheKey := fmt.Sprintf("quotes_%s_%s_%.8f", from, to, amount)
	if cached := a.cache.Get(cacheKey); cached != nil {
		if quotes, ok := cached.([]*models.Quote); ok {
			return quotes
		}
	}

	a.mu.RLock()
	exchanges := a.exchanges
	a.mu.RUnlock()

	quotes := make([]*models.Quote, 0, len(exchanges))
	var wg sync.WaitGroup
	var quotesMu sync.Mutex

	// Canal para timeout
	timeout := time.After(10 * time.Second)
	done := make(chan bool)

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
					log.Printf("Error getting quote from %s: %v", ex.GetName(), err)
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
				log.Printf("Timeout getting quote from %s", ex.GetName())
			}
		}(exchange)
	}

	// Esperar a que terminen todos o timeout global
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		// Todos terminaron
	case <-timeout:
		log.Println("Global timeout reached for quotes")
	}

	// Ordenar por mejor precio (mayor ToAmount)
	sort.Slice(quotes, func(i, j int) bool {
		return quotes[i].ToAmount > quotes[j].ToAmount
	})

	// Guardar en cache por 30 segundos
	if len(quotes) > 0 {
		a.cache.Set(cacheKey, quotes, 30*time.Second)
	}

	log.Printf("Got %d quotes for %s -> %s", len(quotes), from, to)
	return quotes
}

// GetMinAmounts obtiene los montos mínimos de todos los exchanges
func (a *Aggregator) GetMinAmounts(from, to string) map[string]float64 {
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
				log.Printf("Error getting min amount from %s: %v", ex.GetName(), err)
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
