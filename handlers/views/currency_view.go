package views

import (
	"cryptoswap/models"
	"cryptoswap/services"
	"fmt"
	"net/http"
	"strings"
)

type CurrencyViewController struct {
	aggregator *services.Aggregator
}

func NewCurrencyViewController(aggregator *services.Aggregator) *CurrencyViewController {
	return &CurrencyViewController{
		aggregator: aggregator,
	}
}

// RenderCurrencyList renderiza las monedas como opciones de select para HTMX
func (vc *CurrencyViewController) RenderCurrencyList(w http.ResponseWriter, r *http.Request) {
	popular, others, err := vc.aggregator.GetAllCurrencies()
	if err != nil {
		w.Write([]byte(`<option value="">Error loading currencies</option>`))
		return
	}

	vc.renderSelectOptions(w, popular, others)
}

// renderSelectOptions renderiza las monedas como opciones HTML
func (vc *CurrencyViewController) renderSelectOptions(w http.ResponseWriter,
	popular, others []models.Currency) {
	html := vc.getOptionHTML("Popular", popular)
	html += vc.getOptionHTML("All Currencies", others)
	w.Write([]byte(html))
}

func (vc *CurrencyViewController) getOptionHTML(label string,
	currencies []models.Currency) string {
	html := fmt.Sprintf(`<optgroup label="%s">`, label)
	for _, curr := range currencies {
		html += fmt.Sprintf(`<option value="%s">%s - (%s)</option>`,
			curr.GetLowerSymbol(), curr.GetUpperSymbol(), curr.Network)
	}
	html += `</optgroup>`
	return html
}

// SearchCurrencies busca monedas por término (útil para autocomplete)
func (vc *CurrencyViewController) SearchCurrencies(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("q")
	if query == "" {
		w.Write([]byte(""))
		return
	}

	popular, others, err := vc.aggregator.GetAllCurrencies()
	if err != nil {
		w.Write([]byte(""))
		return
	}

	// Combinar y filtrar
	allCurrencies := append(popular, others...)
	query = strings.ToLower(query)

	html := ""
	count := 0
	for _, curr := range allCurrencies {
		if count >= 10 { // Limitar resultados
			break
		}

		symbol := strings.ToLower(curr.Symbol)
		name := strings.ToLower(curr.Name)

		if strings.Contains(symbol, query) || strings.Contains(name, query) {
			html += fmt.Sprintf(`
            <div onclick="selectCurrency('%s')" style="padding: 8px; cursor: pointer; hover: background: rgba(255,255,255,0.05);">
                <strong>%s</strong> - %s
            </div>`, symbol, strings.ToUpper(curr.Symbol), curr.Name)
			count++
		}
	}

	if html == "" {
		html = `<div style="padding: 8px; color: #666;">No currencies found</div>`
	}

	w.Write([]byte(html))
}
