package main

import (
	"context"
	"cryptoswap/exchanges"
	"cryptoswap/handlers"
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
			mainLogger.Error(ctx, "❌ Error loading currencies: %v", err)
			return
		}

		mainLogger.Info(ctx, "✅ Loaded %d currencies in %.2fs",
			len(popular)+len(others), time.Since(start).Seconds())
	}()

	// CoinGecko (para /api/ticker) — opcional API key en free
	cgKey := os.Getenv("COINGECKO_API_KEY")
	cgBase := os.Getenv("COINGECKO_BASE_URL") // opcional (Pro)
	coinGecko := services.NewCoinGeckoService(cgBase, cgKey)

	// Crear handlers
	quoteHandler := handlers.NewQuoteHandler(aggregator)
	currencyHandler := handlers.NewCurrencyHandler(aggregator)
	swapHandler := handlers.NewSwapHandler(aggregator)

	// Configurar router
	r := mux.NewRouter()

	// Middleware de logging y CORS
	r.Use(middlewares.LoggingMiddleware(mainLogger))
	r.Use(middlewares.CorsMiddleware)

	// API endpoints
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

	// Ticker: usamos TU mismo handler (HTML/JSON) pero con datos reales + fallback
	api.HandleFunc("/ticker", handlers.NewGeckoHandler(coinGecko)).Methods("GET")

	// Health check
	api.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{"status":"healthy","exchanges":%d}`, exchangesAdded)))
	}).Methods("GET")

	// Servir archivos estáticos (frontend)
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
	mainLogger.Info(ctx, "🚀 Server starting on http://localhost:%s", port)
	mainLogger.Info(ctx, "📝 Endpoints:")
	mainLogger.Info(ctx, "   - Frontend: http://localhost:%s", port)
	mainLogger.Info(ctx, "   - API Health: http://localhost:%s/api/health", port)
	mainLogger.Info(ctx, "   - Ticker: http://localhost:%s/api/ticker", port)
	mainLogger.Info(ctx, "   - Currencies: http://localhost:%s/api/currencies", port)
	mainLogger.Info(ctx, "──────────────────────────────────────")

	if err := server.ListenAndServe(); err != nil {
		mainLogger.Fatal(ctx, "❌ Server failed to start: %v", err)
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
