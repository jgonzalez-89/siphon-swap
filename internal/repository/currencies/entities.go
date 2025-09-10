package currencies

import (
	"cryptoswap/internal/services/models"
	"time"

	"github.com/samber/lo"
)

type Currencies []Currency

func (c Currencies) ToModel() []models.Currency {
	return lo.Map(c, func(c Currency, _ int) models.Currency {
		return c.ToModel()
	})
}

type Currency struct {
	Symbol            string            `gorm:"column:symbol;primaryKey"`
	Name              string            `gorm:"column:name"`
	Image             string            `gorm:"column:image"`
	Available         bool              `gorm:"column:available"`
	Popular           bool              `gorm:"column:popular"`
	Price             float64           `gorm:"column:price"`
	AddressValidation string            `gorm:"column:address_validation"`
	Networks          []CurrencyNetwork `gorm:"foreignKey:Symbol"`
}

func (c Currency) TableName() string {
	return "currency"
}

func (c *Currency) ToModel() models.Currency {
	return models.Currency{
		Symbol:            c.Symbol,
		Name:              c.Name,
		Image:             c.Image,
		Available:         c.Available,
		AddressValidation: c.AddressValidation,
		Price:             c.Price,
	}.WithNetworks(lo.Map(c.Networks, func(cn CurrencyNetwork, _ int) string {
		return cn.Network
	})...)
}

type CurrencyNetwork struct {
	Symbol  string `gorm:"column:symbol;primaryKey"`
	Network string `gorm:"column:network;primaryKey"`
}

func (cn CurrencyNetwork) TableName() string {
	return "currencies_networks"
}

type Swap struct {
	Id            string    `gorm:"column:id;primaryKey"`
	FromSymbol    string    `gorm:"column:from_symbol"`
	FromNetwork   string    `gorm:"column:from_network"`
	ToSymbol      string    `gorm:"column:to_symbol"`
	ToNetwork     string    `gorm:"column:to_network"`
	ToAddress     string    `gorm:"column:to_address"`
	RefundAddress string    `gorm:"column:refund_address"`
	Exchange      string    `gorm:"column:exchange"`
	Status        string    `gorm:"column:status"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
	Reason        string    `gorm:"column:reason"`
	PayoutAddress string    `gorm:"column:payout_address"`
	PayoutAmount  float64   `gorm:"column:payout_amount"`
	PayinAmount   float64   `gorm:"column:payin_amount"`
	ExchangeId    string    `gorm:"column:exchange_id"`
}

func (s Swap) TableName() string {
	return "swap"
}

func (s Swap) ToModel() models.Swap {
	return models.Swap{
		Id:            s.Id,
		From:          models.NetworkPair{Symbol: s.FromSymbol, Network: s.FromNetwork},
		To:            models.NetworkPair{Symbol: s.ToSymbol, Network: s.ToNetwork},
		PayoutAddress: s.PayoutAddress,
		PayoutAmount:  s.PayoutAmount,
		PayinAmount:   s.PayinAmount,
		Status:        s.Status,
		CreatedAt:     s.CreatedAt,
		UpdatedAt:     s.UpdatedAt,
		Reason:        s.Reason,
		Exchange:      s.Exchange,
		ToAddress:     s.ToAddress,
		RefundAddress: s.RefundAddress,
		ExchangeId:    s.ExchangeId,
	}
}
