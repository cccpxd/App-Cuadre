package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ServerPort    string
	DatabaseURL   string
	AdminEmail    string
	AdminPassword string
	SessionTTL    time.Duration
	CookieName    string
	CookieSecure  bool
}

func Load() (Config, error) {
	port := getEnv("PORT", "8080")
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	sessionTTL, err := parseSessionTTL(getEnv("SESSION_TTL_HOURS", "12"))
	if err != nil {
		return Config{}, err
	}

	return Config{
		ServerPort:    port,
		DatabaseURL:   resolveDatabaseURL(),
		AdminEmail:    getEnv("ADMIN_EMAIL", ""),
		AdminPassword: getEnv("ADMIN_PASSWORD", ""),
		SessionTTL:    sessionTTL,
		CookieName:    "cuadre_session",
		CookieSecure:  isProduction(),
	}, nil
}

func parseSessionTTL(raw string) (time.Duration, error) {
	hours, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || hours <= 0 || hours > 720 {
		return 0, fmt.Errorf("SESSION_TTL_HOURS debe ser un número de horas entre 1 y 720, y recibí %q", raw)
	}
	return time.Duration(hours) * time.Hour, nil
}

// isProduction indica si la app corre en Vercel o con APP_ENV=production.
// En producción la cookie de sesión solo viaja por HTTPS.
func isProduction() bool {
	return strings.EqualFold(getEnv("APP_ENV", ""), "production") || getEnv("VERCEL", "") != ""
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
