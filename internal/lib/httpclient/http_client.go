package httpclient

import (
	"context"
	"cryptoswap/internal/lib/logger"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

type HttpConfig struct {
	BaseURL string
	Timeout time.Duration
}

type httpClient struct {
	req    *resty.Request
	ctx    context.Context
	logger logger.Logger
}

type HttpClient interface {
	WithHeader(key string, value string) HttpClient
	WithAuthHeader(apiKey string) HttpClient
	WithApiKeyInQuery(apiKey string) HttpClient
	WithQueryParams(key string, value any) HttpClient
	WithBody(body any) HttpClient
	Get(endpoint string) ([]byte, int, error)
	Post(endpoint string) ([]byte, int, error)
}

func (c *httpClient) WithHeader(key string, value string) HttpClient {
	c.req.SetHeader(key, value)
	return c
}

func (c *httpClient) WithAuthHeader(apiKey string) HttpClient {
	c.req.SetAuthToken(apiKey)
	return c
}

func (c *httpClient) WithApiKeyInQuery(apiKey string) HttpClient {
	c.req.SetQueryParam("api_key", apiKey)
	return c
}

func (c *httpClient) WithQueryParams(key string, value any) HttpClient {
	c.req.SetQueryParam(key, fmt.Sprintf("%v", value))
	return c
}

func (c *httpClient) WithBody(body any) HttpClient {
	c.req.SetBody(body)
	return c
}

func (c *httpClient) Get(endpoint string) ([]byte, int, error) {
	response, err := c.req.Get(endpoint)
	if err != nil {
		c.logger.Errorf(c.ctx, "Error fetching %s, reason: %v", endpoint, err)
		return nil, http.StatusInternalServerError, err
	}

	return response.Body(), response.StatusCode(), nil
}

func (c *httpClient) Post(endpoint string) ([]byte, int, error) {
	response, err := c.req.SetHeader("Content-Type", "application/json").Post(endpoint)
	if err != nil {
		c.logger.Errorf(c.ctx, "Error fetching %s, reason: %v", endpoint, err)
		return nil, http.StatusInternalServerError, err
	}

	return response.Body(), response.StatusCode(), nil
}

type Request func(endpoint string) ([]byte, int, error)

func HandleRequest[T any](request Request, endpoint string, wantStatus int) (T, error) {
	var cast T
	body, gotStatus, err := request(endpoint)
	if err != nil {
		return cast, err
	}

	if gotStatus != wantStatus {
		return cast, fmt.Errorf("expected %d status code, got %d", wantStatus, gotStatus)
	}

	if err := json.Unmarshal(body, &cast); err != nil {
		return cast, fmt.Errorf("error unmarshalling body: %w", err)
	}

	return cast, nil
}
