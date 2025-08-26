package views

import (
	"cryptoswap/internal/services/swap"
	"html/template"
	"net/http"
	"strings"
)

type PageViewController struct {
	templates  *template.Template
	aggregator *swap.Aggregator
}

func NewPageViewController(aggregator *swap.Aggregator) *PageViewController {
	// Cargar todos los templates
	tmpl := template.Must(template.ParseGlob("templates/layouts/*.html"))
	tmpl = template.Must(tmpl.ParseGlob("templates/components/*.html"))
	tmpl = template.Must(tmpl.ParseGlob("templates/pages/*.html"))

	return &PageViewController{
		templates:  tmpl,
		aggregator: aggregator,
	}
}

func (vc *PageViewController) RenderIndex(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Obtener datos dinámicos
	exchanges := vc.aggregator.GetExchanges(ctx)
	popular, others, _ := vc.aggregator.GetAllCurrencies(ctx)

	data := PageData{
		Title:         "Siphon - Privacy-First Crypto Exchange",
		ExchangeCount: len(exchanges),
		ExchangeNames: strings.Join(exchanges, ", "),
		PairCount:     len(popular) + len(others),
	}

	err := vc.templates.ExecuteTemplate(w, "index", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type PageData struct {
	Title         string
	ExchangeCount int
	ExchangeNames string
	PairCount     int
}
