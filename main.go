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
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// En producción Gin corre en modo release (sin los logs de depuración).
	if cfg.Production {
		gin.SetMode(gin.ReleaseMode)
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	if cfg.RunMigrations {
		if err := db.AutoMigrate(
			&models.AdminUser{},
			&models.Session{},
			&models.CashBoxRecord{},
			&models.CashBoxEntry{},
			&models.AvicolaLedger{},
			&models.Configuration{},
			&models.AvicolaProfitRecord{},
			&models.AvicolaProfitLine{},
		); err != nil {
			log.Fatal(err)
		}
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	authService, err := services.NewAuthService(db, cfg.SessionTTL)
	if err != nil {
		log.Fatal(err)
	}

	prepareUsers(logger, db, authService, cfg.Users)

	profitService, err := services.NewProfitService(db, cfg.Location, nil)
	if err != nil {
		log.Fatal(err)
	}

	cashBoxService, err := services.NewCashBoxService(db)
	if err != nil {
		log.Fatal(err)
	}

	avicolaService, err := services.NewAvicolaService(db)
	if err != nil {
		log.Fatal(err)
	}

	configurationService, err := services.NewConfigurationService(db)
	if err != nil {
		log.Fatal(err)
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
	profitHandler := handlers.NewProfitHandler(profitService, logger)
	appHandler := handlers.NewAppHandler(cashBoxService, avicolaService, configurationService, logger)

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
	protected.GET("/ganancias-avicola", profitHandler.RenderGanancias)
	protected.GET("/consulta-ganancias", profitHandler.RenderConsulta)

	protected.GET("/api/cuadres", appHandler.ListCuadres)
	protected.POST("/api/cuadres", appHandler.CreateCuadre)
	protected.GET("/api/cuadres/:id", appHandler.GetCuadre)

	protected.GET("/api/avicola", appHandler.ListAvicola)
	protected.POST("/api/avicola", appHandler.CreateAvicolaMovement)

	protected.GET("/api/configuracion", appHandler.GetConfiguration)
	protected.PUT("/api/configuracion", appHandler.UpdateConfiguration)

	protected.GET("/api/ganancias-avicola", profitHandler.List)
	protected.POST("/api/ganancias-avicola", profitHandler.Create)

	if err := r.Run(cfg.ServerPort); err != nil {
		log.Fatal(err)
	}
}

// prepareUsers crea los usuarios configurados y asigna los datos históricos al
// primer usuario. Estos pasos son idempotentes para que el arranque pueda
// repetirse sin duplicar cuentas ni reasignar registros ya existentes.
func prepareUsers(logger *slog.Logger, db *gorm.DB, authService *services.AuthService, users []config.UserCredentials) {
	if len(users) == 0 {
		logger.Warn("no hay usuarios configurados; define ADMIN_EMAIL/ADMIN_PASSWORD y, si aplica, ADMIN_EMAIL_2/ADMIN_PASSWORD_2")
		return
	}

	for _, credentials := range users {
		created, err := authService.EnsureUser(context.Background(), credentials.Email, credentials.Password)
		if err != nil {
			log.Fatalf("no se pudo preparar el usuario %s: %v", credentials.Email, err)
		}
		if created {
			logger.Info("usuario creado", "email", credentials.Email)
		}
	}

	var primaryUser models.AdminUser
	if err := db.Where("email = ?", users[0].Email).First(&primaryUser).Error; err != nil {
		log.Fatalf("no se pudo encontrar el usuario principal: %v", err)
	}

	updated, err := database.BackfillUserOwnership(db, primaryUser.ID)
	if err != nil {
		log.Fatalf("no se pudo asignar los datos existentes: %v", err)
	}
	for table, count := range updated {
		if count > 0 {
			logger.Info("datos históricos asignados", "tabla", table, "registros", count, "usuario", primaryUser.Email)
		}
	}
}
