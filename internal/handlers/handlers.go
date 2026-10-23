package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"appcuadre/internal/database"
	"appcuadre/internal/models"
	"appcuadre/internal/services"

	"github.com/gin-gonic/gin"
)

func RenderDashboard(c *gin.Context) {
	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"title":       "Dashboard",
		"currentDate": time.Now().Format("2006-01-02"),
	})
}

func RenderCuadre(c *gin.Context) {
	c.HTML(http.StatusOK, "cuadre.html", gin.H{
		"title":   "Cuadre de caja",
		"entries": services.BuildDefaultCashBoxEntries(),
	})
}

func RenderAvicola(c *gin.Context) {
	c.HTML(http.StatusOK, "avicola.html", gin.H{
		"title": "Avícola",
	})
}

func RenderConfiguracion(c *gin.Context) {
	c.HTML(http.StatusOK, "configuracion.html", gin.H{
		"title": "Configuración",
	})
}

func RenderHistorial(c *gin.Context) {
	c.HTML(http.StatusOK, "historial.html", gin.H{
		"title": "Historial",
	})
}

func ListCuadres(c *gin.Context) {
	records, err := services.ListCashBoxRecords()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	selectedDate := strings.TrimSpace(c.Query("date"))
	if selectedDate == "" {
		c.JSON(http.StatusOK, records)
		return
	}

	filtered := make([]models.CashBoxRecord, 0, len(records))
	for _, record := range records {
		if record.RecordDate.Format("2006-01-02") == selectedDate {
			filtered = append(filtered, record)
		}
	}

	c.JSON(http.StatusOK, filtered)
}

func CreateCuadre(c *gin.Context) {
	var record models.CashBoxRecord
	if err := c.ShouldBindJSON(&record); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(record.Entries) == 0 {
		record.Entries = services.BuildDefaultCashBoxEntries()
	}

	for _, entry := range record.Entries {
		if err := services.ValidateCashBoxEntry(entry); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	saved, err := services.CreateCashBoxRecord(record)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, saved)
}

func GetCuadre(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	record, err := services.GetCashBoxRecordByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "cuadre no encontrado"})
		return
	}
	c.JSON(http.StatusOK, record)
}

func ListAvicola(c *gin.Context) {
	selectedDate := strings.TrimSpace(c.Query("date"))
	movements, err := services.ListAvicolaMovements(selectedDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if selectedDate == "" {
		selectedDate = time.Now().Format("2006-01-02")
	}

	balance, err := services.GetAvicolaBalance(selectedDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"balance":   balance,
		"movements": movements,
		"date":      selectedDate,
	})
}

func CreateAvicolaMovement(c *gin.Context) {
	var movement models.AvicolaLedger
	if err := c.ShouldBindJSON(&movement); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	saved, err := services.CreateAvicolaMovement(movement)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, saved)
}

func GetConfiguration(c *gin.Context) {
	cfg, err := services.GetSystemConfiguration()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cfg)
}

func UpdateConfiguration(c *gin.Context) {
	var cfg models.Configuration
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := services.SaveSystemConfiguration(cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func init() {
	_ = database.DB
}
