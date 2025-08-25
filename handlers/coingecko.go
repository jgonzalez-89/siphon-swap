package handlers

import (
	"cryptoswap/services"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// makeHandleTicker devuelve TU handler original pero con datos de CoinGecko si hay,
// y con fallback a los mock para que la UI nunca se quede en blanco.
func NewGeckoHandler(cg *services.CoinGeckoService) http.HandlerFunc {
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
