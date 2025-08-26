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
	if ctx.Value(RequestId) == nil {
		return ""
	}
	return ctx.Value(RequestId).(string)
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
