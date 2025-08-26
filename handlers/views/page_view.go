package views

import (
	"cryptoswap/services"
	"html/template"
	"net/http"
	"strings"
)

type PageViewController struct {
	templates  *template.Template
	aggregator *services.Aggregator
}

func NewPageViewController(aggregator *services.Aggregator) *PageViewController {
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
	// Obtener datos dinámicos
	exchanges := vc.aggregator.GetExchanges()
	popular, others, _ := vc.aggregator.GetAllCurrencies()

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
