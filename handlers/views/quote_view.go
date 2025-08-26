package views

import (
	"cryptoswap/internal/lib/parser"
	"cryptoswap/internal/services/models"
	"cryptoswap/internal/services/swap"
	"fmt"
	"html/template"
	"net/http"
)

type QuoteViewController struct {
	aggregator *swap.Aggregator
	templates  *template.Template
}

func NewQuoteViewController(aggregator *swap.Aggregator) *QuoteViewController {
	// Definir funciones helper para los templates
	funcMap := template.FuncMap{
		"printf": fmt.Sprintf,
		"calculateDifference": func(a, b float64) float64 {
			if b == 0 {
				return 0
			}
			return ((b - a) / b) * 100
		},
		"calculateSavings": func(quotes []*models.Quote) float64 {
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

	// Por ahora usaremos templates inline, después los moveremos a archivos
	tmpl := template.New("quotes").Funcs(funcMap)

	return &QuoteViewController{
		aggregator: aggregator,
		templates:  tmpl,
	}
}

// RenderQuotes maneja las solicitudes HTMX para mostrar cotizaciones
func (vc *QuoteViewController) RenderQuotes(w http.ResponseWriter, r *http.Request) {
	// Parsear parámetros del form (HTMX envía form data)
	var quote models.QuoteRequest
	if err := parser.Unmarshal(r, &quote); err != nil {
		w.Write([]byte(""))
		return
	}

	// Obtener cotizaciones
	quotes := vc.aggregator.GetAllQuotes(r.Context(), quote.From, quote.To, quote.Amount)

	if len(quotes) == 0 {
		vc.renderNoQuotes(w)
		return
	}

	// Renderizar HTML
	vc.renderQuoteCards(w, quotes)
}

func (vc *QuoteViewController) renderQuoteCards(w http.ResponseWriter, quotes []*models.Quote) {
	// Por ahora mantenemos el HTML inline, después lo moveremos a templates
	html := vc.generateQuoteHTML(quotes)
	w.Write([]byte(html))
}

func (vc *QuoteViewController) renderNoQuotes(w http.ResponseWriter) {
	html := `
    <div class="alert alert-warning">
        <div class="alert-title">No quotes available</div>
        <div class="alert-description">
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

// generateQuoteHTML genera el HTML para las cotizaciones
// (temporalmente aquí, después lo moveremos a templates)
func (vc *QuoteViewController) generateQuoteHTML(quotes []*models.Quote) string {
	// Aquí puedes reutilizar el HTML que ya tienes en tu handler actual
	// Lo simplificaré para el ejemplo

	bestQuote := quotes[0]
	savings := vc.calculateSavingsPercent(quotes)

	html := fmt.Sprintf(`
    <div class="quote-result">
        <!-- Best Quote -->
        <div class="best-quote-card">
            <div class="quote-header">
                <span class="badge">Best Rate</span>
                <span class="exchange">%s</span>
            </div>
            <div class="quote-amount">
                %.8f %s
            </div>
            <div class="quote-rate">
                Rate: 1 %s = %.8f %s
            </div>
            <button onclick="selectExchange('%s')" class="btn-primary">
                Use This
            </button>
        </div>

        <!-- Comparison -->
        <div class="exchanges-comparison">
            <h4>Compare All Exchanges (%d available)</h4>
    `, bestQuote.Exchange, bestQuote.ToAmount, bestQuote.To,
		bestQuote.From, bestQuote.Rate, bestQuote.To,
		bestQuote.Exchange, len(quotes))

	for _, quote := range quotes {
		isBest := quote.Exchange == bestQuote.Exchange
		difference := ((bestQuote.ToAmount - quote.ToAmount) / bestQuote.ToAmount) * 100

		html += fmt.Sprintf(`
            <div class="exchange-option" onclick="selectExchange('%s')">
                <div>
                    <span class="exchange-name">%s</span>
                    %s
                </div>
                <div>
                    <span class="amount">%.8f %s</span>
                    %s
                </div>
            </div>
        `, quote.Exchange, quote.Exchange,
			ternary(isBest, `<span class="badge-best">BEST</span>`, ""),
			quote.ToAmount, quote.To,
			ternary(!isBest, fmt.Sprintf(`<span class="difference">-%.2f%%</span>`, difference), ""))
	}

	html += fmt.Sprintf(`
            <div class="savings-info">
                You save %.2f%% using Siphon
            </div>
        </div>
    </div>

    <script>
        document.getElementById('toAmount').value = '%.8f';
        var button = document.getElementById('swapButton');
        button.disabled = false;
        button.textContent = 'Swap via %s';
        button.setAttribute('data-exchange', '%s');

        function selectExchange(exchange) {
            button.textContent = 'Swap via ' + exchange;
            button.setAttribute('data-exchange', exchange);
        }
    </script>
    `, savings, bestQuote.ToAmount, bestQuote.Exchange, bestQuote.Exchange)

	return html
}

func (vc *QuoteViewController) calculateSavingsPercent(quotes []*models.Quote) float64 {
	if len(quotes) < 2 {
		return 0
	}
	worst := quotes[len(quotes)-1].ToAmount
	best := quotes[0].ToAmount
	if worst == 0 {
		return 0
	}
	return ((best - worst) / worst) * 100
}

func ternary(condition bool, ifTrue, ifFalse string) string {
	if condition {
		return ifTrue
	}
	return ifFalse
}
