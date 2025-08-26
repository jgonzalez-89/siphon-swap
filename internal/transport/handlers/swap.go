package handlers

import (
	"cryptoswap/internal/services/models"
	"cryptoswap/internal/services/swap"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type SwapHandler struct {
	aggregator *swap.Aggregator
}

func NewSwapHandler(aggregator *swap.Aggregator) *SwapHandler {
	return &SwapHandler{
		aggregator: aggregator,
	}
}

// CreateSwap crea un nuevo intercambio (JSON API)
func (h *SwapHandler) CreateSwap(w http.ResponseWriter, r *http.Request) {
	var req models.SwapRequest

	// Decodificar JSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validar request
	if req.From == "" || req.To == "" || req.Amount <= 0 || req.ToAddress == "" {
		respondWithError(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Si no se especificó exchange, usar el mejor
	if req.Exchange == "" {
		quote, err := h.aggregator.GetBestQuote(r.Context(), req.From, req.To, req.Amount)
		if err != nil || quote == nil {
			respondWithError(w, "No exchange available", http.StatusBadRequest)
			return
		}
		req.Exchange = quote.Exchange
	}

	// Crear el intercambio
	swap, err := h.aggregator.CreateExchange(req)
	if err != nil {
		respondWithError(w, fmt.Sprintf("Error creating swap: %v", err), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, swap, http.StatusCreated)
}

// GetStatus obtiene el estado de un intercambio
func (h *SwapHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	swapID := vars["id"]

	if swapID == "" {
		respondWithError(w, "Swap ID required", http.StatusBadRequest)
		return
	}

	// TODO: Implementar consulta real al exchange
	// Por ahora devolvemos un estado mock
	status := map[string]interface{}{
		"id":      swapID,
		"status":  "waiting", // waiting, confirming, exchanging, sending, finished, failed
		"message": "Waiting for deposit",
	}

	respondWithJSON(w, status, http.StatusOK)
}
