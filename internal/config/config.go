package config

import (
    "os"
    "github.com/joho/godotenv"
)

type Config struct {
    DB       *DBConfig
    HTTP     *HTTPConfig
    Ethereum *EthConfig
    LogLevel string
}

func Load() *Config {
    _ = godotenv.Overload() // Ensure .env values overwrite existing env vars
    cfg := &Config{
        DB:       LoadDBConfig(),
        HTTP:     LoadHTTPConfig(),
        Ethereum: LoadEthConfig(),
        LogLevel: getEnv("LOG_LEVEL", "info"),
    }
    return cfg
}

// Helper to read environment variables
func getEnv(key, fallback string) string {
    if value, ok := os.LookupEnv(key); ok {
        return value
    }
    return fallback
}