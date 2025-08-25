package handlers

import (
	"cryptoswap/models"
	"cryptoswap/services"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type SwapHandler struct {
	aggregator *services.Aggregator
}

func NewSwapHandler(aggregator *services.Aggregator) *SwapHandler {
	return &SwapHandler{
		aggregator: aggregator,
	}
}

// CreateSwap crea un nuevo intercambio
func (h *SwapHandler) CreateSwap(w http.ResponseWriter, r *http.Request) {
	var req models.SwapRequest

	// Si es form data (desde HTMX)
	if r.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
		r.ParseForm()

		amount := 0.0
		fmt.Sscanf(r.FormValue("amount"), "%f", &amount)

		req = models.SwapRequest{
			From:          r.FormValue("from"),
			To:            r.FormValue("to"),
			Amount:        amount,
			ToAddress:     r.FormValue("to_address"),
			RefundAddress: r.FormValue("refund_address"),
			Exchange:      r.FormValue("exchange"),
		}
	} else {
		// Si es JSON
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
	}

	// Validar request
	if req.From == "" || req.To == "" || req.Amount <= 0 || req.ToAddress == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Si no se especificó exchange, usar el mejor
	if req.Exchange == "" {
		quote, err := h.aggregator.GetBestQuote(req.From, req.To, req.Amount)
		if err != nil || quote == nil {
			http.Error(w, "No exchange available", http.StatusBadRequest)
			return
		}
		req.Exchange = quote.Exchange
	}

	// Crear el intercambio
	swap, err := h.aggregator.CreateExchange(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating swap: %v", err), http.StatusInternalServerError)
		return
	}

	// Si es HTMX, devolver HTML de confirmación
	if r.Header.Get("HX-Request") == "true" {
		h.renderSwapCreated(w, swap)
		return
	}

	// Si no, devolver JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(swap)
}

// GetStatus obtiene el estado de un intercambio
func (h *SwapHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	swapID := vars["id"]

	if swapID == "" {
		http.Error(w, "Swap ID required", http.StatusBadRequest)
		return
	}

	// Por ahora devolvemos un estado mock
	// En producción, esto consultaría el exchange correspondiente
	status := map[string]interface{}{
		"id":      swapID,
		"status":  "waiting", // waiting, confirming, exchanging, sending, finished, failed
		"message": "Waiting for deposit",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// renderSwapCreated renderiza la confirmación de swap creado
func (h *SwapHandler) renderSwapCreated(w http.ResponseWriter, swap *models.SwapResponse) {
	html := fmt.Sprintf(`
	<div style="background: rgba(16, 185, 129, 0.1); border: 1px solid rgba(16, 185, 129, 0.3);
	            border-radius: 16px; padding: 20px; text-align: center;">
		<div style="width: 60px; height: 60px; margin: 0 auto 16px;
		            background: linear-gradient(135deg, #10b981, #059669);
		            border-radius: 50%%; display: flex; align-items: center; justify-content: center;">
			<svg style="width: 32px; height: 32px; color: white;" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7"/>
			</svg>
		</div>

		<h3 style="color: white; font-size: 20px; margin-bottom: 8px;">Swap Created!</h3>
		<p style="color: #94a3b8; margin-bottom: 20px;">Send your funds to the address below</p>

		<div style="background: rgba(15, 23, 42, 0.5); border-radius: 12px; padding: 16px; margin-bottom: 16px;">
			<div style="color: #64748b; font-size: 12px; margin-bottom: 4px;">Deposit Address</div>
			<div style="color: white; font-family: monospace; font-size: 14px; word-break: break-all;
			            background: rgba(0,0,0,0.3); padding: 8px; border-radius: 6px; margin-top: 8px;">
				%s
			</div>
			<button onclick="navigator.clipboard.writeText('%s')"
			        style="margin-top: 8px; padding: 6px 12px; background: rgba(59, 130, 246, 0.2);
			               color: #60a5fa; border: 1px solid rgba(59, 130, 246, 0.3);
			               border-radius: 6px; cursor: pointer; font-size: 12px;">
				Copy Address
			</button>
		</div>

		<div style="display: grid; grid-template-columns: 1fr 1fr; gap: 12px; margin-bottom: 16px;">
			<div style="background: rgba(15, 23, 42, 0.3); border-radius: 8px; padding: 12px;">
				<div style="color: #64748b; font-size: 12px;">You Send</div>
				<div style="color: white; font-weight: 600;">%.6f %s</div>
			</div>
			<div style="background: rgba(15, 23, 42, 0.3); border-radius: 8px; padding: 12px;">
				<div style="color: #64748b; font-size: 12px;">You Receive</div>
				<div style="color: white; font-weight: 600;">~%.6f %s</div>
			</div>
		</div>

		<div style="background: rgba(251, 146, 60, 0.1); border: 1px solid rgba(251, 146, 60, 0.3);
		            border-radius: 8px; padding: 12px; margin-bottom: 16px;">
			<div style="display: flex; align-items: center; justify-content: center; gap: 8px;">
				<svg style="width: 20px; height: 20px; color: #fb923c;" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
					      d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
				</svg>
				<span style="color: #fed7aa; font-size: 14px;">
					Send only %s to this address!
				</span>
			</div>
		</div>

		<div style="display: flex; gap: 12px;">
			<button onclick="checkSwapStatus('%s')"
			        style="flex: 1; padding: 12px; background: rgba(59, 130, 246, 0.2);
			               color: #60a5fa; border: 1px solid rgba(59, 130, 246, 0.3);
			               border-radius: 8px; cursor: pointer;">
				Check Status
			</button>
			<button onclick="location.reload()"
			        style="flex: 1; padding: 12px; background: linear-gradient(135deg, #a855f7, #6366f1);
			               color: white; border: none; border-radius: 8px; cursor: pointer;">
				New Swap
			</button>
		</div>

		<div style="margin-top: 16px; padding-top: 16px; border-top: 1px solid rgba(255, 255, 255, 0.05);">
			<p style="color: #64748b; font-size: 11px;">
				Exchange ID: <span style="font-family: monospace;">%s</span><br>
				Powered by %s
			</p>
		</div>
	</div>

	<script>
		function checkSwapStatus(id) {
			fetch('/api/swap/' + id + '/status')
				.then(r => r.json())
				.then(data => {
					alert('Status: ' + data.status + '\\n' + data.message);
				});
		}
	</script>`,
		swap.PayinAddress,
		swap.PayinAddress,
		swap.PayinAmount,
		strings.ToUpper(swap.From),
		swap.PayoutAmount,
		strings.ToUpper(swap.To),
		strings.ToUpper(swap.From),
		swap.ID,
		swap.ID,
		swap.Exchange,
	)

	w.Write([]byte(html))
}
