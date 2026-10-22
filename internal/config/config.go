package config

import (
    "fmt"
    "os"
)

type Config struct {
    ServerPort   string
    DatabaseURL  string
}

func Load() (Config, error) {
    cfg := Config{
        ServerPort:  getEnv("PORT", ":8080"),
        DatabaseURL: getEnv("DATABASE_URL", "host=localhost user=postgres password=postgres dbname=appcuadre port=5432 sslmode=disable"),
    }
    return cfg, nil
}

func getEnv(key, fallback string) string {
    if value, ok := os.LookupEnv(key); ok && value != "" {
        return value
    }
    return fallback
}

func (c Config) String() string {
    return fmt.Sprintf("Port=%s DatabaseURL=%s", c.ServerPort, c.DatabaseURL)
}
