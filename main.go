package main

import (
	"context"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"

	"appcuadre/internal/config"
	"appcuadre/internal/database"
	"appcuadre/internal/handlers"
	"appcuadre/internal/middleware"
	"appcuadre/internal/models"
	"appcuadre/internal/services"
	"appcuadre/web"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.AutoMigrate(
		&models.CashBoxRecord{},
		&models.CashBoxEntry{},
		&models.AvicolaLedger{},
		&models.Configuration{},
		&models.AdminUser{},
		&models.Session{},
	); err != nil {
		log.Fatal(err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	authService, err := services.NewAuthService(db, cfg.SessionTTL)
	if err != nil {
		log.Fatal(err)
	}

	if cfg.AdminEmail != "" || cfg.AdminPassword != "" {
		created, err := authService.EnsureAdmin(context.Background(), cfg.AdminEmail, cfg.AdminPassword)
		if err != nil {
			log.Fatalf("no se pudo preparar el administrador: %v", err)
		}
		if created {
			logger.Info("administrador creado")
		}
	} else {
		logger.Warn("ADMIN_EMAIL y ADMIN_PASSWORD no están configurados; no se crea ningún administrador")
	}

	r := gin.Default()
	if err := r.SetTrustedProxies(nil); err != nil {
		log.Fatal(err)
	}
	r.Use(middleware.SecurityHeaders())

	staticFiles, err := fs.Sub(web.Files, "static")
	if err != nil {
		log.Fatal(err)
	}
	r.StaticFS("/static", http.FS(staticFiles))

	renderer, err := handlers.NewTemplateRenderer(web.Files, "templates")
	if err != nil {
		log.Fatal(err)
	}
	r.HTMLRender = renderer

	authHandler := handlers.NewAuthHandler(authService, logger, handlers.CookieConfig{
		Name:     cfg.CookieName,
		TTL:      cfg.SessionTTL,
		Secure:   cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})

	// Rutas públicas: solo el inicio de sesión (y los estáticos de arriba).
	r.GET("/login", authHandler.LoginPage)
	r.POST("/login", authHandler.Login)
	r.POST("/logout", authHandler.Logout)

	// Todo lo demás exige una sesión válida.
	protected := r.Group("/")
	protected.Use(middleware.RequireAuth(authService, cfg.CookieName, logger))

	protected.GET("/", handlers.RenderDashboard)
	protected.GET("/cuadre", handlers.RenderCuadre)
	protected.GET("/avicola", handlers.RenderAvicola)
	protected.GET("/configuracion", handlers.RenderConfiguracion)
	protected.GET("/historial", handlers.RenderHistorial)

	protected.GET("/api/cuadres", handlers.ListCuadres)
	protected.POST("/api/cuadres", handlers.CreateCuadre)
	protected.GET("/api/cuadres/:id", handlers.GetCuadre)

	protected.GET("/api/avicola", handlers.ListAvicola)
	protected.POST("/api/avicola", handlers.CreateAvicolaMovement)

	protected.GET("/api/configuracion", handlers.GetConfiguration)
	protected.PUT("/api/configuracion", handlers.UpdateConfiguration)

	if err := r.Run(cfg.ServerPort); err != nil {
		log.Fatal(err)
	}
}
