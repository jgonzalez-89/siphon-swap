package main

import (
	"context"
	"cryptoswap/exchanges"
	viewHandlers "cryptoswap/handlers/views"
	"cryptoswap/internal/lib/httpclient"
	"cryptoswap/internal/lib/logger"
	"cryptoswap/internal/lib/middlewares"
	"cryptoswap/internal/repository/coingecko"
	"cryptoswap/internal/services/swap"
	apiHandlers "cryptoswap/internal/transport/handlers"
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
	aggregator := swap.NewAggregator(factory.NewLogger("aggregator"))

	// Añadir exchanges disponibles
	exchangesAdded := 0
	if changeNowKey != "" {
		aggregator.AddExchange(exchanges.NewChangeNow(changeNowKey))
		mainLogger.Infof(ctx, "✅ ChangeNOW exchange added")
		exchangesAdded++
	}
	if simpleSwapKey != "" {
		aggregator.AddExchange(exchanges.NewSimpleSwap(simpleSwapKey))
		mainLogger.Infof(ctx, "✅ SimpleSwap exchange added")
		exchangesAdded++
	}
	if stealthExKey != "" {
		aggregator.AddExchange(exchanges.NewStealthEx(stealthExKey))
		mainLogger.Infof(ctx, "✅ StealthEX exchange added")
		exchangesAdded++
	}
	if letsExchangeKey != "" {
		aggregator.AddExchange(exchanges.NewLetsExchange(letsExchangeKey))
		mainLogger.Infof(ctx, "✅ LetsExchange exchange added")
		exchangesAdded++
	}

	mainLogger.Infof(ctx, "📊 Total exchanges configured: %d", exchangesAdded)

	// Pre-cargar currencies en background
	go func() {
		mainLogger.Info(ctx, "🔄 Pre-loading currencies...")
		start := time.Now()
		popular, others, err := aggregator.GetAllCurrencies(ctx)
		if err != nil {
			mainLogger.Errorf(ctx, "❌ Error loading currencies: %v", err)
			return
		}

		mainLogger.Infof(ctx, "✅ Loaded %d currencies in %.2fs",
			len(popular)+len(others), time.Since(start).Seconds())
	}()

	// CoinGecko service
	cgKey := getEnv("COINGECKO_API_KEY", "")
	cgBase := getEnv("COINGECKO_BASE_URL", "https://api.coingecko.com/api/v3")
	cgHttpFactory := httpclient.NewFactory(httpclient.NewConfigWithAuthHeader(cgBase, cgKey,
		"x-cg-demo-api-key", 10*time.Second), factory.NewLogger("coingecko"))
	coinGecko := swap.NewCoinGeckoService(factory.NewLogger("coingecko"),
		coingecko.NewCoinGecko(factory.NewLogger("coingecko"), cgHttpFactory))

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
	pageViewController := viewHandlers.NewPageViewController(aggregator)

	// Configurar router
	r := mux.NewRouter()

	// Middleware de logging y CORS
	middlewareLogger := factory.NewLogger("logging-middleware")
	r.Use(middlewares.LoggingMiddleware(middlewareLogger))
	r.Use(middlewares.CorsMiddleware)

	// ========================================
	// SERVIR ARCHIVOS ESTÁTICOS
	// ========================================
	// IMPORTANTE: Estas rutas deben ir PRIMERO para que tengan prioridad
	r.PathPrefix("/static/css/").Handler(
		http.StripPrefix("/static/css/", http.FileServer(http.Dir("./static/css/"))),
	)
	r.PathPrefix("/static/js/").Handler(
		http.StripPrefix("/static/js/", http.FileServer(http.Dir("./static/js/"))),
	)
	r.PathPrefix("/static/images/").Handler(
		http.StripPrefix("/static/images/", http.FileServer(http.Dir("./static/images/"))),
	)

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

	// Health check
	api.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"healthy","exchanges":%d}`, exchangesAdded)
	}).Methods("GET")

	// ========================================
	// PÁGINA PRINCIPAL Y FALLBACK
	// ========================================
	// Página principal con templates
	r.HandleFunc("/", pageViewController.RenderIndex).Methods("GET")

	// Fallback para servir el index.html estático si falla el template
	r.HandleFunc("/index.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	}).Methods("GET")

	// Cualquier otra ruta estática no capturada
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
	mainLogger.Infof(ctx, "   - Frontend (Templates): http://localhost:%s/", port)
	mainLogger.Infof(ctx, "   - Frontend (Static Fallback): http://localhost:%s/index.html", port)
	mainLogger.Infof(ctx, "   - Static Assets: http://localhost:%s/static/*", port)
	mainLogger.Infof(ctx, "   ─────────────────────────────")
	mainLogger.Infof(ctx, "   API Legacy: http://localhost:%s/api/*", port)
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
	mainLogger.Infof(ctx, "   API Legacy: http://localhost:%s/api/*", port)
	mainLogger.Infof(ctx, "   NEW API v2: http://localhost:%s/api/v2/*", port)
	mainLogger.Infof(ctx, "   NEW HTMX: http://localhost:%s/htmx/*", port)
	mainLogger.Infof(ctx, "──────────────────────────────────────")

	if err := server.ListenAndServe(); err != nil {
		mainLogger.Fatalf(ctx, "❌ Server failed to start: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
