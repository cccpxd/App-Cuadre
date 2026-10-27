package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata"
)

// UserCredentials representa una cuenta configurada desde variables de
// entorno. La aplicación no tiene registro público: los usuarios nacen desde
// el entorno y luego inician sesión con su email y contraseña.
type UserCredentials struct {
	Email    string
	Password string
}

type Config struct {
	ServerPort   string
	DatabaseURL  string
	Users        []UserCredentials
	SessionTTL   time.Duration
	CookieName   string
	CookieSecure bool
	Production   bool
	// RunMigrations indica si se ejecuta AutoMigrate al arrancar. Por defecto
	// sí. En Vercel cada arranque en frío lo repetiría; cuando las tablas ya
	// existen se puede apagar con RUN_MIGRATIONS=false.
	RunMigrations bool
	Timezone      string
	Location      *time.Location
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

	users, err := loadUsers()
	if err != nil {
		return Config{}, err
	}

	timezone := getEnv("APP_TIMEZONE", "America/Bogota")
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return Config{}, fmt.Errorf("APP_TIMEZONE no es una zona horaria válida (%q): %w", timezone, err)
	}

	production := isProduction()
	runMigrations := !strings.EqualFold(strings.TrimSpace(getEnv("RUN_MIGRATIONS", "true")), "false")

	return Config{
		ServerPort:    port,
		DatabaseURL:   resolveDatabaseURL(),
		Users:         users,
		SessionTTL:    sessionTTL,
		CookieName:    "cuadre_session",
		CookieSecure:  production,
		Production:    production,
		RunMigrations: runMigrations,
		Timezone:      timezone,
		Location:      location,
	}, nil
}

// loadUsers arma la lista de usuarios desde el entorno. ADMIN_EMAIL y
// ADMIN_PASSWORD son el primer usuario; ADMIN_EMAIL_2 y ADMIN_PASSWORD_2 son
// el segundo. Si un par está incompleto se reporta un error para no arrancar
// con una configuración ambigua.
func loadUsers() ([]UserCredentials, error) {
	pairs := []struct {
		emailKey    string
		passwordKey string
	}{
		{emailKey: "ADMIN_EMAIL", passwordKey: "ADMIN_PASSWORD"},
		{emailKey: "ADMIN_EMAIL_2", passwordKey: "ADMIN_PASSWORD_2"},
	}

	users := make([]UserCredentials, 0, len(pairs))
	seen := make(map[string]struct{}, len(pairs))

	for _, pair := range pairs {
		email := strings.TrimSpace(getEnv(pair.emailKey, ""))
		password := getEnv(pair.passwordKey, "")

		if email == "" && password == "" {
			continue
		}
		if email == "" || password == "" {
			return nil, fmt.Errorf("%s y %s deben configurarse juntas", pair.emailKey, pair.passwordKey)
		}

		normalizedEmail := strings.ToLower(email)
		if _, exists := seen[normalizedEmail]; exists {
			return nil, fmt.Errorf("el email %q está repetido en las variables de entorno", normalizedEmail)
		}
		seen[normalizedEmail] = struct{}{}

		users = append(users, UserCredentials{Email: normalizedEmail, Password: password})
	}

	return users, nil
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

// String no incluye la URL de la base de datos ni las contraseñas, porque
// trae usuario y clave y no debe terminar en los logs.
func (c Config) String() string {
	return fmt.Sprintf("Port=%s Timezone=%s Users=%d", c.ServerPort, c.Timezone, len(c.Users))
}
