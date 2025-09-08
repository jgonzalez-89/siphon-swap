package currencies

import (
	"context"
	"cryptoswap/internal/lib/apierrors"
	"cryptoswap/internal/lib/logger"
	"cryptoswap/internal/services/interfaces"
	"cryptoswap/internal/services/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func NewDB(logger logger.Logger, db *gorm.DB) interfaces.CurrencyRepository {
	return &currenciesRepository{
		logger: logger,
		db:     db,
	}
}

type currenciesRepository struct {
	logger logger.Logger
	db     *gorm.DB
}

func (cr *currenciesRepository) GetCurrencies(ctx context.Context,
	filters models.Filters) ([]models.Currency, *apierrors.ApiError) {
	cr.logger.Infof(ctx, "Getting currencies from the database")

	filterMap := filters.ToMap()
	entities := Currencies{}
	if err := cr.db.WithContext(ctx).
		Preload("Networks").
		Where(filterMap).
		Find(&entities).
		Error; err != nil {
		cr.logger.Infof(ctx, "Error getting currencies from the database: %v", err)
		return nil, apierrors.NewApiError(apierrors.InternalServerError, err)
	}

	return entities.ToModel(), nil
}

func (cr *currenciesRepository) GetCurrenciesByPairs(ctx context.Context,
	pairs ...models.NetworkPair) ([]models.Currency, *apierrors.ApiError) {
	if len(pairs) == 0 {
		return []models.Currency{}, nil
	}

	symbols := cr.db.Model(&CurrencyNetwork{}).
		Select("symbol").
		Where("(symbol, network) IN ?", toPairSlice(pairs))

	entities := Currencies{}
	if err := cr.db.WithContext(ctx).
		Preload("Networks").
		Where("symbol IN (?)", symbols).
		Find(&entities).
		Error; err != nil {
		return nil, apierrors.NewApiError(apierrors.InternalServerError, err)
	}

	return entities.ToModel(), nil
}

func (cr *currenciesRepository) InsertCurrencies(ctx context.Context, currencies []models.Currency,
) *apierrors.ApiError {
	cr.logger.Infof(ctx, "Inserting %d currencies into the database", len(currencies))

	entities := toCurrenciesEntity(currencies)
	if err := cr.db.Transaction(func(tx *gorm.DB) error {
		// Upsert currencies first
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "symbol"}},
			DoUpdates: clause.AssignmentColumns([]string{"name", "image", "available",
				"address_validation", "popular"}),
		}).Create(&entities).Error; err != nil {
			return err
		}

		// Delete ALL networks (careful!)
		if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).
			Delete(&CurrencyNetwork{}).Error; err != nil {
			return err
		}

		// Flatten and re-insert networks
		var allNets []CurrencyNetwork
		for _, entity := range entities {
			for _, network := range entity.Networks {
				network.Symbol = entity.Symbol
				allNets = append(allNets, network)
			}
		}
		if len(allNets) > 0 {
			if err := tx.Create(&allNets).Error; err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		cr.logger.Infof(ctx, "Error inserting currencies into the database: %v", err)
		return apierrors.NewApiError(apierrors.InternalServerError, err)
	}

	return nil
}

func (cr *currenciesRepository) UpdatePrices(ctx context.Context, currencies []models.Currency) *apierrors.ApiError {
	cr.logger.Infof(ctx, "Updating %d prices in the database", len(currencies))
	if len(currencies) == 0 {
		return nil
	}

	entities := toCurrenciesEntity(currencies)
	if err := cr.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "symbol"}},
		DoUpdates: clause.AssignmentColumns([]string{"price"}),
	}).Create(&entities).Error; err != nil {
		return apierrors.NewApiError(apierrors.InternalServerError, err)
	}

	return nil
}
