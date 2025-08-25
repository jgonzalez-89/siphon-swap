package stealthex

import (
	"context"
	"cryptoswap/internal/lib/httpclient"
	"cryptoswap/internal/lib/logger"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCallCreateExchange(t *testing.T) {
	lFact := logger.NewLoggerFactory("stealthex", "test")

	factory := httpclient.NewFactory(httpclient.HttpConfig{
		BaseURL:    "https://api.stealthex.io/api/v2",
		Timeout:    10 * time.Second,
		ApiKey:     "96d73f44-987c-42ba-a2e7-658778e02940",
		AuthScheme: "Bearer",
	}, lFact.NewLogger("http-client"))

	stealthex := NewStealthClient(lFact.NewLogger("stealthex"), factory)

	ctx := context.Background()
	exchange, err := stealthex.GetCurrencies(ctx)

	assert.Nil(t, err)
	assert.NotNil(t, exchange)
}
