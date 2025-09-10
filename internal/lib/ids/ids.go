package ids

import "github.com/google/uuid"

const (
	requestIdPrefix     = "RQ_"
	swapRequestIdPrefix = "SWP_"
)

func NewRequestId() string {
	return requestIdPrefix + uuid.New().String()
}

func NewSwapRequestId() string {
	return swapRequestIdPrefix + uuid.New().String()
}
