package swap

import (
	"context"
	"strconv"
	"strings"
	"time"

	"cryptoswap/internal/lib/cache"
	"cryptoswap/internal/lib/logger"
	"cryptoswap/internal/repository/coingecko"
	"cryptoswap/internal/services/models"
)

type CoinGeckoService struct {
	logger    logger.Logger
	coinGecko coingecko.CoinGecko
	cache     *cache.Cache
}

func NewCoinGeckoService(logger logger.Logger, coinGecko coingecko.CoinGecko) *CoinGeckoService {
	return &CoinGeckoService{
		logger:    logger,
		coinGecko: coinGecko,
		cache:     cache.NewCache(15 * time.Second),
	}
}

// TopTickers devuelve top N por market cap con precio y % 24h.
func (s *CoinGeckoService) TopTickers(ctx context.Context, vs string, n int) ([]models.Ticker, error) {
	vs, n = s.cleanupVsAndN(vs, n)
	cacheKey := "cg:top:" + strings.ToLower(vs) + ":" + strconv.Itoa(n)

	if v, ok := s.cache.Get(cacheKey); ok {
		return v.([]models.Ticker), nil
	}

	tickers, err := s.coinGecko.TopTickers(ctx, vs, n)
	if err != nil {
		return nil, err
	}

	s.cache.Set(cacheKey, tickers, 15*time.Second)
	return tickers, nil
}

func (s *CoinGeckoService) cleanupVsAndN(vs string, n int) (string, int) {
	vs = strings.ToLower(vs)
	if vs == "" {
		vs = "usd"
	}

	return vs, min(max(n, 0), 250)
}
