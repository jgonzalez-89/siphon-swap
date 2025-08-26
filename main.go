package main

import (
	"context"
	"cryptoswap/exchanges"
	"cryptoswap/handlers"
	apiHandlers "cryptoswap/handlers/api"
	viewHandlers "cryptoswap/handlers/views"
	"cryptoswap/internal/lib/logger"
	"cryptoswap/internal/lib/middlewares"
	"cryptoswap/services"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	ctx := context.Background()
	factory := logger.NewLoggerFactory("siphon-swap", "info")
	mainLogger := factory.NewLogger("main")

	// Cargar variables de entorno
	if err := godotenv.Load(); err != nil {
		mainLogger.Error(ctx, "No .env file found, using environment variables")
	}

	// Verificar API keys
	changeNowKey := os.Getenv("CHANGENOW_API_KEY")
	simpleSwapKey := os.Getenv("SIMPLESWAP_API_KEY")
	stealthExKey := os.Getenv("STEALTHEX_API_KEY")
	letsExchangeKey := os.Getenv("LETSEXCHANGE_API_KEY")

	if changeNowKey == "" && simpleSwapKey == "" && stealthExKey == "" && letsExchangeKey == "" {
		mainLogger.Fatal(ctx, "❌ At least one API key is required. Please set CHANGENOW_API_KEY, "+
			"SIMPLESWAP_API_KEY, STEALTHEX_API_KEY, or LETSEXCHANGE_API_KEY in .env file")
	}

	// Crear aggregator
	aggregator := services.NewAggregator()

	// Añadir exchanges disponibles
	exchangesAdded := 0
	exchangesAdded += loadExchange(mainLogger, ctx, aggregator, changeNowKey, "ChangeNOW")
	exchangesAdded += loadExchange(mainLogger, ctx, aggregator, simpleSwapKey, "SimpleSwap")
	exchangesAdded += loadExchange(mainLogger, ctx, aggregator, stealthExKey, "StealthEX")
	exchangesAdded += loadExchange(mainLogger, ctx, aggregator, letsExchangeKey, "LetsExchange")
	mainLogger.Info(ctx, "📊 Total exchanges configured: %d", exchangesAdded)

	// Pre-cargar currencies en background
	go func() {
		mainLogger.Info(ctx, "🔄 Pre-loading currencies...")
		start := time.Now()
		popular, others, err := aggregator.GetAllCurrencies()
		if err != nil {
			mainLogger.Errorf(ctx, "❌ Error loading currencies: %v", err)
			return
		}

		mainLogger.Infof(ctx, "✅ Loaded %d currencies in %.2fs",
			len(popular)+len(others), time.Since(start).Seconds())
	}()

	// CoinGecko service
	cgKey := os.Getenv("COINGECKO_API_KEY")
	cgBase := os.Getenv("COINGECKO_BASE_URL")
	coinGecko := services.NewCoinGeckoService(cgBase, cgKey)

	// ========================================
	// HANDLERS LEGACY (temporalmente)
	// ========================================
	quoteHandler := handlers.NewQuoteHandler(aggregator)
	currencyHandler := handlers.NewCurrencyHandler(aggregator)
	swapHandler := handlers.NewSwapHandler(aggregator)

	// ========================================
	// NUEVOS HANDLERS - API (JSON)
	// ========================================
	apiQuoteHandler := apiHandlers.NewQuoteHandler(aggregator)
	apiSwapHandler := apiHandlers.NewSwapHandler(aggregator)
	apiCurrencyHandler := apiHandlers.NewCurrencyHandler(aggregator)
	apiTickerHandler := apiHandlers.NewTickerHandler(coinGecko)

	// ========================================
	// NUEVOS HANDLERS - Views (HTML/HTMX)
	// ========================================
	quoteViewController := viewHandlers.NewQuoteViewController(aggregator)
	swapViewController := viewHandlers.NewSwapViewController(aggregator)
	currencyViewController := viewHandlers.NewCurrencyViewController(aggregator)
	tickerViewController := viewHandlers.NewTickerViewController(coinGecko)

	// Configurar router
	r := mux.NewRouter()

	// Middleware de logging y CORS
	middlewareLogger := factory.NewLogger("logging-middleware")
	r.Use(middlewares.LoggingMiddleware(middlewareLogger))
	r.Use(middlewares.CorsMiddleware)

	// ========================================
	// API v2 - JSON endpoints
	// ========================================
	apiv2 := r.PathPrefix("/api/v2").Subrouter()
	
	// Quote endpoints
	apiv2.HandleFunc("/quote", apiQuoteHandler.GetQuote).Methods("GET")
	apiv2.HandleFunc("/quotes", apiQuoteHandler.GetAllQuotes).Methods("POST")
	apiv2.HandleFunc("/min-amounts", apiQuoteHandler.GetMinAmounts).Methods("GET")
	
	// Swap endpoints
	apiv2.HandleFunc("/swap", apiSwapHandler.CreateSwap).Methods("POST")
	apiv2.HandleFunc("/swap/{id}/status", apiSwapHandler.GetStatus).Methods("GET")
	
	// Currency endpoints
	apiv2.HandleFunc("/currencies", apiCurrencyHandler.GetAll).Methods("GET")
	apiv2.HandleFunc("/exchanges", apiCurrencyHandler.GetExchanges).Methods("GET")
	
	// Ticker endpoint
	apiv2.HandleFunc("/ticker", apiTickerHandler.GetTicker).Methods("GET")
	
	// Health check v2
	apiv2.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"healthy","exchanges":%d,"version":"2.0"}`, exchangesAdded)
	}).Methods("GET")

	// ========================================
	// HTMX - HTML endpoints
	// ========================================
	htmx := r.PathPrefix("/htmx").Subrouter()
	
	// Quote views
	htmx.HandleFunc("/quote", quoteViewController.RenderQuotes).Methods("POST")
	
	// Swap views
	htmx.HandleFunc("/swap", swapViewController.RenderSwapResult).Methods("POST")
	htmx.HandleFunc("/swap/{id}/status", swapViewController.RenderStatus).Methods("GET")
	
	// Currency views
	htmx.HandleFunc("/currencies", currencyViewController.RenderCurrencyList).Methods("GET")
	htmx.HandleFunc("/currencies/search", currencyViewController.SearchCurrencies).Methods("POST")
	
	// Ticker view
	htmx.HandleFunc("/ticker", tickerViewController.RenderTicker).Methods("GET")

	// ========================================
	// API LEGACY (mantener funcionando)
	// ========================================
	api := r.PathPrefix("/api").Subrouter()

	// Quotes
	api.HandleFunc("/quote", quoteHandler.GetQuote).Methods("GET", "POST")
	api.HandleFunc("/quotes", quoteHandler.GetAllQuotes).Methods("POST")
	api.HandleFunc("/min-amounts", quoteHandler.GetMinAmounts).Methods("GET")

	// Currencies
	api.HandleFunc("/currencies", currencyHandler.GetAll).Methods("GET")
	api.HandleFunc("/exchanges", currencyHandler.GetExchanges).Methods("GET")

	// Swap
	api.HandleFunc("/swap", swapHandler.CreateSwap).Methods("POST")
	api.HandleFunc("/swap/{id}/status", swapHandler.GetStatus).Methods("GET")

	// Ticker
	api.HandleFunc("/ticker", handlers.NewGeckoHandler(coinGecko)).Methods("GET")

	// Health check
	api.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(fmt.Appendf([]byte{}, `{"status":"healthy","exchanges":%d}`, exchangesAdded))
	}).Methods("GET")

	// Servir archivos estáticos
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	// Configurar servidor
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Iniciar servidor
	mainLogger.Infof(ctx, "🚀 Server starting on http://localhost:%s", port)
	mainLogger.Infof(ctx, "📝 Endpoints:")
	mainLogger.Infof(ctx, "   - Frontend: http://localhost:%s", port)
	mainLogger.Infof(ctx, "   - API Legacy: http://localhost:%s/api/*", port)
	mainLogger.Infof(ctx, "   ─────────────────────────────")
	mainLogger.Infof(ctx, "   NEW API v2 (JSON):")
	mainLogger.Infof(ctx, "   - GET  /api/v2/quote")
	mainLogger.Infof(ctx, "   - POST /api/v2/quotes")
	mainLogger.Infof(ctx, "   - GET  /api/v2/min-amounts")
	mainLogger.Infof(ctx, "   - POST /api/v2/swap")
	mainLogger.Infof(ctx, "   - GET  /api/v2/swap/{id}/status")
	mainLogger.Infof(ctx, "   - GET  /api/v2/currencies")
	mainLogger.Infof(ctx, "   - GET  /api/v2/exchanges")
	mainLogger.Infof(ctx, "   - GET  /api/v2/ticker")
	mainLogger.Infof(ctx, "   - GET  /api/v2/health")
	mainLogger.Infof(ctx, "   ─────────────────────────────")
	mainLogger.Infof(ctx, "   NEW HTMX (HTML):")
	mainLogger.Infof(ctx, "   - POST /htmx/quote")
	mainLogger.Infof(ctx, "   - POST /htmx/swap")
	mainLogger.Infof(ctx, "   - GET  /htmx/swap/{id}/status")
	mainLogger.Infof(ctx, "   - GET  /htmx/currencies")
	mainLogger.Infof(ctx, "   - POST /htmx/currencies/search")
	mainLogger.Infof(ctx, "   - GET  /htmx/ticker")
	mainLogger.Infof(ctx, "──────────────────────────────────────")

	if err := server.ListenAndServe(); err != nil {
		mainLogger.Fatalf(ctx, "❌ Server failed to start: %v", err)
	}
}

func loadExchange(mainLogger logger.Logger, ctx context.Context,
	aggregator *services.Aggregator, key, name string) int {
	if key != "" {
		aggregator.AddExchange(exchanges.NewStealthEx(key))
		mainLogger.Infof(ctx, "✅ %s exchange added", name)
		return 1
	}
	return 0
}