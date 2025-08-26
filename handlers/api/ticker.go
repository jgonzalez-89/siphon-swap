package api

import (
    "cryptoswap/services"
    "net/http"
    "strconv"
)

type TickerHandler struct {
    cg *services.CoinGeckoService
}

func NewTickerHandler(cg *services.CoinGeckoService) *TickerHandler {
    return &TickerHandler{cg: cg}
}

// GetTicker devuelve los tickers en formato JSON
func (h *TickerHandler) GetTicker(w http.ResponseWriter, r *http.Request) {
    vs := r.URL.Query().Get("vs")
    n := 6
    if s := r.URL.Query().Get("n"); s != "" {
        if v, err := strconv.Atoi(s); err == nil {
            n = v
        }
    }
    
    tickers, err := h.cg.TopTickers(r.Context(), vs, n)
    if err != nil {
        respondWithError(w, "Failed to fetch tickers", http.StatusBadGateway)
        return
    }
    
    respondWithJSON(w, tickers, http.StatusOK)
}

