package handlers

import (
	"cryptoswap/internal/lib/api"
	"cryptoswap/internal/lib/logger"
	"cryptoswap/internal/services/currencies"
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
	quote, err := h.service.GetQuote(c, fromPair, toPair, params.Amount)
	if err != nil {
		h.handler.Error(c, err)
		return
	}

	h.handler.OK(c, http.StatusOK, quote)
}
