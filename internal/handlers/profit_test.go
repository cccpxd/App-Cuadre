package handlers

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"appcuadre/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func testProfitHandler(t *testing.T, location *time.Location) *ProfitHandler {
	t.Helper()

	service, err := services.NewProfitService(&gorm.DB{}, location, func() time.Time {
		return time.Date(2026, time.October, 27, 15, 0, 0, 0, location)
	})
	if err != nil {
		t.Fatalf("no se pudo crear el servicio: %v", err)
	}
	return NewProfitHandler(service, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func queryContext(target string) *gin.Context {
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, target, nil)
	return context
}

func TestResolveQueryRangeUsesBrowserTimezone(t *testing.T) {
	bogota, err := time.LoadLocation("America/Bogota")
	if err != nil {
		t.Fatalf("no se pudo cargar la zona de respaldo: %v", err)
	}

	handler := testProfitHandler(t, bogota)
	context := queryContext("/api/ganancias-avicola?period=semanal&reference_date=2026-10-27&timezone=Pacific/Kiritimati")

	from, to, period, location, err := handler.resolveQueryRange(context)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if period != services.ProfitPeriodWeekly {
		t.Fatalf("periodo = %q, se esperaba %q", period, services.ProfitPeriodWeekly)
	}
	if location.String() != "Pacific/Kiritimati" {
		t.Fatalf("zona = %q, se esperaba Pacific/Kiritimati", location.String())
	}

	// La semana del 27 de octubre de 2026 empieza el lunes 26 en la zona del
	// navegador, sin importar la zona de la aplicación.
	if got := from.In(location).Format("2006-01-02"); got != "2026-10-26" {
		t.Fatalf("fecha inicial = %q, se esperaba 2026-10-26", got)
	}
	if got := to.In(location).Format("2006-01-02"); got != "2026-11-01" {
		t.Fatalf("fecha final = %q, se esperaba 2026-11-01", got)
	}
}

func TestResolveQueryRangeFallsBackToAppTimezone(t *testing.T) {
	bogota, err := time.LoadLocation("America/Bogota")
	if err != nil {
		t.Fatalf("no se pudo cargar la zona de respaldo: %v", err)
	}

	handler := testProfitHandler(t, bogota)
	context := queryContext("/api/ganancias-avicola?from=2026-10-01&to=2026-10-05&timezone=zona-invalida")

	from, to, period, location, err := handler.resolveQueryRange(context)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if period != "" {
		t.Fatalf("periodo = %q, se esperaba vacío en consulta manual", period)
	}
	if location.String() != "America/Bogota" {
		t.Fatalf("zona = %q, se esperaba America/Bogota", location.String())
	}
	if got := from.In(location).Format("2006-01-02 15:04:05"); got != "2026-10-01 00:00:00" {
		t.Fatalf("inicio = %q", got)
	}
	if got := to.In(location).Format("2006-01-02 15:04:05"); got != "2026-10-05 23:59:59" {
		t.Fatalf("fin = %q", got)
	}
}

func TestResolveQueryRangeRejectsIncompleteManualRange(t *testing.T) {
	bogota, err := time.LoadLocation("America/Bogota")
	if err != nil {
		t.Fatalf("no se pudo cargar la zona de respaldo: %v", err)
	}

	handler := testProfitHandler(t, bogota)
	context := queryContext("/api/ganancias-avicola?from=2026-10-01")

	if _, _, _, _, err := handler.resolveQueryRange(context); err == nil {
		t.Fatal("se esperaba un error cuando falta la fecha final")
	}
}
