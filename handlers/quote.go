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
	{{if .BestQuote}}
	<!-- Best Quote Card -->
	<div style="background: rgba(139, 92, 246, 0.1); border: 1px solid rgba(139, 92, 246, 0.2); border-radius: 12px; padding: 16px; margin-bottom: 12px;">
		<div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px;">
			<span style="color: #a78bfa; font-size: 12px; font-weight: 600; text-transform: uppercase;">Best Rate</span>
			<span style="background: linear-gradient(135deg, #8b5cf6 0%, #3b82f6 100%); -webkit-background-clip: text; -webkit-text-fill-color: transparent; font-weight: 600;">
				{{.BestQuote.Exchange}}
			</span>
		</div>
		<div style="display: flex; justify-content: space-between; align-items: center;">
			<div>
				<div style="color: white; font-size: 24px; font-weight: 700;">
					{{printf "%.8f" .BestQuote.ToAmount}} {{.BestQuote.To}}
				</div>
				<div style="color: #64748b; font-size: 12px; margin-top: 4px;">
					Rate: 1 {{.BestQuote.From}} = {{printf "%.8f" .BestQuote.Rate}} {{.BestQuote.To}}
				</div>
			</div>
			<div style="text-align: right;">
				<button onclick="selectExchange('{{.BestQuote.Exchange}}')" 
				        style="background: linear-gradient(135deg, #8b5cf6 0%, #3b82f6 100%); color: white; border: none; padding: 8px 16px; border-radius: 8px; font-size: 12px; font-weight: 600; cursor: pointer;">
					Use This
				</button>
			</div>
		</div>
	</div>
	{{end}}
	
	{{if gt (len .AllQuotes) 1}}
	<!-- All Exchanges Comparison -->
	<div style="background: rgba(8, 10, 28, 0.4); border-radius: 12px; padding: 16px;">
		<div style="color: #64748b; font-size: 12px; font-weight: 600; text-transform: uppercase; margin-bottom: 12px;">
			Compare All Exchanges ({{len .AllQuotes}} available)
		</div>
		
		{{range $index, $quote := .AllQuotes}}
		<div style="background: rgba(15, 23, 42, 0.4); border-radius: 8px; padding: 12px; margin-bottom: 8px; cursor: pointer; transition: all 0.2s; border: 1px solid {{if eq $quote.Exchange $.BestQuote.Exchange}}rgba(139, 92, 246, 0.3){{else}}rgba(255, 255, 255, 0.05){{end}};"
		     onclick="selectExchange('{{$quote.Exchange}}')"
		     onmouseover="this.style.background='rgba(15, 23, 42, 0.6)'; this.style.borderColor='rgba(139, 92, 246, 0.4)'"
		     onmouseout="this.style.background='rgba(15, 23, 42, 0.4)'; this.style.borderColor='{{if eq $quote.Exchange $.BestQuote.Exchange}}rgba(139, 92, 246, 0.3){{else}}rgba(255, 255, 255, 0.05){{end}}'">
			<div style="display: flex; justify-content: space-between; align-items: center;">
				<div>
					<div style="display: flex; align-items: center; gap: 8px; margin-bottom: 4px;">
						<span style="color: white; font-weight: 600; font-size: 14px;">{{$quote.Exchange}}</span>
						{{if eq $quote.Exchange $.BestQuote.Exchange}}
						<span style="background: linear-gradient(135deg, #8b5cf6 0%, #3b82f6 100%); color: white; font-size: 10px; padding: 2px 6px; border-radius: 4px; font-weight: 600;">
							BEST
						</span>
						{{end}}
					</div>
					<div style="color: #64748b; font-size: 12px;">
						Rate: {{printf "%.8f" $quote.Rate}} {{$quote.To}}/{{$quote.From}}
					</div>
				</div>
				<div style="text-align: right;">
					<div style="color: white; font-size: 16px; font-weight: 600;">
						{{printf "%.8f" $quote.ToAmount}} {{$quote.To}}
					</div>
					{{if ne $quote.Exchange $.BestQuote.Exchange}}
					<div style="color: #ef4444; font-size: 11px; margin-top: 2px;">
						-{{printf "%.2f" ($.Difference $quote.ToAmount $.BestQuote.ToAmount)}}%
					</div>
					{{end}}
				</div>
			</div>
		</div>
		{{end}}
		
		<div style="margin-top: 12px; padding: 12px; background: rgba(16, 185, 129, 0.1); border: 1px solid rgba(16, 185, 129, 0.2); border-radius: 8px;">
			<div style="display: flex; align-items: center; gap: 8px;">
				<svg style="width: 16px; height: 16px; color: #10b981;" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
				</svg>
				<span style="color: #10b981; font-size: 12px;">
					You save {{printf "%.2f" ($.SavingsPercent)}}% using Siphon vs single exchange
				</span>
			</div>
		</div>
	</div>
	{{else if .BestQuote}}
	<!-- Single Exchange Available -->
	<div style="background: rgba(251, 146, 60, 0.1); border: 1px solid rgba(251, 146, 60, 0.2); border-radius: 8px; padding: 12px;">
		<div style="display: flex; align-items: center; gap: 8px;">
			<svg style="width: 16px; height: 16px; color: #fb923c;" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
			</svg>
			<span style="color: #fed7aa; font-size: 12px;">
				Only 1 exchange supports this pair currently
			</span>
		</div>
	</div>
	{{end}}
	
	<script>
		// Update the "to" amount field
		document.getElementById('toAmount').value = '{{printf "%.8f" .BestQuote.ToAmount}}';
		// Enable swap button
		var button = document.getElementById('swapButton');
		button.disabled = false;
		var buttonText = button.querySelector('span') || button;
		if (buttonText.textContent) {
			buttonText.textContent = 'Swap via {{.BestQuote.Exchange}}';
		} else {
			button.textContent = 'Swap via {{.BestQuote.Exchange}}';
		}
		button.setAttribute('data-exchange', '{{.BestQuote.Exchange}}');
		
		// Function to select a specific exchange
		function selectExchange(exchange) {
			var button = document.getElementById('swapButton');
			var buttonText = button.querySelector('span') || button;
			if (buttonText.textContent) {
				buttonText.textContent = 'Swap via ' + exchange;
			} else {
				button.textContent = 'Swap via ' + exchange;
			}
			button.setAttribute('data-exchange', exchange);
			
			// Highlight selected exchange
			document.querySelectorAll('[onclick*="selectExchange"]').forEach(el => {
				el.style.borderColor = 'rgba(255, 255, 255, 0.05)';
			});
			event.currentTarget.style.borderColor = 'rgba(139, 92, 246, 0.5)';
		}
	</script>`
	
	// Create template with helper functions
	funcMap := template.FuncMap{
		"Difference": func(a, b float64) float64 {
			if b == 0 {
				return 0
			}
			return ((b - a) / b) * 100
		},
		"SavingsPercent": func() float64 {
			if len(quotes) < 2 {
				return 0
			}
			worst := quotes[len(quotes)-1].ToAmount
			best := quotes[0].ToAmount
			if worst == 0 {
				return 0
			}
			return ((best - worst) / worst) * 100
		},
	}
	
	t, err := template.New("quotes").Funcs(funcMap).Parse(tmpl)
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