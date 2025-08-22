package handlers

import (
	"cryptoswap/models"
	"cryptoswap/services"
	"encoding/json"
	"html/template"
	"net/http"
	"strings"
)

type CurrencyHandler struct {
	aggregator *services.Aggregator
}

func NewCurrencyHandler(aggregator *services.Aggregator) *CurrencyHandler {
	return &CurrencyHandler{
		aggregator: aggregator,
	}
}

// GetAll obtiene todas las monedas disponibles
func (h *CurrencyHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	currencies, err := h.aggregator.GetAllCurrencies()
	if err != nil {
		http.Error(w, "Error fetching currencies", http.StatusInternalServerError)
		return
	}
	
	// Si es una petición HTMX para un select, devolver opciones HTML
	if r.Header.Get("HX-Request") == "true" && r.URL.Query().Get("format") == "options" {
		h.renderOptions(w, currencies)
		return
	}
	
	// Si no, devolver JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(currencies)
}

// renderOptions renderiza las monedas como opciones de select
func (h *CurrencyHandler) renderOptions(w http.ResponseWriter, currencies []models.Currency) {
	// Priorizar monedas populares
	popularSymbols := map[string]bool{
		"btc": true, "eth": true, "usdt": true, "usdc": true,
		"bnb": true, "sol": true, "ada": true, "dot": true,
		"matic": true, "avax": true, "link": true, "uni": true,
	}
	
	tmpl := `
	<optgroup label="Popular">
		{{range .Popular}}
		<option value="{{.Symbol}}">{{.Symbol | upper}} - {{.Name}}</option>
		{{end}}
	</optgroup>
	<optgroup label="All Currencies">
		{{range .All}}
		<option value="{{.Symbol}}">{{.Symbol | upper}} - {{.Name}}</option>
		{{end}}
	</optgroup>`
	
	// Separar populares del resto
	popular := make([]models.Currency, 0)
	all := make([]models.Currency, 0)
	
	for _, curr := range currencies {
		if popularSymbols[curr.Symbol] {
			popular = append(popular, curr)
		} else {
			all = append(all, curr)
		}
	}
	
	funcMap := template.FuncMap{
		"upper": strings.ToUpper,
	}
	
	t, err := template.New("options").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	
	data := struct {
		Popular []models.Currency
		All     []models.Currency
	}{
		Popular: popular,
		All:     all,
	}
	
	t.Execute(w, data)
}

// GetExchanges obtiene la lista de exchanges disponibles
func (h *CurrencyHandler) GetExchanges(w http.ResponseWriter, r *http.Request) {
	exchanges := h.aggregator.GetExchanges()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exchanges)
}