package handlers

import (
	"cryptoswap/models"
	"cryptoswap/services"
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
)

type QuoteHandler struct {
	aggregator *services.Aggregator
}

func NewQuoteHandler(aggregator *services.Aggregator) *QuoteHandler {
	return &QuoteHandler{
		aggregator: aggregator,
	}
}

// GetQuote maneja las solicitudes de cotización
func (h *QuoteHandler) GetQuote(w http.ResponseWriter, r *http.Request) {
	// Parsear parámetros del form o query
	r.ParseForm()
	
	from := r.FormValue("from")
	to := r.FormValue("to")
	amountStr := r.FormValue("amount")
	
	// Validar parámetros
	if from == "" || to == "" || amountStr == "" {
		// Si es HTMX, devolver vacío
		if r.Header.Get("HX-Request") == "true" {
			w.Write([]byte(""))
			return
		}
		http.Error(w, "Missing parameters", http.StatusBadRequest)
		return
	}
	
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		if r.Header.Get("HX-Request") == "true" {
			w.Write([]byte(""))
			return
		}
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}
	
	// Obtener todas las cotizaciones
	quotes := h.aggregator.GetAllQuotes(from, to, amount)
	
	if len(quotes) == 0 {
		if r.Header.Get("HX-Request") == "true" {
			h.renderNoQuotes(w)
			return
		}
		http.Error(w, "No quotes available", http.StatusNotFound)
		return
	}
	
	// Si es una petición HTMX, devolver HTML
	if r.Header.Get("HX-Request") == "true" {
		h.renderHTMLQuotes(w, quotes, from, to)
		return
	}
	
	// Si no, devolver JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(quotes)
}

// GetAllQuotes devuelve todas las cotizaciones en JSON
func (h *QuoteHandler) GetAllQuotes(w http.ResponseWriter, r *http.Request) {
	var req models.QuoteRequest
	
	// Decodificar JSON request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	
	// Validar
	if req.From == "" || req.To == "" || req.Amount <= 0 {
		http.Error(w, "Invalid parameters", http.StatusBadRequest)
		return
	}
	
	// Obtener cotizaciones
	quotes := h.aggregator.GetAllQuotes(req.From, req.To, req.Amount)
	
	// Devolver JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(quotes)
}

// renderHTMLQuotes renderiza las cotizaciones como HTML para HTMX
func (h *QuoteHandler) renderHTMLQuotes(w http.ResponseWriter, quotes []*models.Quote, from, to string) {
	tmpl := `
	<div style="background: rgba(15, 23, 42, 0.3); border-radius: 12px; padding: 12px; margin-bottom: 16px;">
		{{if .BestQuote}}
		<div style="display: flex; justify-content: space-between; margin-bottom: 8px;">
			<span style="color: #94a3b8;">Best Rate</span>
			<span style="color: #10b981; font-weight: 600;">{{.BestQuote.Exchange}}</span>
		</div>
		<div style="display: flex; justify-content: space-between; margin-bottom: 8px;">
			<span style="color: #94a3b8;">You receive</span>
			<span style="color: white; font-size: 18px; font-weight: 600;">
				{{printf "%.6f" .BestQuote.ToAmount}} {{.BestQuote.To}}
			</span>
		</div>
		<div style="display: flex; justify-content: space-between;">
			<span style="color: #94a3b8;">Rate</span>
			<span style="color: white;">
				1 {{.BestQuote.From}} = {{printf "%.6f" .BestQuote.Rate}} {{.BestQuote.To}}
			</span>
		</div>
		{{end}}
	</div>
	
	{{if gt (len .AllQuotes) 1}}
	<details style="margin-top: 12px;">
		<summary style="cursor: pointer; color: #94a3b8; font-size: 14px;">
			Compare all {{len .AllQuotes}} exchanges
		</summary>
		<div style="margin-top: 8px;">
			{{range .AllQuotes}}
			<div style="background: rgba(15, 23, 42, 0.2); border-radius: 8px; padding: 10px; margin-bottom: 8px; cursor: pointer; transition: all 0.2s;"
				 onclick="selectExchange('{{.Exchange}}')"
				 onmouseover="this.style.background='rgba(15, 23, 42, 0.4)'"
				 onmouseout="this.style.background='rgba(15, 23, 42, 0.2)'">
				<div style="display: flex; justify-content: space-between; align-items: center;">
					<div>
						<span style="color: white; font-weight: 500;">{{.Exchange}}</span>
						<div style="color: #64748b; font-size: 12px; margin-top: 2px;">
							Rate: {{printf "%.6f" .Rate}}
						</div>
					</div>
					<div style="text-align: right;">
						<span style="color: white; font-size: 16px; font-weight: 600;">
							{{printf "%.6f" .ToAmount}} {{.To}}
						</span>
						{{if eq .Exchange $.BestQuote.Exchange}}
						<div style="color: #10b981; font-size: 11px; margin-top: 2px;">BEST RATE</div>
						{{end}}
					</div>
				</div>
			</div>
			{{end}}
		</div>
	</details>
	{{end}}
	
	<script>
		// Actualizar el campo de cantidad recibida
		document.getElementById('toAmount').value = '{{printf "%.6f" .BestQuote.ToAmount}}';
		// Habilitar el botón de swap
		document.getElementById('swapButton').disabled = false;
		document.getElementById('swapButton').textContent = 'Swap via {{.BestQuote.Exchange}}';
		// Guardar el mejor exchange
		document.getElementById('swapButton').setAttribute('data-exchange', '{{.BestQuote.Exchange}}');
		
		// Función para seleccionar un exchange específico
		function selectExchange(exchange) {
			document.getElementById('swapButton').textContent = 'Swap via ' + exchange;
			document.getElementById('swapButton').setAttribute('data-exchange', exchange);
		}
	</script>`
	
	t, err := template.New("quotes").Parse(tmpl)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	
	data := struct {
		BestQuote *models.Quote
		AllQuotes []*models.Quote
	}{
		BestQuote: quotes[0],
		AllQuotes: quotes,
	}
	
	t.Execute(w, data)
}

// renderNoQuotes renderiza mensaje cuando no hay cotizaciones
func (h *QuoteHandler) renderNoQuotes(w http.ResponseWriter) {
	html := `
	<div style="background: rgba(239, 68, 68, 0.1); border: 1px solid rgba(239, 68, 68, 0.3); 
	            border-radius: 12px; padding: 12px; color: #f87171;">
		<div style="font-weight: 600; margin-bottom: 4px;">No quotes available</div>
		<div style="font-size: 14px; color: #fca5a5;">
			This pair might not be supported or all exchanges are experiencing issues.
		</div>
	</div>
	<script>
		document.getElementById('toAmount').value = '';
		document.getElementById('swapButton').disabled = true;
		document.getElementById('swapButton').textContent = 'No quotes available';
	</script>`
	
	w.Write([]byte(html))
}

// GetMinAmounts obtiene los montos mínimos para un par
func (h *QuoteHandler) GetMinAmounts(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	
	if from == "" || to == "" {
		http.Error(w, "Missing parameters", http.StatusBadRequest)
		return
	}
	
	minAmounts := h.aggregator.GetMinAmounts(from, to)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(minAmounts)
}