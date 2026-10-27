package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"appcuadre/internal/models"
	"appcuadre/internal/services"

	"github.com/gin-gonic/gin"
)

const maxProfitBodyBytes = 64 << 10

type ProfitHandler struct {
	service *services.ProfitService
	logger  *slog.Logger
}

func NewProfitHandler(service *services.ProfitService, logger *slog.Logger) *ProfitHandler {
	return &ProfitHandler{service: service, logger: logger}
}

type profitRecordResponse struct {
	models.AvicolaProfitRecord
	Totals services.ProfitTotals `json:"totals"`
}

type profitRangeResponse struct {
	From     string `json:"from"`
	To       string `json:"to"`
	FromDate string `json:"from_date"`
	ToDate   string `json:"to_date"`
	Timezone string `json:"timezone"`
	Period   string `json:"period,omitempty"`
}

type profitSummaryResponse struct {
	Range   profitRangeResponse    `json:"range"`
	Totals  services.ProfitTotals  `json:"totals"`
	Records []profitRecordResponse `json:"records"`
}

func (h *ProfitHandler) RenderGanancias(c *gin.Context) {
	c.HTML(http.StatusOK, "ganancias_avicola.tmpl", pageData(c, "Ganancias Avícola", gin.H{
		"sections": services.ProfitSectionDefinitions(),
		"nowLocal": h.service.Now().Format("2006-01-02T15:04"),
	}))
}

func (h *ProfitHandler) RenderConsulta(c *gin.Context) {
	c.HTML(http.StatusOK, "consulta_ganancias.tmpl", pageData(c, "Consulta Ganancia Avícola", gin.H{
		"periods":     services.ProfitPeriodDefinitions(),
		"today":       h.service.Now().Format("2006-01-02"),
		"appTimezone": h.service.Location().String(),
	}))
}

func (h *ProfitHandler) Create(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxProfitBodyBytes)

	var input services.ProfitRecordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "los datos del cuadre no son válidos"})
		return
	}

	record, err := h.service.CreateRecord(c.Request.Context(), user.ID, input)
	if err != nil {
		respondServiceError(c, h.logger, err)
		return
	}

	c.JSON(http.StatusCreated, buildProfitRecordResponse(record))
}

func (h *ProfitHandler) List(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		return
	}

	from, to, period, location, err := h.resolveQueryRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	records, err := h.service.ListRecords(c.Request.Context(), user.ID, from, to)
	if err != nil {
		respondServiceError(c, h.logger, err)
		return
	}

	responseRecords := make([]profitRecordResponse, 0, len(records))
	for _, record := range records {
		responseRecords = append(responseRecords, buildProfitRecordResponse(record))
	}

	c.JSON(http.StatusOK, profitSummaryResponse{
		Range: profitRangeResponse{
			From:     from.In(location).Format(time.RFC3339Nano),
			To:       to.In(location).Format(time.RFC3339Nano),
			FromDate: from.In(location).Format("2006-01-02"),
			ToDate:   to.In(location).Format("2006-01-02"),
			Timezone: location.String(),
			Period:   period,
		},
		Totals:  services.SummarizeProfitRecords(records),
		Records: responseRecords,
	})
}

func (h *ProfitHandler) resolveQueryRange(c *gin.Context) (time.Time, time.Time, string, *time.Location, error) {
	location := services.ResolveProfitLocation(c.Query("timezone"), h.service.Location())

	fromKey := strings.TrimSpace(c.Query("from"))
	toKey := strings.TrimSpace(c.Query("to"))
	if fromKey != "" || toKey != "" {
		if fromKey == "" || toKey == "" {
			return time.Time{}, time.Time{}, "", location, errors.New("debes indicar la fecha inicial y la final")
		}
		from, to, err := services.ResolveProfitDateRange(fromKey, toKey, location)
		if err != nil {
			return time.Time{}, time.Time{}, "", location, err
		}
		return from, to, "", location, nil
	}

	period := strings.TrimSpace(c.Query("period"))
	if period == "" {
		period = services.ProfitPeriodWeekly
	}

	// La fecha de referencia se interpreta en la zona del navegador: "hoy"
	// debe ser el día del usuario, no el del servidor.
	reference := h.service.Now().In(location)
	referenceKey := strings.TrimSpace(c.Query("reference_date"))
	if referenceKey != "" {
		parsed, err := time.ParseInLocation("2006-01-02", referenceKey, location)
		if err != nil {
			return time.Time{}, time.Time{}, "", location, errors.New("la fecha de referencia no es válida")
		}
		reference = parsed
	}

	from, to, err := services.ResolveProfitPeriodRange(period, reference, location)
	if err != nil {
		return time.Time{}, time.Time{}, "", location, err
	}
	return from, to, period, location, nil
}

func buildProfitRecordResponse(record models.AvicolaProfitRecord) profitRecordResponse {
	return profitRecordResponse{
		AvicolaProfitRecord: record,
		Totals:              services.CalculateProfitTotals(record.Lines),
	}
}
