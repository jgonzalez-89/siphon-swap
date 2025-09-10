package models

import (
	"cryptoswap/internal/lib/apierrors"
	"cryptoswap/internal/lib/ids"
	"errors"
	"regexp"
	"time"
)

const (
	swapStatusPending   = "PENDING"
	swapStatusCompleted = "COMPLETED"
	swapStatusFailed    = "FAILED"
)

// TODO: validate min amount -> do a swap processor with repository calls
func NewSwap(payinAmount float64, from, to NetworkPair, toAdress,
	refundAddress, exchange string) Swap {
	return Swap{
		Id:            ids.NewSwapRequestId(),
		From:          from,
		To:            to,
		PayinAmount:   payinAmount,
		ToAddress:     toAdress,
		RefundAddress: refundAddress,
		Status:        swapStatusPending,
		Exchange:      exchange,
	}
}

type Swap struct {
	Id            string      `json:"id"`
	From          NetworkPair `json:"from"`
	To            NetworkPair `json:"to"`
	ExchangeId    string      `json:"exchangeId"`
	PayinAmount   float64     `json:"payinAmount"`
	PayoutAmount  float64     `json:"payoutAmount"`
	PayoutAddress string      `json:"payoutAddress"`
	ToAddress     string      `json:"toAddress"`
	RefundAddress string      `json:"refundAddress"`
	Exchange      string      `json:"exchange"`
	Reason        string      `json:"reason"`
	Status        string      `json:"status"`
	CreatedAt     time.Time   `json:"createdAt"`
	UpdatedAt     time.Time   `json:"updatedAt"`
}

func (s *Swap) WithBillingConditions(payoutAddress, exchangeId string, payoutAmount float64) *Swap {
	s.ExchangeId = exchangeId
	s.PayoutAddress = payoutAddress
	s.PayoutAmount = payoutAmount
	return s
}

func (s *Swap) Complete() *Swap {
	s.Status = swapStatusCompleted
	return s
}

func (s *Swap) Fail(reason string) *Swap {
	s.Status = swapStatusFailed
	s.Reason = reason
	return s
}

func (s *Swap) HasValidAddress(curr Currency) *apierrors.ApiError {
	if curr.AddressValidation != "" {
		regexp, err := regexp.Compile(curr.AddressValidation)
		if err != nil {
			return apierrors.NewApiError(apierrors.InternalServer, err)
		}

		if !regexp.MatchString(s.ToAddress) {
			return apierrors.NewApiError(apierrors.BadRequest, errors.New("invalid address"))
		}
	}
	return nil
}
