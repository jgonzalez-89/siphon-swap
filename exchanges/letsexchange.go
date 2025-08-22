package exchanges

import (
	"bytes"
	"cryptoswap/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type LetsExchange struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

func NewLetsExchange(apiKey string) *LetsExchange {
	return &LetsExchange{
		apiKey:  apiKey,
		baseURL: "https://api.letsexchange.io/api/v1",
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (l *LetsExchange) GetName() string {
	return "LetsExchange"
}

// GetCurrencies obtiene todas las monedas disponibles
func (l *LetsExchange) GetCurrencies() ([]models.Currency, error) {
	url := fmt.Sprintf("%s/currencies?api_key=%s", l.baseURL, l.apiKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+l.apiKey)

	resp, err := l.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching currencies: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var apiCurrencies []struct {
		Symbol    string `json:"symbol"`
		Name      string `json:"name"`
		Image     string `json:"image"`
		Network   string `json:"network"`
		Available bool   `json:"available"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiCurrencies); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	currencies := make([]models.Currency, 0, len(apiCurrencies))
	for _, curr := range apiCurrencies {
		if curr.Available {
			currencies = append(currencies, models.Currency{
				Symbol:    curr.Symbol,
				Name:      curr.Name,
				Image:     curr.Image,
				Network:   curr.Network,
				Available: true,
			})
		}
	}

	return currencies, nil
}

// GetQuote obtiene una cotización
func (l *LetsExchange) GetQuote(from, to string, amount float64) (*models.Quote, error) {
	url := fmt.Sprintf("%s/estimate?api_key=%s", l.baseURL, l.apiKey)

	// Preparar el request body
	quoteReq := map[string]interface{}{
		"from_currency": from,
		"to_currency":   to,
		"amount":        amount,
		"fixed":         false,
	}

	jsonBody, err := json.Marshal(quoteReq)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+l.apiKey)

	resp, err := l.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching quote: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Si no encuentra el par, retornar nil sin error para continuar con otros exchanges
		if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusNotFound {
			return nil, nil
		}

		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var result struct {
		EstimatedAmount float64 `json:"estimated_amount"`
		MinAmount       float64 `json:"min_amount"`
		MaxAmount       float64 `json:"max_amount"`
		Rate            float64 `json:"rate"`
		Fee             float64 `json:"fee"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	// Si no hay resultado válido
	if result.EstimatedAmount <= 0 {
		return nil, nil
	}

	return &models.Quote{
		Exchange:   l.GetName(),
		From:       from,
		To:         to,
		FromAmount: amount,
		ToAmount:   result.EstimatedAmount,
		Rate:       result.Rate,
		MinAmount:  result.MinAmount,
		MaxAmount:  result.MaxAmount,
		Timestamp:  time.Now(),
	}, nil
}

// GetMinAmount obtiene el monto mínimo para un par
func (l *LetsExchange) GetMinAmount(from, to string) (float64, error) {
	url := fmt.Sprintf("%s/range?api_key=%s", l.baseURL, l.apiKey)

	// Preparar el request body
	rangeReq := map[string]interface{}{
		"from_currency": from,
		"to_currency":   to,
	}

	jsonBody, err := json.Marshal(rangeReq)
	if err != nil {
		return 0, fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return 0, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+l.apiKey)

	resp, err := l.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("error fetching range: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var result struct {
		MinAmount float64 `json:"min_amount"`
		MaxAmount float64 `json:"max_amount"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("error decoding response: %w", err)
	}

	return result.MinAmount, nil
}

// CreateExchange crea un intercambio real
func (l *LetsExchange) CreateExchange(req models.SwapRequest) (*models.SwapResponse, error) {
	url := fmt.Sprintf("%s/exchange?api_key=%s", l.baseURL, l.apiKey)

	// Preparar el request body
	exchangeReq := map[string]interface{}{
		"from_currency":  req.From,
		"to_currency":    req.To,
		"amount":         req.Amount,
		"to_address":     req.ToAddress,
		"refund_address": req.RefundAddress,
		"fixed":          false,
	}

	jsonBody, err := json.Marshal(exchangeReq)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	reqHTTP, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	reqHTTP.Header.Set("Content-Type", "application/json")
	reqHTTP.Header.Set("Authorization", "Bearer "+l.apiKey)

	resp, err := l.client.Do(reqHTTP)
	if err != nil {
		return nil, fmt.Errorf("error creating exchange: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var result struct {
		ID              string  `json:"id"`
		Status          string  `json:"status"`
		FromCurrency    string  `json:"from_currency"`
		ToCurrency      string  `json:"to_currency"`
		Amount          float64 `json:"amount"`
		EstimatedAmount float64 `json:"estimated_amount"`
		PayinAddress    string  `json:"payin_address"`
		PayoutAddress   string  `json:"payout_address"`
		CreatedAt       string  `json:"created_at"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	// Parsear la fecha de creación
	createdAt := time.Now()
	if result.CreatedAt != "" {
		if parsed, err := time.Parse(time.RFC3339, result.CreatedAt); err == nil {
			createdAt = parsed
		}
	}

	return &models.SwapResponse{
		ID:            result.ID,
		Status:        result.Status,
		From:          req.From,
		To:            req.To,
		PayinAddress:  result.PayinAddress,
		PayinAmount:   result.Amount,
		PayoutAmount:  result.EstimatedAmount,
		PayoutAddress: result.PayoutAddress,
		Exchange:      l.GetName(),
		CreatedAt:     createdAt,
	}, nil
}

// GetExchangeStatus obtiene el estado de un intercambio
func (l *LetsExchange) GetExchangeStatus(id string) (string, error) {
	url := fmt.Sprintf("%s/exchange/%s/status?api_key=%s", l.baseURL, id, l.apiKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+l.apiKey)

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
