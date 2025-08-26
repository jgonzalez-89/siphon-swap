package handlers

import (
	"cryptoswap/internal/services/models"
	"cryptoswap/internal/services/swap"
	"encoding/json"
	"net/http"
	"strconv"
)

type QuoteHandler struct {
	aggregator *swap.Aggregator
}

func NewQuoteHandler(aggregator *swap.Aggregator) *QuoteHandler {
	return &QuoteHandler{
		aggregator: aggregator,
	}
}

// GetQuote maneja solicitudes JSON de cotización
func (h *QuoteHandler) GetQuote(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	amountStr := r.URL.Query().Get("amount")

	if from == "" || to == "" || amountStr == "" {
		respondWithError(w, "Missing parameters", http.StatusBadRequest)
		return
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		respondWithError(w, "Invalid amount", http.StatusBadRequest)
		return
	}

	quotes := h.aggregator.GetAllQuotes(r.Context(), from, to, amount)
	if len(quotes) == 0 {
		respondWithError(w, "No quotes available", http.StatusNotFound)
		return
	}

	respondWithJSON(w, quotes, http.StatusOK)
}

// GetAllQuotes maneja solicitudes POST para múltiples cotizaciones
func (h *QuoteHandler) GetAllQuotes(w http.ResponseWriter, r *http.Request) {
	var req models.QuoteRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.From == "" || req.To == "" || req.Amount <= 0 {
		respondWithError(w, "Invalid parameters", http.StatusBadRequest)
		return
	}

	quotes := h.aggregator.GetAllQuotes(r.Context(), req.From, req.To, req.Amount)
	respondWithJSON(w, quotes, http.StatusOK)
}

// GetMinAmounts obtiene los montos mínimos
func (h *QuoteHandler) GetMinAmounts(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	if from == "" || to == "" {
		respondWithError(w, "Missing parameters", http.StatusBadRequest)
		return
	}

	minAmounts := h.aggregator.GetMinAmounts(r.Context(), from, to)
	respondWithJSON(w, minAmounts, http.StatusOK)
}
