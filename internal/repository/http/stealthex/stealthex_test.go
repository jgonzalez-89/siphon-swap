package stealthex

import (
	"context"
	"cryptoswap/internal/lib/httpclient"
	"cryptoswap/internal/lib/logger"
	"cryptoswap/internal/services/models"
	"testing"
	"time"
)

func Test_Stealth(t *testing.T) {
	fact := logger.NewLoggerFactory("test", "info")

	httpFact := httpclient.NewFactory(httpclient.HttpConfig{
		BaseURL:    "https://api.stealthex.io/v4",
		Timeout:    10 * time.Second,
		AuthScheme: "Bearer",
		ApiKey:     "86db32f5-f673-4d3e-a0b2-9615701d47f5",
	}, fact.NewLogger("http_client"))
	client := NewStealthExRepository(fact.NewLogger("stealthex"), httpFact)

	ctx := context.Background()
	from := models.NetworkPair{
		Symbol:  "btc",
		Network: "mainnet",
	}
	to := models.NetworkPair{
		Symbol:  "eth",
		Network: "mainnet",
	}
	quote, err := client.GetQuote(ctx, from, to, 0.1)
	if err != nil {
		t.Fatalf("error getting quote: %v", err)
	}
	t.Logf("quote: %+v", quote)
}
