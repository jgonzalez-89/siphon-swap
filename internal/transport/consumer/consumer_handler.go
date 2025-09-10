package consumer

import (
	"context"
	"cryptoswap/internal/lib/constants"
	"cryptoswap/internal/lib/logger"
	"cryptoswap/internal/lib/messaging"
	"cryptoswap/internal/services/currencies"
	"cryptoswap/internal/services/models"
	"encoding/json"
	"errors"
)

func NewMessagingConsumer(logger logger.Logger,
	service currencies.CurrencyService) messaging.ConsumerBuilder {
	return &messagingConsumer{
		logger:  logger,
		service: service,
	}
}

type messagingConsumer struct {
	logger  logger.Logger
	service currencies.CurrencyService
}

func (c *messagingConsumer) Build() messaging.Handler {
	return func(ctx context.Context, msg messaging.Message) error {
		ctx = constants.SetRequestId(ctx, msg.RequestId)
		swap := models.Swap{}
		if err := json.Unmarshal(msg.Body, &swap); err != nil {
			return err
		}

		// TODO: fetch current transaction status
		if err := c.service.ProcessSwap(ctx, swap); err != nil {
			return errors.New(err.Error())
		}

		return nil
	}
}
