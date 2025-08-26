package views

import (
	"cryptoswap/internal/services/swap"
	"fmt"
	"net/http"
	"strconv"
)

type TickerViewController struct {
	cg *swap.CoinGeckoService
}

func NewTickerViewController(cg *swap.CoinGeckoService) *TickerViewController {
	return &TickerViewController{cg: cg}
}

// RenderTicker renderiza el ticker en HTML para HTMX
func (vc *TickerViewController) RenderTicker(w http.ResponseWriter, r *http.Request) {
	vs := r.URL.Query().Get("vs")
	n := 6
	if s := r.URL.Query().Get("n"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			n = v
		}
	}

	tickers, err := vc.cg.TopTickers(r.Context(), vs, n)
	if err != nil {
		w.Write([]byte(`<div style="color: #ef4444;">Error loading tickers</div>`))
		return
	}

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

	w.Write([]byte(html))
}
