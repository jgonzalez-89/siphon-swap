package httpclient

import (
	"context"
	"cryptoswap/internal/lib/logger"
	"time"

	"github.com/go-resty/resty/v2"
)

type HttpConfig struct {
	BaseURL    string
	ApiKey     string
	AuthScheme string
	Timeout    time.Duration
	AuthHeader string
}

func NewConfig(baseURL, apiKey, authScheme string, timeout time.Duration) HttpConfig {
	return HttpConfig{
		BaseURL:    baseURL,
		ApiKey:     apiKey,
		AuthScheme: authScheme,
		Timeout:    timeout,
	}
}

func NewConfigWithAuthHeader(baseURL, apiKey, authHeader string, timeout time.Duration) HttpConfig {
	return HttpConfig{
		BaseURL:    baseURL,
		ApiKey:     apiKey,
		AuthHeader: authHeader,
		Timeout:    timeout,
	}
}

type Factory interface {
	NewClient(ctx context.Context) HttpClient
}

func NewFactory(config HttpConfig, logger logger.Logger) Factory {
	r := resty.New().SetBaseURL(config.BaseURL).
		SetTimeout(config.Timeout)

	if config.AuthScheme != "" {
		r = r.SetAuthScheme(config.AuthScheme).
			SetAuthToken(config.ApiKey)
	}

	if config.AuthHeader != "" {
		r = r.SetHeader(config.AuthHeader, config.ApiKey)
	}

	return &factory{
		client: r,
		logger: logger,
	}
}

type factory struct {
	client *resty.Client
	logger logger.Logger
}

func (f *factory) NewClient(ctx context.Context) HttpClient {
	return &httpClient{
		req: f.client.R().
			SetHeaderMultiValues(f.client.Header),
		ctx:    ctx,
		logger: f.logger,
	}
}
