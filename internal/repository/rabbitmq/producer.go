package rabbitmq

import (
	"context"
	"cryptoswap/internal/lib/apierrors"
	"cryptoswap/internal/lib/constants"
	"cryptoswap/internal/lib/logger"
	"cryptoswap/internal/lib/messaging"
	"cryptoswap/internal/services/interfaces"
	"cryptoswap/internal/services/models"
)

func NewExchangeNotifier(logger logger.Logger, conn messaging.Publisher) interfaces.SwapNotifier {
	return &exchangeNotifier{
		logger: logger,
		conn:   conn,
	}
}

type exchangeNotifier struct {
	logger logger.Logger
	conn   messaging.Publisher
}

func (e *exchangeNotifier) NotifySwap(ctx context.Context, swap models.Swap) *apierrors.ApiError {
	err := e.conn.Publish(ctx, messaging.NewMessageBuilder().
		WithRoutingKey(constants.SwapRoutingKey).
		WithRequestId(constants.GetRequestId(ctx)).
		WithJSONBody(swap).
		Build())
	if err != nil {
		return apierrors.NewApiError(apierrors.InternalServer, err)
	}
	return nil
}
