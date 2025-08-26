package main

import (
	"html/template"
	"net/http"
	"strings"
)

type Currency struct{ Code, Name string }
type PageData struct {
	Query      string
	Currencies []Currency
}

var currencies = []Currency{
	{"USD", "US Dollar"}, {"EUR", "Euro"}, {"JPY", "Japanese Yen"},
	{"GBP", "Pound Sterling"}, {"AUD", "Australian Dollar"},
	{"CAD", "Canadian Dollar"}, {"CHF", "Swiss Franc"},
	{"CNY", "Chinese Yuan"}, {"HKD", "Hong Kong Dollar"},
	{"NZD", "New Zealand Dollar"}, {"SEK", "Swedish Krona"},
	{"NOK", "Norwegian Krone"}, {"DKK", "Danish Krone"},
	{"PLN", "Polish Złoty"}, {"MXN", "Mexican Peso"},
}

var page = template.Must(template.New("page").Parse(`
<!doctype html>
<meta charset="utf-8">
<title>Currency Filter (HTMX + Go)</title>
<script src="https://unpkg.com/htmx.org@1.9.12"></script>
<style>
  body { font-family: system-ui, sans-serif; max-width: 44rem; margin: 2rem auto; }
  .field { width: 100%; padding: .6rem .8rem; border: 1px solid #ccc; border-radius: .5rem; }
  ul { list-style: none; padding-left: 0; margin: .5rem 0 0; }
  li { padding: .5rem .6rem; border-bottom: 1px solid #eee; display: flex; gap: .5rem; align-items: baseline; }
  code { font-weight: 600; }
  .muted { color: #666; }
</style>

<h1>Pick a currency</h1>

<input
  id="currency-query"
  class="field"
  type="text"
  name="q"
  placeholder="Start typing: e.g. 'eur' or 'dollar'"
  autocomplete="off"
  hx-get="/currencies"
  hx-trigger="keyup changed delay:250ms"
  hx-target="#currency-list"
  hx-select="#currency-list"
/>

<!-- Initial list render; subsequent updates replace this UL only -->
<ul id="currency-list" role="listbox" aria-label="Currencies">
  {{range .Currencies}}
    <li role="option">
      <code>{{.Code}}</code>
      <span class="muted">— {{.Name}}</span>
    </li>
  {{else}}
    <li class="muted">No matches.</li>
  {{end}}
</ul>
`))

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_ = page.Execute(w, PageData{Currencies: currencies})
	})

	http.HandleFunc("/currencies", func(w http.ResponseWriter, r *http.Request) {
		q := strings.TrimSpace(r.URL.Query().Get("q"))
		data := PageData{Query: q, Currencies: filterCurrencies(q)}
		// Return the full page; hx-select will extract only #currency-list
		if err := page.Execute(w, data); err != nil {
			http.Error(w, err.Error(), 500)
		}
	})

	http.ListenAndServe(":8080", nil)
}

func filterCurrencies(q string) []Currency {
	if q == "" {
		return currencies
	}
	q = strings.ToLower(q)
	var out []Currency
	for _, c := range currencies {
		if strings.Contains(strings.ToLower(c.Code), q) || strings.Contains(strings.ToLower(c.Name), q) {
			out = append(out, c)
		}
	}
	return out
}
