package httpclient

import (
	"context"
	"cryptoswap/internal/lib/logger"

	"github.com/go-resty/resty/v2"
)

type Factory interface {
	NewClient(ctx context.Context) HttpClient
}

func NewHttpClientFactory(config HttpConfig, logger logger.Logger) Factory {
	return &factory{
		client: resty.New().
			SetBaseURL(config.BaseURL).
			SetTimeout(config.Timeout),
	}
}

type factory struct {
	client *resty.Client
	logger logger.Logger
}

func (f *factory) NewClient(ctx context.Context) HttpClient {
	return &httpClient{
		req:         f.client.R(),
		ctx:         ctx,
		logger:      f.logger,
		headers:     make(map[string]string),
		queryParams: make(map[string]any),
	}
}
