package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"cryptoswap/services"
)

type TickerHandler struct {
	CG *services.CoinGeckoService
}

func NewTickerHandler(cg *services.CoinGeckoService) *TickerHandler {
	return &TickerHandler{CG: cg}
}

// Soporta query params: ?vs=eur&n=8
func (h *TickerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vs := r.URL.Query().Get("vs")
	n := 6
	if s := r.URL.Query().Get("n"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			n = v
		}
	}

	tickers, err := h.CG.TopTickers(r.Context(), vs, n)
	if err != nil {
		// Puedes devolver último caché conocido aquí si lo almacenaras aparte.
		http.Error(w, "failed to fetch tickers", http.StatusBadGateway)
		return
	}

	// Si es HTMX, devolver HTML igual que antes
	if r.Header.Get("HX-Request") == "true" {
		html := ""
		for _, t := range tickers {
			color := "#10b981"
			sign := "+"
			if t.Change < 0 {
				color = "#ef4444"
				sign = ""
			}
			html += fmt.Sprintf(`
			<div style="display:flex;align-items:center;gap:8px;">
				<span style="font-weight:600;color:white;">%s</span>
				<span>$%.2f</span>
				<span style="font-size:12px;color:%s;">%s%.1f%%</span>
			</div>`, t.Symbol, t.Price, color, sign, t.Change)
		}
		_, _ = w.Write([]byte(html))
		return
	}

	// JSON por defecto
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tickers)
}