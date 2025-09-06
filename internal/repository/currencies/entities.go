package currencies

import (
	"cryptoswap/internal/services/models"

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
