// exchanges/changenow.go
package exchanges

import (
	"bytes"
	"cryptoswap/internal/services/models"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type ChangeNow struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

func NewChangeNow(apiKey string) *ChangeNow {
	return &ChangeNow{
		apiKey:  apiKey,
		baseURL: "https://api.changenow.io/v1",
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *ChangeNow) GetName() string {
	return "ChangeNOW"
}

// GetCurrencies obtiene todas las monedas disponibles
func (c *ChangeNow) GetCurrencies() ([]models.Currency, error) {
	url := fmt.Sprintf("%s/currencies?active=true", c.baseURL)

	resp, err := c.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching currencies: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: status %d", resp.StatusCode)
	}

	var apiCurrencies []struct {
		Ticker      string `json:"ticker"`
		Name        string `json:"name"`
		Image       string `json:"image"`
		IsAvailable bool   `json:"isAvailable"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiCurrencies); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	currencies := make([]models.Currency, 0, len(apiCurrencies))
	for _, curr := range apiCurrencies {
		if curr.IsAvailable {
			currencies = append(currencies, models.Currency{
				Symbol:    curr.Ticker,
				Name:      curr.Name,
				Image:     curr.Image,
				Available: true,
			}.WithProvider(c.GetName()))
		}
	}

	return currencies, nil
}

// GetQuote obtiene una cotización
func (c *ChangeNow) GetQuote(from, to string, amount float64) (*models.Quote, error) {
	url := fmt.Sprintf("%s/exchange-amount/%f/%s_%s?api_key=%s",
		c.baseURL, amount, from, to, c.apiKey)

	resp, err := c.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching quote: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Si no encuentra el par, retornar nil sin error para continuar con otros exchanges
		if resp.StatusCode == http.StatusBadRequest {
			return nil, nil
		}
		return nil, fmt.Errorf("API error: status %d", resp.StatusCode)
	}

	var result struct {
		EstimatedAmount  float64 `json:"estimatedAmount"`
		TransactionSpeed string  `json:"transactionSpeedForecast"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	// Si no hay resultado válido
	if result.EstimatedAmount <= 0 {
		return nil, nil
	}

	return &models.Quote{
		Exchange:   c.GetName(),
		From:       from,
		To:         to,
		FromAmount: amount,
		ToAmount:   result.EstimatedAmount,
		Rate:       result.EstimatedAmount / amount,
		Timestamp:  time.Now(),
	}, nil
}

// GetMinAmount obtiene el monto mínimo para un par
func (c *ChangeNow) GetMinAmount(from, to string) (float64, error) {
	url := fmt.Sprintf("%s/min-amount/%s_%s?api_key=%s",
		c.baseURL, from, to, c.apiKey)

	resp, err := c.client.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result struct {
		MinAmount float64 `json:"minAmount"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	return result.MinAmount, nil
}

// CreateExchange crea un intercambio real
func (c *ChangeNow) CreateExchange(req models.SwapRequest) (*models.SwapResponse, error) {
	// Preparar el request body
	exchangeReq := map[string]interface{}{
		"from":          req.From,
		"to":            req.To,
		"address":       req.ToAddress,
		"amount":        req.Amount,
		"extraId":       "",
		"refundAddress": req.RefundAddress,
	}

	jsonBody, err := json.Marshal(exchangeReq)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/transactions/%s", c.baseURL, c.apiKey)
	resp, err := c.client.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API error: status %d", resp.StatusCode)
	}

	var result struct {
		Id             string  `json:"id"`
		PayinAddress   string  `json:"payinAddress"`
		PayoutAddress  string  `json:"payoutAddress"`
		FromCurrency   string  `json:"fromCurrency"`
		ToCurrency     string  `json:"toCurrency"`
		Amount         float64 `json:"amount"`
		ExpectedAmount float64 `json:"expectedAmount"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &models.SwapResponse{
		ID:            result.Id,
		Status:        "waiting",
		From:          req.From,
		To:            req.To,
		PayinAddress:  result.PayinAddress,
		PayinAmount:   result.Amount,
		PayoutAmount:  result.ExpectedAmount,
		PayoutAddress: result.PayoutAddress,
		Exchange:      c.GetName(),
		CreatedAt:     time.Now(),
	}, nil
}
