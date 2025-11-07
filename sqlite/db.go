package sqlite

import (
	"fmt"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DB struct {
	*gorm.DB
	autoMigrate bool
}

func NewDB(config *Config) (*DB, error) {

	dial := dial(config)
	db, err := gorm.Open(dial, &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if config.Debug {
		db = db.Debug()
	}
	sqlDB, err := db.DB()
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Second * time.Duration(config.MaxLifetime))

	return &DB{DB: db, autoMigrate: config.AutoMigrate}, nil
}

func dial(cfg *Config) gorm.Dialector {
	dsn := cfg.DatabasePath
	if dsn == "" {
		dsn = ":memory:"
	}

	dialector := sqlite.Open(dsn)
	return dialector
}