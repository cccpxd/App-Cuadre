package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	ServerPort  string
	DatabaseURL string
}

func Load() (Config, error) {
	port := getEnv("PORT", "8080")
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	return Config{
		ServerPort:  port,
		DatabaseURL: resolveDatabaseURL(),
	}, nil
}

func resolveDatabaseURL() string {
	for _, key := range []string{"DATABASE_URL", "POSTGRES_URL", "POSTGRES_PRISMA_URL", "PGDATABASE_URL"} {
		if value := strings.TrimSpace(getEnv(key, "")); value != "" {
			return value
		}
	}

	if strings.EqualFold(getEnv("APP_ENV", ""), "production") || getEnv("VERCEL", "") != "" {
		return ""
	}

	return "host=db user=postgres password=postgres dbname=appcuadre port=5432 sslmode=disable"
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
