package exchanges

import (
	"bytes"
	"cryptoswap/internal/services/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type SimpleSwap struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

func NewSimpleSwap(apiKey string) *SimpleSwap {
	return &SimpleSwap{
		apiKey:  apiKey,
		baseURL: "https://api.simpleswap.io",
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *SimpleSwap) GetName() string {
	return "SimpleSwap"
}

// GetCurrencies obtiene todas las monedas disponibles
func (s *SimpleSwap) GetCurrencies() ([]models.Currency, error) {
	url := fmt.Sprintf("%s/get_all_currencies?api_key=%s", s.baseURL, s.apiKey)

	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching currencies: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: status %d", resp.StatusCode)
	}

	var apiCurrencies []struct {
		Symbol  string `json:"symbol"`
		Name    string `json:"name"`
		Image   string `json:"image"`
		Network string `json:"network"`
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
		}.WithProvider(s.GetName()))
	}

	return currencies, nil
}

// GetQuote obtiene una cotización
func (s *SimpleSwap) GetQuote(from, to string, amount float64) (*models.Quote, error) {
	url := fmt.Sprintf("%s/get_estimated?api_key=%s&fixed=false&currency_from=%s&currency_to=%s&amount=%f",
		s.baseURL, s.apiKey, from, to, amount)

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

	// SimpleSwap devuelve directamente el número estimado como string
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	estimatedAmount, err := strconv.ParseFloat(string(body), 64)
	if err != nil {
		// Intentar decodificar como JSON en caso de error
		var jsonResp float64
		if err := json.Unmarshal(body, &jsonResp); err != nil {
			return nil, fmt.Errorf("error parsing response: %s", string(body))
		}
		estimatedAmount = jsonResp
	}

	// Si no hay resultado válido
	if estimatedAmount <= 0 {
		return nil, nil
	}

	return &models.Quote{
		Exchange:   s.GetName(),
		From:       from,
		To:         to,
		FromAmount: amount,
		ToAmount:   estimatedAmount,
		Rate:       estimatedAmount / amount,
		Timestamp:  time.Now(),
	}, nil
}

// GetMinAmount obtiene el monto mínimo para un par
func (s *SimpleSwap) GetMinAmount(from, to string) (float64, error) {
	url := fmt.Sprintf("%s/get_ranges?api_key=%s&fixed=false&currency_from=%s&currency_to=%s",
		s.baseURL, s.apiKey, from, to)

	resp, err := s.client.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result struct {
		Min float64 `json:"min"`
		Max float64 `json:"max"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	return result.Min, nil
}

// CreateExchange crea un intercambio real
func (s *SimpleSwap) CreateExchange(req models.Swap) (*models.SwapResponse, error) {
	// Preparar el request body
	exchangeReq := map[string]interface{}{
		"fixed":               false,
		"currency_from":       req.From,
		"currency_to":         req.To,
		"amount":              req.Amount,
		"address_to":          req.ToAddress,
		"extra_id_to":         "",
		"user_refund_address": req.RefundAddress,
		"api_key":             s.apiKey,
	}

	jsonBody, err := json.Marshal(exchangeReq)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/create_exchange?api_key=%s", s.baseURL, s.apiKey)
	resp, err := s.client.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
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
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

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
		CreatedAt:     time.Now(),
	}, nil
}
