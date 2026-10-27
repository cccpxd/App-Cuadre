package handlers

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"appcuadre/internal/models"
	"appcuadre/internal/services"

	"github.com/gin-gonic/gin"
)

// AppHandler agrupa los endpoints de cuadres, avícola y configuración. Recibe
// sus servicios por el constructor, así no depende de variables globales.
type AppHandler struct {
	cashBox       *services.CashBoxService
	avicola       *services.AvicolaService
	configuration *services.ConfigurationService
	logger        *slog.Logger
}

func NewAppHandler(
	cashBox *services.CashBoxService,
	avicola *services.AvicolaService,
	configuration *services.ConfigurationService,
	logger *slog.Logger,
) *AppHandler {
	return &AppHandler{cashBox: cashBox, avicola: avicola, configuration: configuration, logger: logger}
}

func RenderDashboard(c *gin.Context) {
	c.HTML(http.StatusOK, "dashboard.tmpl", pageData(c, "Dashboard", gin.H{
		"currentDate": time.Now().Format("2006-01-02"),
	}))
}

func RenderCuadre(c *gin.Context) {
	c.HTML(http.StatusOK, "cuadre.tmpl", pageData(c, "Cuadre de caja", gin.H{
		"entries": services.BuildDefaultCashBoxEntries(),
	}))
}

func RenderAvicola(c *gin.Context) {
	c.HTML(http.StatusOK, "avicola.tmpl", pageData(c, "Avícola", gin.H{}))
}

func RenderConfiguracion(c *gin.Context) {
	c.HTML(http.StatusOK, "configuracion.tmpl", pageData(c, "Configuración", gin.H{}))
}

func RenderHistorial(c *gin.Context) {
	c.HTML(http.StatusOK, "historial.tmpl", pageData(c, "Historial", gin.H{}))
}

func (h *AppHandler) ListCuadres(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		return
	}

	records, err := h.cashBox.ListCashBoxRecords(user.ID)
	if err != nil {
		respondServiceError(c, h.logger, err)
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

func (h *AppHandler) CreateCuadre(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		return
	}

	var record models.CashBoxRecord
	if err := c.ShouldBindJSON(&record); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": invalidBodyMessage})
		return
	}
	record.PreparedBy = strings.TrimSpace(record.PreparedBy)
	if record.PreparedBy == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Indica quién hizo el cuadre"})
		return
	}
	if len([]rune(record.PreparedBy)) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El nombre del responsable no puede pasar de 100 caracteres"})
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

	saved, err := h.cashBox.CreateCashBoxRecord(user.ID, record)
	if err != nil {
		respondServiceError(c, h.logger, err)
		return
	}
	c.JSON(http.StatusCreated, saved)
}

func (h *AppHandler) GetCuadre(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	record, err := h.cashBox.GetCashBoxRecordByID(user.ID, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "cuadre no encontrado"})
		return
	}
	c.JSON(http.StatusOK, record)
}

func (h *AppHandler) ListAvicola(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		return
	}

	selectedDate := strings.TrimSpace(c.Query("date"))
	movements, err := h.avicola.ListAvicolaMovements(user.ID, selectedDate)
	if err != nil {
		respondServiceError(c, h.logger, err)
		return
	}

	if selectedDate == "" {
		selectedDate = time.Now().Format("2006-01-02")
	}

	balance, err := h.avicola.GetAvicolaBalance(user.ID, selectedDate)
	if err != nil {
		respondServiceError(c, h.logger, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"balance":   balance,
		"movements": movements,
		"date":      selectedDate,
	})
}

func (h *AppHandler) CreateAvicolaMovement(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		return
	}

	var movement models.AvicolaLedger
	if err := c.ShouldBindJSON(&movement); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": invalidBodyMessage})
		return
	}

	saved, err := h.avicola.CreateAvicolaMovement(user.ID, movement)
	if err != nil {
		respondServiceError(c, h.logger, err)
		return
	}
	c.JSON(http.StatusCreated, saved)
}

func (h *AppHandler) GetConfiguration(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		return
	}

	cfg, err := h.configuration.GetSystemConfiguration(user.ID)
	if err != nil {
		respondServiceError(c, h.logger, err)
		return
	}
	c.JSON(http.StatusOK, cfg)
}

func (h *AppHandler) UpdateConfiguration(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		return
	}

	var cfg models.Configuration
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": invalidBodyMessage})
		return
	}

	updated, err := h.configuration.SaveSystemConfiguration(user.ID, cfg)
	if err != nil {
		respondServiceError(c, h.logger, err)
		return
	}
	c.JSON(http.StatusOK, updated)
}
