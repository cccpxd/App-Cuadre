package main

import (
    "log"

    "appcuadre/internal/config"
    "appcuadre/internal/database"
    "appcuadre/internal/handlers"
    "appcuadre/internal/models"

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
    ); err != nil {
        log.Fatal(err)
    }

    r := gin.Default()

    r.Static("/static", "./web/static")
    renderer, err := handlers.NewTemplateRenderer("./web/templates")
    if err != nil {
        log.Fatal(err)
    }
    r.HTMLRender = renderer

    r.GET("/", handlers.RenderDashboard)
    r.GET("/cuadre", handlers.RenderCuadre)
    r.GET("/avicola", handlers.RenderAvicola)
    r.GET("/configuracion", handlers.RenderConfiguracion)
    r.GET("/historial", handlers.RenderHistorial)

    r.GET("/api/cuadres", handlers.ListCuadres)
    r.POST("/api/cuadres", handlers.CreateCuadre)
    r.GET("/api/cuadres/:id", handlers.GetCuadre)

    r.GET("/api/avicola", handlers.ListAvicola)
    r.POST("/api/avicola", handlers.CreateAvicolaMovement)

    r.GET("/api/configuracion", handlers.GetConfiguration)
    r.PUT("/api/configuracion", handlers.UpdateConfiguration)

    if err := r.Run(cfg.ServerPort); err != nil {
        log.Fatal(err)
    }
}
