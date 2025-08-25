package constants

import (
	"context"
	"cryptoswap/internal/ids"
)

const (
	RequestId = "requestId"
)

func GetRequestId(ctx context.Context) string {
	if ctx.Value(RequestId) == nil {
		return ids.NewRequestId()
	}
	return ctx.Value(RequestId).(string)
}

func NewContextWithRequestId() context.Context {
	return context.WithValue(context.Background(),
		RequestId, ids.NewRequestId())
}
