package main

import (
	"cryptoswap/exchanges"
	"cryptoswap/handlers"
	"cryptoswap/services"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	// Cargar variables de entorno
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Verificar API keys
	changeNowKey := os.Getenv("CHANGENOW_API_KEY")
	simpleSwapKey := os.Getenv("SIMPLESWAP_API_KEY")
	stealthExKey := os.Getenv("STEALTHEX_API_KEY")
	letsExchangeKey := os.Getenv("LETSEXCHANGE_API_KEY")

	if changeNowKey == "" && simpleSwapKey == "" && stealthExKey == "" && letsExchangeKey == "" {
		log.Fatal("❌ At least one API key is required. Please set CHANGENOW_API_KEY, SIMPLESWAP_API_KEY, STEALTHEX_API_KEY, or LETSEXCHANGE_API_KEY in .env file")
	}

	// Crear aggregator
	aggregator := services.NewAggregator()

	// Añadir exchanges disponibles
	exchangesAdded := 0

	if changeNowKey != "" {
		aggregator.AddExchange(exchanges.NewChangeNow(changeNowKey))
		log.Println("✅ ChangeNOW exchange added")
		exchangesAdded++
	}

	if simpleSwapKey != "" {
		aggregator.AddExchange(exchanges.NewSimpleSwap(simpleSwapKey))
		log.Println("✅ SimpleSwap exchange added")
		exchangesAdded++
	}

	if stealthExKey != "" {
		aggregator.AddExchange(exchanges.NewStealthEx(stealthExKey))
		log.Println("✅ StealthEX exchange added")
		exchangesAdded++
	}

	if letsExchangeKey != "" {
		aggregator.AddExchange(exchanges.NewLetsExchange(letsExchangeKey))
		log.Println("✅ LetsExchange exchange added")
		exchangesAdded++
	}

	log.Printf("📊 Total exchanges configured: %d", exchangesAdded)

	// Pre-cargar currencies en background
	go func() {
		log.Println("🔄 Pre-loading currencies...")
		start := time.Now()
		currencies, err := aggregator.GetAllCurrencies()
		if err != nil {
			log.Printf("❌ Error loading currencies: %v", err)
		} else {
			log.Printf("✅ Loaded %d currencies in %.2fs", len(currencies), time.Since(start).Seconds())
		}
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
	r.Use(loggingMiddleware)
	r.Use(corsMiddleware)

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
	api.HandleFunc("/ticker", makeHandleTicker(coinGecko)).Methods("GET")

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
	log.Printf("🚀 Server starting on http://localhost:%s", port)
	log.Println("📝 Endpoints:")
	log.Printf("   - Frontend: http://localhost:%s", port)
	log.Printf("   - API Health: http://localhost:%s/api/health", port)
	log.Printf("   - Ticker: http://localhost:%s/api/ticker", port)
	log.Printf("   - Currencies: http://localhost:%s/api/currencies", port)
	log.Println("──────────────────────────────────────")

	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

// loggingMiddleware registra todas las peticiones
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap ResponseWriter para capturar el status
		wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}

		next.ServeHTTP(wrapped, r)

		// No loguear archivos estáticos
		if r.URL.Path != "/" && r.URL.Path != "/favicon.ico" {
			log.Printf("%s %s %d %v", r.Method, r.URL.Path, wrapped.statusCode, time.Since(start))
		}
	})
}

// corsMiddleware añade headers CORS
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, HX-Request")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// responseWriter wrapper para capturar status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// makeHandleTicker devuelve TU handler original pero con datos de CoinGecko si hay,
// y con fallback a los mock para que la UI nunca se quede en blanco.
func makeHandleTicker(cg *services.CoinGeckoService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Fallback por defecto (tus mocks)
		type T struct {
			Symbol string  `json:"symbol"`
			Price  float64 `json:"price"`
			Change float64 `json:"change"`
		}
		tickers := []T{
			{"BTC", 67432.50, 2.3},
			{"ETH", 3856.75, -0.8},
			{"SOL", 178.23, 5.4},
			{"BNB", 456.40, 1.2},
			{"MATIC", 0.92, 3.1},
			{"ADA", 0.65, -1.2},
		}

		// Parámetros opcionales
		vs := r.URL.Query().Get("vs")
		n := 6
		if s := r.URL.Query().Get("n"); s != "" {
			if v, err := strconv.Atoi(s); err == nil && v > 0 {
				n = v
			}
		}

		// Intentar datos reales de CoinGecko
		if cg != nil {
			if live, err := cg.TopTickers(r.Context(), vs, n); err == nil && len(live) > 0 {
				tmp := make([]T, 0, len(live))
				for _, t := range live {
					tmp = append(tmp, T{
						Symbol: t.Symbol,
						Price:  t.Price,
						Change: t.Change, // % 24h
					})
				}
				tickers = tmp
			} else if err != nil {
				log.Printf("⚠️ CoinGecko fallback (usando mocks): %v", err)
			}
		}

		// Si es HTMX, devolver HTML (tu mismo markup)
		if r.Header.Get("HX-Request") == "true" {
			html := ""
			for _, ticker := range tickers {
				color := "#10b981"
				sign := "+"
				if ticker.Change < 0 {
					color = "#ef4444"
					sign = ""
				}
				html += fmt.Sprintf(`
				<div style="display: flex; align-items: center; gap: 8px;">
					<span style="font-weight: 600; color: white;">%s</span>
					<span>$%.2f</span>
					<span style="font-size: 12px; color: %s;">%s%.1f%%</span>
				</div>`, ticker.Symbol, ticker.Price, color, sign, ticker.Change)
			}
			w.Write([]byte(html))
			return
		}

		// Si no, devolver JSON (compat con json:"change")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tickers)
	}
}
