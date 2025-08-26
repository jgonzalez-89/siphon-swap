package handlers

import (
	"cryptoswap/internal/services/swap"
	"net/http"
)

type CurrencyHandler struct {
	aggregator *swap.Aggregator
}

func NewCurrencyHandler(aggregator *swap.Aggregator) *CurrencyHandler {
	return &CurrencyHandler{
		aggregator: aggregator,
	}
}

// GetAll obtiene todas las monedas disponibles (JSON)
func (h *CurrencyHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	popular, others, err := h.aggregator.GetAllCurrencies(r.Context())
	if err != nil {
		respondWithError(w, "Error fetching currencies", http.StatusInternalServerError)
		return
	}

	// Combinar popular y others en una sola respuesta
	allCurrencies := append(popular, others...)
	respondWithJSON(w, allCurrencies, http.StatusOK)
}

// GetExchanges obtiene la lista de exchanges disponibles
func (h *CurrencyHandler) GetExchanges(w http.ResponseWriter, r *http.Request) {
	exchanges := h.aggregator.GetExchanges(r.Context())
	respondWithJSON(w, exchanges, http.StatusOK)
}
