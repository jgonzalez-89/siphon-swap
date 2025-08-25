package handlers

import (
	"cryptoswap/models"
	"cryptoswap/services"
	"encoding/json"
	"fmt"
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
	popular, others, err := h.aggregator.GetAllCurrencies()
	if err != nil {
		http.Error(w, "Error fetching currencies", http.StatusInternalServerError)
		return
	}

	// Si es una petición HTMX (para los selects), devolver opciones HTML
	if r.Header.Get("HX-Request") == "true" {
		h.renderSelectOptions(w, popular, others)
		return
	}

	// Si no, devolver JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(append(popular, others...))
}

// renderSelectOptions renderiza las monedas como opciones de select
func (h *CurrencyHandler) renderSelectOptions(w http.ResponseWriter,
	popular, others []models.Currency) {
	// Renderizar HTML
	html := `<optgroup label="Popular">`
	for _, curr := range popular {
		html += fmt.Sprintf(`<option value="%s">%s - %s</option>`,
			curr.GetLowerSymbol(),
			curr.GetUpperSymbol(),
			curr.Network)
	}
	html += `</optgroup>`

	html += `<optgroup label="All Currencies">`
	for _, curr := range others {
		html += fmt.Sprintf(`<option value="%s">%s - %s</option>`,
			strings.ToLower(curr.Symbol),
			strings.ToUpper(curr.Symbol),
			curr.Name)
	}
	html += `</optgroup>`

	w.Write([]byte(html))
}

// GetExchanges obtiene la lista de exchanges disponibles
func (h *CurrencyHandler) GetExchanges(w http.ResponseWriter, r *http.Request) {
	exchanges := h.aggregator.GetExchanges()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exchanges)
}
