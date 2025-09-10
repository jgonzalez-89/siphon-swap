package handlers

import (
	"cryptoswap/internal/lib/api"
	"cryptoswap/internal/lib/apierrors"
	"cryptoswap/internal/lib/logger"
	"cryptoswap/internal/services/currencies"
	"cryptoswap/internal/services/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewHandlers(logger logger.Logger, handler api.ResponseHandler,
	service currencies.CurrencyService) ServerInterface {
	return &handlersImpl{
		logger:  logger,
		handler: handler,
		service: service,
	}
}

type handlersImpl struct {
	logger  logger.Logger
	handler api.ResponseHandler
	service currencies.CurrencyService
}

func (h *handlersImpl) GetV1Currencies(c *gin.Context, params GetV1CurrenciesParams) {
	currencies, err := h.service.GetCurrencies(c, toFilter(params))
	if err != nil {
		h.handler.Error(c, err)
		return
	}

	h.handler.OK(c, http.StatusOK, toCurrencies(currencies))
}

func (h *handlersImpl) GetV1Quotes(c *gin.Context, params GetV1QuotesParams) {
	fromPair := toPair(params.FromSymbol, params.FromNetwork)
	toPair := toPair(params.ToSymbol, params.ToNetwork)
	quote, err := h.service.GetQuotes(c, fromPair, toPair, params.Amount)
	if err != nil {
		h.handler.Error(c, err)
		return
	}

	h.handler.OK(c, http.StatusOK, toQuotes(quote))
}

func (h *handlersImpl) GetV1SwapsId(c *gin.Context, id string) {
	swap, err := h.service.GetSwap(c, id)
	if err != nil {
		h.handler.Error(c, err)
		return
	}
	h.handler.OK(c, http.StatusOK, toSwap(swap))
}

func (h *handlersImpl) PostV1Swaps(c *gin.Context) {
	var swapRequest SwapRequest
	if err := c.ShouldBindJSON(&swapRequest); err != nil {
		h.handler.Error(c, apierrors.NewApiError(apierrors.BadRequest, err))
		return
	}

	from := toPairFromRequest(swapRequest.From)
	to := toPairFromRequest(swapRequest.To)
	swap := models.NewSwap(swapRequest.Amount, from, to, swapRequest.ToAddress,
		swapRequest.RefundAddress, swapRequest.Exchange)

	insertedSwap, err := h.service.InsertSwap(c, swap)
	if err != nil {
		h.handler.Error(c, err)
		return
	}

	h.handler.OK(c, http.StatusOK, toSwap(insertedSwap))
}
