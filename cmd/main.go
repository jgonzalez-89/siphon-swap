package main

import (
	"context"
	"cryptoswap/internal/config"
	"cryptoswap/internal/lib/api"
	"cryptoswap/internal/lib/db"
	"cryptoswap/internal/lib/httpclient"
	"cryptoswap/internal/lib/logger"
	"cryptoswap/internal/lib/messaging"
	"cryptoswap/internal/lib/middlewares"
	"cryptoswap/internal/lib/server"
	"cryptoswap/internal/repository/currencies"
	"cryptoswap/internal/repository/http/changenow"
	"cryptoswap/internal/repository/http/coingecko"
	"cryptoswap/internal/repository/http/stealthex"
	"cryptoswap/internal/repository/rabbitmq"
	currService "cryptoswap/internal/services/currencies"
	"cryptoswap/internal/services/daemon"
	"cryptoswap/internal/transport/consumer"
	currHandlers "cryptoswap/internal/transport/handlers/handlers"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	ctx := context.Background()
	fact := logger.NewLoggerFactory("test", "info")
	mainLogger := fact.NewLogger("main")

	// Load config:
	cfg, err := config.LoadConfig()
	if err != nil {
		mainLogger.Fatalf(ctx, "error loading config: %v", err)
	}

	// Connectors:
	db, err := db.NewGorm(db.Config(cfg.Database), fact.NewLogger("gorm"))
	if err != nil {
		mainLogger.Fatalf(ctx, "error connecting to database: %v", err)
	}

	msgConfig := messaging.NewConfig(cfg.Messaging,
		[]messaging.Queue{messaging.NewQueue("app.events.q")})
	msgConn, err := messaging.NewConnection(fact.NewLogger("messaging"), msgConfig)
	if err != nil {
		mainLogger.Fatalf(ctx, "error creating messaging connection: %v", err)
	}

	// Repositories:
	changenow := changenow.NewChangeNowRepository(fact.NewLogger("changenow"),
		httpclient.NewFactory(httpclient.HttpConfig{
			ApiKey:     cfg.Exchanges.ChangeNow.ApiKey,
			AuthHeader: cfg.Exchanges.ChangeNow.AuthHeader,
			Timeout:    time.Duration(cfg.Exchanges.ChangeNow.TimeoutSeconds) * time.Second,
			BaseURL:    cfg.Exchanges.ChangeNow.BaseURL,
		}, fact.NewLogger("http_client")))

	stealthex := stealthex.NewStealthExRepository(fact.NewLogger("stealthex"),
		httpclient.NewFactory(httpclient.HttpConfig{
			BaseURL:    cfg.Exchanges.StealthEx.BaseURL,
			Timeout:    time.Duration(cfg.Exchanges.StealthEx.TimeoutSeconds) * time.Second,
			ApiKey:     cfg.Exchanges.StealthEx.ApiKey,
			AuthScheme: cfg.Exchanges.StealthEx.AuthScheme,
		}, fact.NewLogger("http_client")))

	coingecko := coingecko.NewCoinGecko(fact.NewLogger("coingecko"), httpclient.NewFactory(httpclient.HttpConfig{
		ApiKey:     cfg.Exchanges.CoinGecko.ApiKey,
		AuthHeader: cfg.Exchanges.CoinGecko.AuthHeader,
		Timeout:    time.Duration(cfg.Exchanges.CoinGecko.TimeoutSeconds) * time.Second,
		BaseURL:    cfg.Exchanges.CoinGecko.BaseURL,
	}, fact.NewLogger("http_client")))

	currDB := currencies.NewDB(fact.NewLogger("database"), db)

	msgNotifier := rabbitmq.NewExchangeNotifier(fact.NewLogger("messaging"), msgConn)

	// Services:
	currencyManager := daemon.NewCurrencyManager(fact.NewLogger("daemon"), currDB, coingecko,
		changenow, stealthex)

	currencyService := currService.NewCurrencyService(fact.NewLogger("currency_service"), currDB,
		msgNotifier, changenow, stealthex)

	// Handlers:
	currencyHandler := currHandlers.NewHandlers(fact.NewLogger("handlers"),
		api.NewResponseManager(), currencyService)

	consumerHandler := consumer.NewMessagingConsumer(fact.NewLogger("consumer"), currencyService).
		Build()

	// Server:
	router := gin.New()
	handlerFactory := server.NewHandlerFactory(ctx, mainLogger)
	serverBuilder := server.NewServerBuilder(router, server.ServerConfig(cfg.Server))

	middlewareLogger := fact.NewLogger("middlewares")
	httpServer := serverBuilder.
		WithHandlers(handlerFactory.New(currencyHandler, currHandlers.RegisterHandlers, currHandlers.GetSwagger)).
		WithMiddlewares(middlewares.CorsMiddleware,
			middlewares.LoggingMiddleware(middlewareLogger)).
		Build()

	// Run processes:
	msgConn.Consume(ctx, consumerHandler)
	if cfg.IsDaemonEnabled() {
		go currencyManager.Start(ctx)
	}

	mainLogger.Printf("Starting server on address: %s", httpServer.Addr)
	if err := httpServer.ListenAndServe(); err != nil {
		mainLogger.Fatalf(ctx, "error starting server: %v", err)
	}
}
