package db

import (
    "database/sql"
    "fmt"
    "time"

    "github.com/sirupsen/logrus"
    "github.com/user/nft-marketplace/internal/config"
    "github.com/user/nft-marketplace/internal/core"

    _ "github.com/jackc/pgx/v5/stdlib"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
    "gorm.io/gorm/schema"
)

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

    err = gormDB.AutoMigrate(
        &core.User{},
        &core.Collection{},
        &core.NFT{},
        &core.Listing{},
        &core.Order{},
    )
    if err != nil {
        logrus.Fatalf("Failed to run auto-migrations: %v", err)
    }

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

    db, err := sql.Open("pgx", dataSourceName)
    if err != nil {
        logrus.Fatalf("Failed to connect to DB: %v", err)
    }

    db.SetMaxOpenConns(cfg.DBMaxOpenConns)
    db.SetMaxIdleConns(cfg.DBMaxIdleConns)
    db.SetConnMaxLifetime(time.Duration(cfg.DBConnMaxLife) * time.Second)

    return db
}