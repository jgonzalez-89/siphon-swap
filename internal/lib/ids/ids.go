package ids

import "github.com/google/uuid"

const (
	requestIdPrefix = "RQ_"
)

func NewRequestId() string {
	return requestIdPrefix + uuid.New().String()
}
