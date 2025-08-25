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
}

type Factory interface {
	NewClient(ctx context.Context) HttpClient
}

func NewFactory(config HttpConfig, logger logger.Logger) Factory {
	return &factory{
		client: resty.New().
			SetBaseURL(config.BaseURL).
			SetTimeout(config.Timeout).
			SetAuthScheme(config.AuthScheme).
			SetAuthToken(config.ApiKey),
		logger: logger,
	}
}

type factory struct {
	client *resty.Client
	logger logger.Logger
}

func (f *factory) NewClient(ctx context.Context) HttpClient {
	return &httpClient{
		req:    f.client.R(),
		ctx:    ctx,
		logger: f.logger,
	}
}
