package db

import (
	"database/sql"
	"fmt"
	"path/filepath"

	"time"

	_ "github.com/lib/pq"
	"github.com/user/nft-marketplace/internal/config"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"github.com/user/nft-marketplace/internal/core"
)

func autoMigrate(db *gorm.DB, models ...interface{}) {
	for _, m := range models {
		if err := db.AutoMigrate(m); err != nil {
			logrus.Fatalf("Failed to migrate %T: %v", m, err)
		}
	}
}

func InitDB(cfg *config.DBConfig) *gorm.DB {
	sqlDB := setupDB(cfg)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		logrus.Fatalf("Failed to open GORM DB: %v", err)
	}

	if err := gormDB.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`).Error; err != nil {
		logrus.Fatalf("Failed to enable uuid-ossp extension: %v", err)
	}

	autoMigrate(
		gormDB, &core.User{}, &core.Collection{}, &core.NFT{}, &core.Listing{}, &core.Order{},
	)

	if cfg.AppEnv == "debug" {
		gormDB = gormDB.Debug()
		logrus.Info("GORM debug mode enabled")
	}

	return gormDB
}

func setupDB(cfg *config.DBConfig) *sql.DB {
	dataSourceName := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SslMode,
	)

	if cfg.SslMode != "disable" {
		authPemPath := filepath.Join(".", "config", "auth.pem")
		dataSourceName += fmt.Sprintf(" sslrootcert=%s", authPemPath)
	}

	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		logrus.Fatalf("Failed to connect to DB: %v", err)
	}

	db.SetMaxOpenConns(cfg.DBMaxOpenConns)
	db.SetMaxIdleConns(cfg.DBMaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.DBConnMaxLife) * time.Second)

	return db
}
