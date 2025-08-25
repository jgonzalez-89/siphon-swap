package exchanges

import (
	"bytes"
	"cryptoswap/models"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type StealthEx struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

func NewStealthEx(apiKey string) *StealthEx {
	return &StealthEx{
		apiKey:  apiKey,
		baseURL: "https://api.stealthex.io/api/v2",
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *StealthEx) GetName() string {
	return "StealthEX"
}

// GetCurrencies obtiene todas las monedas disponibles
func (s *StealthEx) GetCurrencies() ([]models.Currency, error) {
	url := fmt.Sprintf("%s/currency?api_key=%s", s.baseURL, s.apiKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching currencies: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: status %d", resp.StatusCode)
	}

	var apiCurrencies []struct {
		Symbol     string `json:"symbol"`
		Name       string `json:"name"`
		Image      string `json:"image"`
		Network    string `json:"network"`
		HasExtraId bool   `json:"has_extra_id"`
		IsStable   bool   `json:"is_stable"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiCurrencies); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	currencies := make([]models.Currency, 0, len(apiCurrencies))
	for _, curr := range apiCurrencies {
		currencies = append(currencies, models.Currency{
			Symbol:    curr.Symbol,
			Name:      curr.Name,
			Image:     curr.Image,
			Network:   curr.Network,
			Available: true,
		})
	}

	return currencies, nil
}

// GetQuote obtiene una cotización
func (s *StealthEx) GetQuote(from, to string, amount float64) (*models.Quote, error) {
	url := fmt.Sprintf("%s/estimate/%s/%s?amount=%f&api_key=%s&fixed=false",
		s.baseURL, from, to, amount, s.apiKey)

	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching quote: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Si no encuentra el par, retornar nil sin error
		if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("API error: status %d", resp.StatusCode)
	}

	var result rTMP
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	// Si no hay resultado válido
	if result.EstimatedAmount <= 0 {
		return nil, nil
	}

	return &models.Quote{
		Exchange:   s.GetName(),
		From:       from,
		To:         to,
		FromAmount: amount,
		ToAmount:   result.EstimatedAmount,
		Rate:       result.EstimatedAmount / amount,
		MinAmount:  result.Min,
		MaxAmount:  result.Max,
		Timestamp:  time.Now(),
	}, nil
}

type rTMP struct {
	EstimatedAmount float64 `json:"estimated_amount"`
	Min             float64 `json:"min_amount"`
	Max             float64 `json:"max_amount"`
}

func (r *rTMP) UnmarshalJSON(data []byte) error {
	var tmp struct {
		EstimatedAmount string  `json:"estimated_amount"`
		Min             float64 `json:"min_amount"`
		Max             float64 `json:"max_amount"`
	}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	tme, _ := strconv.ParseFloat(tmp.EstimatedAmount, 64)
	r.EstimatedAmount = tme
	r.Min = tmp.Min
	r.Max = tmp.Max

	return nil
}

// GetMinAmount obtiene el monto mínimo para un par
func (s *StealthEx) GetMinAmount(from, to string) (float64, error) {
	url := fmt.Sprintf("%s/range/%s/%s?api_key=%s",
		s.baseURL, from, to, s.apiKey)

	resp, err := s.client.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result struct {
		MinAmount float64 `json:"min_amount"`
		MaxAmount float64 `json:"max_amount"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	return result.MinAmount, nil
}

// CreateExchange crea un intercambio real
func (s *StealthEx) CreateExchange(req models.SwapRequest) (*models.SwapResponse, error) {
	// Preparar el request body
	exchangeReq := map[string]interface{}{
		"currency_from":  req.From,
		"currency_to":    req.To,
		"amount_from":    req.Amount,
		"address_to":     req.ToAddress,
		"extra_id_to":    "",
		"refund_address": req.RefundAddress,
		"rate_id":        "",
		"api_key":        s.apiKey,
	}

	jsonBody, err := json.Marshal(exchangeReq)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/exchange?api_key=%s", s.baseURL, s.apiKey)
	resp, err := s.client.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var errorResp struct {
			Message string `json:"message"`
		}
		json.NewDecoder(resp.Body).Decode(&errorResp)
		return nil, fmt.Errorf("API error: status %d, message: %s", resp.StatusCode, errorResp.Message)
	}

	var result struct {
		Id           string  `json:"id"`
		AddressFrom  string  `json:"address_from"`
		AddressTo    string  `json:"address_to"`
		CurrencyFrom string  `json:"currency_from"`
		CurrencyTo   string  `json:"currency_to"`
		AmountFrom   float64 `json:"amount_from"`
		AmountTo     float64 `json:"amount_to"`
		Status       string  `json:"status"`
		CreatedAt    string  `json:"created_at"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	createdAt, _ := time.Parse(time.RFC3339, result.CreatedAt)

	return &models.SwapResponse{
		ID:            result.Id,
		Status:        result.Status,
		From:          req.From,
		To:            req.To,
		PayinAddress:  result.AddressFrom,
		PayinAmount:   result.AmountFrom,
		PayoutAmount:  result.AmountTo,
		PayoutAddress: result.AddressTo,
		Exchange:      s.GetName(),
		CreatedAt:     createdAt,
	}, nil
}

// GetExchangeStatus obtiene el estado de un intercambio
func (s *StealthEx) GetExchangeStatus(id string) (string, error) {
	url := fmt.Sprintf("%s/exchange/%s?api_key=%s", s.baseURL, id, s.apiKey)

	resp, err := s.client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Status, nil
}
