package database

import (
	"fmt"
	"log"
	"sync"

	"github.com/alimarzban99/video-processor-service/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db   *gorm.DB
	once sync.Once
)

func Init() error {
	var initErr error

	once.Do(func() {
		cfg := config.Cfg.Database

		dsn := fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Asia%%2FTehran",
			cfg.User,
			cfg.Password,
			cfg.Host,
			cfg.Port,
			cfg.Name,
		)

		gormLogger := logger.Default.LogMode(parseLogLevel(cfg.LogLevel))

		db, initErr = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: gormLogger,
		})
		if initErr != nil {
			return
		}

		sqlDB, err := db.DB()
		if err != nil {
			initErr = err
			return
		}

		if err = sqlDB.Ping(); err != nil {
			initErr = err
			return
		}

		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
		sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

		log.Printf("✅ MySQL connected (%s:%s/%s)",
			cfg.Host,
			cfg.Port,
			cfg.Name,
		)
	})

	return initErr
}

func DB() *gorm.DB {
	return db
}

func Close() error {
	if db == nil {
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	log.Println("🛑 MySQL connection closed")
	return sqlDB.Close()
}

func parseLogLevel(level string) logger.LogLevel {
	switch level {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "warn":
		return logger.Warn
	default:
		return logger.Info
	}
}
