package pg

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
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
	dsn := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s TimeZone=%s",
		cfg.Host,
		cfg.Port,
		cfg.Username,
		cfg.DBName,
		cfg.Password,
		cfg.SSLMode,
		cfg.TimeZone,
	)
	dialector := postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	})
    return dialector
}

// AutoMigrate 在开启了自动迁移标志时执行模型结构迁移；
// 当未开启或未提供模型时直接返回 nil，避免误迁移与不必要的操作。
func (d *DB) AutoMigrate(models ...any) error {
    // d 为空或未开启自动迁移，直接跳过
    if d == nil || !d.autoMigrate {
        return nil
    }
    // 未提供任何模型，视为无事可做
    if len(models) == 0 {
        return nil
    }
    return d.DB.AutoMigrate(models...)
}
