package constants

import (
	"context"
	"cryptoswap/internal/lib/ids"
	"time"
)

const (
	RequestId = "requestId"
	startTick = "startTick"
)

func GetRequestId(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if ctx.Value(RequestId) == nil {
		return ""
	}
	return ctx.Value(RequestId).(string)
}

func SetRequestId(ctx context.Context, requestId string) context.Context {
	return context.WithValue(ctx, RequestId, requestId)
}

func AddRequestIdToContext(ctx context.Context) context.Context {
	return context.WithValue(ctx,
		RequestId, ids.NewRequestId())
}

func NewContextWithRequestId() context.Context {
	return context.WithValue(context.Background(),
		RequestId, ids.NewRequestId())
}

func Tick(ctx context.Context) context.Context {
	return context.WithValue(ctx, startTick, time.Now())
}

func Tock(ctx context.Context) time.Time {
	return ctx.Value(startTick).(time.Time)
}
