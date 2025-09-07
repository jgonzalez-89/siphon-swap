package db

import (
	"cryptoswap/internal/lib/logger"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	Schema   string
}

func NewGorm(config Config, logger logger.Logger) (*gorm.DB, error) {
	uri := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", config.User, config.Password, config.Host,
		config.Port, config.Schema)

	return gorm.Open(mysql.Open(uri), &gorm.Config{
		Logger: gormlogger.New(logger, gormlogger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  gormlogger.Warn,
			IgnoreRecordNotFoundError: false,
			Colorful:                  true,
		}),
	})
}
