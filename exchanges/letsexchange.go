package exchanges

import (
	"bytes"
	"cryptoswap/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type LetsExchange struct {
	apiKey  string
	baseURL string
	client  *http.Client

	// cache sencillo de redes por defecto para cada coin (p.ej. USDT -> TRC20)
	mu          sync.RWMutex
	defaultNets map[string]string // CODE -> default_network_code
}

// NewLetsExchange crea el cliente para LetsExchange
func NewLetsExchange(apiKey string) *LetsExchange {
	return &LetsExchange{
		apiKey:  apiKey,
		baseURL: "https://api.letsexchange.io/api", // base correcta
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
		defaultNets: make(map[string]string),
	}
}

func (l *LetsExchange) GetName() string {
	return "LetsExchange"
}

// ===== Helpers =====

func (l *LetsExchange) auth(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+l.apiKey)
	req.Header.Set("Accept", "application/json")
}

func (l *LetsExchange) ensureDefaultNetworksLoaded() error {
	l.mu.RLock()
	loaded := len(l.defaultNets) > 0
	l.mu.RUnlock()
	if loaded {
		return nil
	}

	// Cargar mapa CODE -> default_network_code desde /v2/coins
	type coin struct {
		Code               string `json:"code"`
		DefaultNetworkCode string `json:"default_network_code"`
		Networks           []struct {
			Code     string `json:"code"`
			IsActive int    `json:"is_active"`
		} `json:"networks"`
		IsActive int `json:"is_active"`
	}

	url := l.baseURL + "/v2/coins"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	l.auth(req)

	resp, err := l.client.Do(req)
	if err != nil {
		return fmt.Errorf("error fetching coins: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var coins []coin
	if err := json.NewDecoder(resp.Body).Decode(&coins); err != nil {
		return fmt.Errorf("error decoding coins: %w", err)
	}

	tmp := make(map[string]string, len(coins))
	for _, c := range coins {
		code := strings.ToUpper(c.Code)
		net := c.DefaultNetworkCode
		if net == "" {
			// Fallback: si no hay default, usa la primera red activa si existe
			for _, n := range c.Networks {
				if n.IsActive == 1 {
					net = n.Code
					break
				}
			}
		}
		if net == "" {
			// último fallback: usa el propio code (BTC -> BTC)
			net = code
		}
		tmp[code] = net
	}

	l.mu.Lock()
	l.defaultNets = tmp
	l.mu.Unlock()

	return nil
}

func (l *LetsExchange) defaultNet(code string) string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if n, ok := l.defaultNets[strings.ToUpper(code)]; ok && n != "" {
		return n
	}
	// fallback simple
	return strings.ToUpper(code)
}

// ===== API =====

// GetCurrencies obtiene todas las monedas disponibles (1 entrada por red)
func (l *LetsExchange) GetCurrencies() ([]models.Currency, error) {
	type coin struct {
		Code               string `json:"code"`
		Name               string `json:"name"`
		IsActive           int    `json:"is_active"`
		DefaultNetworkCode string `json:"default_network_code"`
		Icon               string `json:"icon"`
		Networks           []struct {
			Name     string `json:"name"`
			Code     string `json:"code"`
			IsActive int    `json:"is_active"`
		} `json:"networks"`
	}

	url := l.baseURL + "/v2/coins"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	l.auth(req)

	resp, err := l.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching coins: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var coins []coin
	if err := json.NewDecoder(resp.Body).Decode(&coins); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	out := make([]models.Currency, 0, len(coins))
	for _, c := range coins {
		for _, n := range c.Networks {
			available := (c.IsActive == 1 && n.IsActive == 1)
			out = append(out, models.Currency{
				Symbol:    strings.ToUpper(c.Code),
				Name:      ifEmpty(c.Name, c.Code),
				Image:     c.Icon,
				Network:   n.Code,
				Available: available,
			}.WithProvider(l.GetName()))
		}
	}

	// Guarda default networks para usarlas en GetQuote/CreateExchange
	l.mu.Lock()
	tmp := make(map[string]string, len(coins))
	for _, c := range coins {
		net := c.DefaultNetworkCode
		if net == "" {
			// fallback si no trae default
			for _, n := range c.Networks {
				if n.IsActive == 1 {
					net = n.Code
					break
				}
			}
		}
		if net == "" {
			net = strings.ToUpper(c.Code)
		}
		tmp[strings.ToUpper(c.Code)] = net
	}
	l.defaultNets = tmp
	l.mu.Unlock()

	return out, nil
}

// GetQuote obtiene una cotización (rate + cantidades) vía POST /v1/info
func (l *LetsExchange) GetQuote(from, to string, amount float64) (*models.Quote, error) {
	// Asegura que tenemos redes por defecto
	if err := l.ensureDefaultNetworksLoaded(); err != nil {
		// no falles duro; pero si no cargan, seguirá con fallback simple
	}

	netFrom := l.defaultNet(from)
	netTo := l.defaultNet(to)

	url := l.baseURL + "/v1/info"
	body := map[string]any{
		"from":         strings.ToUpper(from),
		"to":           strings.ToUpper(to),
		"network_from": netFrom,
		"network_to":   netTo,
		"amount":       amount,
		"float":        true, // market rate
	}
	jsonBody, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	l.auth(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := l.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching quote: %w", err)
	}
	defer resp.Body.Close()

	// Si el par no existe / inválido, devuelve nil para que el agregador continúe con otros
	if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusUnprocessableEntity {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		bodyB, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(bodyB))
	}

	var result struct {
		MinAmount     string `json:"min_amount"`
		MaxAmount     string `json:"max_amount"`
		Amount        string `json:"amount"` // amount the user receives (to)
		Fee           string `json:"fee"`
		Rate          string `json:"rate"`
		WithdrawalFee string `json:"withdrawal_fee"`
		RateID        string `json:"rate_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	toAmount, _ := strconv.ParseFloat(result.Amount, 64)
	rate, _ := strconv.ParseFloat(result.Rate, 64)
	minAmt, _ := strconv.ParseFloat(result.MinAmount, 64)
	maxAmt, _ := strconv.ParseFloat(result.MaxAmount, 64)

	// Sin resultado útil
	if toAmount <= 0 || rate <= 0 {
		return nil, nil
	}

	return &models.Quote{
		Exchange:   l.GetName(),
		From:       strings.ToUpper(from),
		To:         strings.ToUpper(to),
		FromAmount: amount,
		ToAmount:   toAmount,
		Rate:       rate,
		MinAmount:  minAmt,
		MaxAmount:  maxAmt,
		Timestamp:  time.Now(),
	}, nil
}

// GetMinAmount usa /v1/info y devuelve el min_amount para el par
func (l *LetsExchange) GetMinAmount(from, to string) (float64, error) {
	// Asegura redes por defecto
	if err := l.ensureDefaultNetworksLoaded(); err != nil {
		// continúa con fallback
	}
	netFrom := l.defaultNet(from)
	netTo := l.defaultNet(to)

	url := l.baseURL + "/v1/info"
	body := map[string]any{
		"from":         strings.ToUpper(from),
		"to":           strings.ToUpper(to),
		"network_from": netFrom,
		"network_to":   netTo,
		"amount":       1,    // cualquier valor; el endpoint devuelve min/max igualmente
		"float":        true, // market
	}
	jsonBody, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return 0, fmt.Errorf("error creating request: %w", err)
	}
	l.auth(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := l.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("error fetching min amount: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusUnprocessableEntity {
		return 0, nil // par inválido → sin mínimo
	}
	if resp.StatusCode != http.StatusOK {
		bodyB, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(bodyB))
	}

	var result struct {
		MinAmount string `json:"min_amount"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("error decoding response: %w", err)
	}
	minAmt, _ := strconv.ParseFloat(result.MinAmount, 64)
	return minAmt, nil
}

// CreateExchange crea un intercambio real vía POST /v1/transaction
func (l *LetsExchange) CreateExchange(req models.SwapRequest) (*models.SwapResponse, error) {
	// Asegura redes por defecto
	if err := l.ensureDefaultNetworksLoaded(); err != nil {
		// continúa con fallback
	}
	netFrom := l.defaultNet(req.From)
	netTo := l.defaultNet(req.To)

	url := l.baseURL + "/v1/transaction"
	payload := map[string]any{
		"float":               true, // market rate
		"coin_from":           strings.ToUpper(req.From),
		"coin_to":             strings.ToUpper(req.To),
		"network_from":        netFrom,
		"network_to":          netTo,
		"deposit_amount":      req.Amount,
		"withdrawal":          req.ToAddress,     // dirección destino
		"withdrawal_extra_id": "",                // si aplica (XRP/MEMO, etc.)
		"return":              req.RefundAddress, // dirección de refund
		"return_extra_id":     "",
		// "rate_id": "", // para fixed rate (si antes llamas a /v1/info con float=false)
		// "affiliate_id": "tuAffiliateIdOpcional",
	}
	jsonBody, _ := json.Marshal(payload)

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	l.auth(httpReq)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := l.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error creating exchange: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var result struct {
		TransactionID     string  `json:"transaction_id"`
		Status            string  `json:"status"`
		CoinFrom          string  `json:"coin_from"`
		CoinTo            string  `json:"coin_to"`
		CoinFromNetwork   string  `json:"coin_from_network"`
		CoinToNetwork     string  `json:"coin_to_network"`
		DepositAmount     string  `json:"deposit_amount"`
		WithdrawalAmount  string  `json:"withdrawal_amount"`
		DepositAddress    string  `json:"deposit"`
		DepositExtraID    *string `json:"deposit_extra_id"`
		Withdrawal        string  `json:"withdrawal"`
		WithdrawalExtraID string  `json:"withdrawal_extra_id"`
		Rate              string  `json:"rate"`
		Fee               string  `json:"fee"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	payinAmount, _ := strconv.ParseFloat(result.DepositAmount, 64)
	payoutAmount, _ := strconv.ParseFloat(result.WithdrawalAmount, 64)

	return &models.SwapResponse{
		ID:            result.TransactionID,
		Status:        result.Status,
		From:          strings.ToUpper(req.From),
		To:            strings.ToUpper(req.To),
		PayinAddress:  result.DepositAddress,
		PayinAmount:   payinAmount,
		PayoutAmount:  payoutAmount,
		PayoutAddress: result.Withdrawal,
		Exchange:      l.GetName(),
		CreatedAt:     time.Now(), // el endpoint no devuelve timestamp explícito
	}, nil
}

// GetExchangeStatus obtiene el estado de un intercambio
func (l *LetsExchange) GetExchangeStatus(id string) (string, error) {
	url := l.baseURL + "/v1/transaction/" + id

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}
	l.auth(req)

	resp, err := l.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error fetching status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("error decoding response: %w", err)
	}

	return result.Status, nil
}

// ===== utils =====

func ifEmpty(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}
