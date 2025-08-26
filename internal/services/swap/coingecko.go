package swap

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"cryptoswap/internal/services/models"
)

type CoinGeckoService struct {
	baseURL string
	apiKey  string
	httpc   *http.Client

	mu    sync.Mutex
	cache map[string]cacheItem
}

type cacheItem struct {
	val     any
	expires time.Time
}

func NewCoinGeckoService(baseURL, apiKey string) *CoinGeckoService {
	if baseURL == "" {
		baseURL = "https://api.coingecko.com/api/v3" // Free
	}
	return &CoinGeckoService{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		httpc:   &http.Client{Timeout: 10 * time.Second},
		cache:   make(map[string]cacheItem),
	}
}

func (s *CoinGeckoService) getCached(key string) (any, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	it, ok := s.cache[key]
	if !ok || time.Now().After(it.expires) {
		return nil, false
	}
	return it.val, true
}
func (s *CoinGeckoService) setCached(key string, val any, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache[key] = cacheItem{val: val, expires: time.Now().Add(ttl)}
}

type marketsCoin struct {
	Symbol                   string   `json:"symbol"`
	CurrentPrice             float64  `json:"current_price"`
	PriceChangePercentage24h *float64 `json:"price_change_percentage_24h"`
}

func (s *CoinGeckoService) authHeaderName() string {
	// Si apuntas al endpoint Pro, usa header Pro; si no, el de la capa gratuita
	if strings.Contains(strings.ToLower(s.baseURL), "pro-api.coingecko.com") {
		return "x-cg-pro-api-key"
	}
	return "x-cg-demo-api-key"
}

// TopTickers devuelve top N por market cap con precio y % 24h.
func (s *CoinGeckoService) TopTickers(ctx context.Context, vs string, n int) ([]models.Ticker, error) {
	if vs == "" {
		vs = "usd"
	}
	if n <= 0 || n > 250 {
		n = 6
	}

	cacheKey := "cg:top:" + strings.ToLower(vs) + ":" + strconv.Itoa(n)
	if v, ok := s.getCached(cacheKey); ok {
		return v.([]models.Ticker), nil
	}

	u, _ := url.Parse(s.baseURL + "/coins/markets")
	q := u.Query()
	q.Set("vs_currency", strings.ToLower(vs))
	q.Set("order", "market_cap_desc")
	q.Set("per_page", strconv.Itoa(n))
	q.Set("page", "1")
	q.Set("price_change_percentage", "24h")
	u.RawQuery = q.Encode()

	req, _ := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	req.Header.Set("accept", "application/json")

	// Siempre envía algún header válido: demo o tu key
	headerName := s.authHeaderName()
	key := s.apiKey
	if key == "" {
		key = "DEMO-API-KEY" // muy rate limited, pero evita 400/401
	}
	req.Header.Set(headerName, key)

	resp, err := s.httpc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("coingecko error: %s - %s", resp.Status, string(body))
	}

	var raw []marketsCoin
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	out := make([]models.Ticker, 0, len(raw))
	for _, c := range raw {
		ch := 0.0
		if c.PriceChangePercentage24h != nil {
			ch = *c.PriceChangePercentage24h
		}
		out = append(out, models.Ticker{
			Symbol: strings.ToUpper(c.Symbol),
			Price:  c.CurrentPrice,
			Change: ch, // mapea a json:"change_24h" si tu modelo lo usa así
		})
	}

	s.setCached(cacheKey, out, 15*time.Second)
	return out, nil
}
