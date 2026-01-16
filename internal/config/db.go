package config

import (
    "strconv"
)

type DBConfig struct {
    User           string
    Password       string
    Driver         string
    Name           string
    Host           string
    Port           string
    SslMode        string
    DBMaxOpenConns int
    DBMaxIdleConns int
    DBConnMaxLife  int
    AppEnv         string
}

func (c *DBConfig) GetDSN() string {
    if c.Driver == "postgres" {
        return "host=" + c.Host + " user=" + c.User + " password=" + c.Password + " dbname=" + c.Name + " port=" + c.Port + " sslmode=" + c.SslMode
    }
    return c.Name // For sqlite
}

func LoadDBConfig() *DBConfig {
    maxOpenConns, _ := strconv.Atoi(getEnv("DB_MAX_OPEN_CONNS", "10"))
    maxIdleConns, _ := strconv.Atoi(getEnv("DB_MAX_IDLE_CONNS", "5"))
    connMaxLife, _ := strconv.Atoi(getEnv("DB_CONN_MAX_LIFE", "360"))

    return &DBConfig{
        User:           getEnv("DB_USER", "admin"),
        Password:       getEnv("DB_PASSWORD", "admin"),
        Driver:         getEnv("DB_DRIVER", "postgres"),
        Name:           getEnv("DB_NAME", "nft_marketplace"),
        Host:           getEnv("DB_HOST", "localhost"),
        Port:           getEnv("DB_PORT", "5432"),
        SslMode:        getEnv("DB_SSL", "disable"),
        DBMaxOpenConns: maxOpenConns,
        DBMaxIdleConns: maxIdleConns,
        DBConnMaxLife:  connMaxLife,
        AppEnv:         getEnv("GIN_MODE", "debug"),
    }
}