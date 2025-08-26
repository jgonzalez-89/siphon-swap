package views

import (
	"cryptoswap/internal/services/models"
	"cryptoswap/internal/services/swap"
	"fmt"
	"net/http"
	"strings"
)

type CurrencyViewController struct {
	aggregator *swap.Aggregator
}

func NewCurrencyViewController(aggregator *swap.Aggregator) *CurrencyViewController {
	return &CurrencyViewController{
		aggregator: aggregator,
	}
}

// RenderCurrencyList renderiza las monedas como opciones de select para HTMX
func (vc *CurrencyViewController) RenderCurrencyList(w http.ResponseWriter, r *http.Request) {
	popular, others, err := vc.aggregator.GetAllCurrencies(r.Context())
	if err != nil {
		w.Write([]byte(`<option value="">Error loading currencies</option>`))
		return
	}

	vc.renderSelectOptions(w, popular, others)
}

// RenderCurrencySelector renders a full currency selector component with HTMX filtering
func (vc *CurrencyViewController) RenderCurrencySelector(w http.ResponseWriter, r *http.Request) {
	popular, others, err := vc.aggregator.GetAllCurrencies(r.Context())
	if err != nil {
		w.Write([]byte(`<div class="error">Error loading currencies</div>`))
		return
	}

	// Get filter from query params
	filter := r.URL.Query().Get("filter")
	filter = strings.ToLower(filter)

	// Filter currencies if filter is provided
	var filteredPopular, filteredOthers []models.Currency
	if filter != "" {
		filteredPopular = vc.filterCurrencies(popular, filter)
		filteredOthers = vc.filterCurrencies(others, filter)
	} else {
		filteredPopular = popular
		filteredOthers = others
	}

	// Limit results for better performance
	if len(filteredPopular) > 20 {
		filteredPopular = filteredPopular[:20]
	}
	if len(filteredOthers) > 30 {
		filteredOthers = filteredOthers[:30]
	}

	// Get target from query params to determine which dropdown this is for
	target := r.URL.Query().Get("target")
	html := vc.renderCurrencySelectorHTML(filteredPopular, filteredOthers, filter, target)
	w.Write([]byte(html))
}

func (vc *CurrencyViewController) filterCurrencies(currencies []models.Currency, filter string) []models.Currency {
	var filtered []models.Currency
	for _, curr := range currencies {
		symbol := strings.ToLower(curr.Symbol)
		name := strings.ToLower(curr.Name)
		network := strings.ToLower(curr.Network)

		if strings.Contains(symbol, filter) ||
			strings.Contains(name, filter) ||
			strings.Contains(network, filter) {
			filtered = append(filtered, curr)
		}
	}
	return filtered
}

func (vc *CurrencyViewController) renderCurrencySelectorHTML(popular, others []models.Currency, currentFilter, target string) string {
	// Determine the target ID based on the target parameter
	targetID := "currency-selector-content"
	if target == "from" {
		targetID = "from-currency-selector-content"
	} else if target == "to" {
		targetID = "to-currency-selector-content"
	}

	html := fmt.Sprintf(`
	<div class="currency-selector" style="background: rgba(8, 10, 28, 0.95); border-radius: 12px; padding: 16px; border: 1px solid rgba(255, 255, 255, 0.1); box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);">
		<!-- Search Input -->
		<div style="margin-bottom: 16px;">
			<input type="text"
				   placeholder="Search currencies..."
				   value="%s"
				   style="width: 100%%; background: rgba(8, 10, 28, 0.6); border: 1px solid rgba(255, 255, 255, 0.1); border-radius: 8px; padding: 12px; color: white; font-size: 14px; outline: none;"
				   hx-get="/htmx/currencies/selector?target=%s"
				   hx-trigger="keyup changed delay:300ms"
				   hx-target="#%s"
				   hx-include="[name='filter']"
				   name="filter">
		</div>

		<!-- Currency List -->
		<div id="%s" style="max-height: 300px; overflow-y: auto;">
			<div style="display: flex; flex-direction: column; gap: 8px;">
	`, currentFilter, target, targetID, targetID)

	// Popular currencies section
	if len(popular) > 0 {
		html += `<div style="margin-bottom: 12px;">
					<div style="color: #94a3b8; font-size: 12px; font-weight: 600; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 8px;">Popular</div>`

		for _, curr := range popular {
			html += vc.renderCurrencyOption(curr, target)
		}
		html += `</div>`
	}

	// All currencies section
	if len(others) > 0 {
		html += `<div>
					<div style="color: #94a3b8; font-size: 12px; font-weight: 600; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 8px;">All Currencies</div>`

		for _, curr := range others {
			html += vc.renderCurrencyOption(curr, target)
		}
		html += `</div>`
	}

	// No results message
	if len(popular) == 0 && len(others) == 0 {
		html += fmt.Sprintf(`<div style="text-align: center; color: #64748b; padding: 20px;">
					No currencies found matching "%s"
				</div>`, currentFilter)
	}

	html += `
			</div>
		</div>
	</div>`

	return html
}

func (vc *CurrencyViewController) renderCurrencyOption(curr models.Currency, target string) string {
	return fmt.Sprintf(`
		<div class="currency-option"
			 style="display: flex; align-items: center; gap: 12px; padding: 12px; border-radius: 8px; cursor: pointer; transition: background 0.2s; hover: background: rgba(255, 255, 255, 0.05);"
			 onclick="selectCurrency('%s', '%s', '%s', '%s')">
			<div class="currency-icon" style="width: 32px; height: 32px; background: linear-gradient(135deg, #8b5cf6 0%%, #3b82f6 100%%); border-radius: 50%%; display: flex; align-items: center; justify-content: center; font-weight: 600; font-size: 12px; color: white;">
				%s
			</div>
			<div style="flex: 1;">
				<div style="font-weight: 600; color: white;">%s</div>
				<div style="font-size: 12px; color: #64748b;">%s</div>
			</div>
			<div style="font-size: 11px; color: #94a3b8; background: rgba(255, 255, 255, 0.05); padding: 4px 8px; border-radius: 4px;">
				%s
			</div>
		</div>
	`,
		curr.GetLowerSymbol(),
		curr.GetUpperSymbol(),
		curr.Name,
		target,
		strings.ToUpper(curr.Symbol[:1]),
		curr.GetUpperSymbol(),
		curr.Name,
		curr.Network)
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

	popular, others, err := vc.aggregator.GetAllCurrencies(r.Context())
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
