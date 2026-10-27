package services

import (
	"fmt"
	"math"
	"strings"
	"time"

	"appcuadre/internal/models"
)

const (
	ProfitSectionMainCash = "caja_principal"
	ProfitSectionAuxCash  = "caja_auxiliar"

	ProfitPeriodWeekly     = "semanal"
	ProfitPeriodBiweekly   = "quincenal"
	ProfitPeriodMonthly    = "mensual"
	ProfitPeriodLast7Days  = "ultimos7"
	ProfitPeriodLast15Days = "ultimos15"
	ProfitPeriodLast30Days = "ultimos30"

	maxProfitLines         = 150
	maxProfitLinesPerGroup = 50
	maxProfitConceptLength = 120
	maxProfitNotesLength   = 500
)

// ProfitSectionDefinition describe una sección del formulario y su campo
// predeterminado. La ganancia no es una sección: se calcula sumando la caja
// principal y la caja auxiliar. La plantilla y el servicio usan la misma definición para no
// duplicar nombres.
type ProfitSectionDefinition struct {
	Code           string
	Title          string
	DefaultConcept string
}

// ProfitPeriodDefinition describe un botón de consulta automática.
type ProfitPeriodDefinition struct {
	Code  string
	Label string
}

// ProfitTotals trae los totales de las dos cajas y la ganancia, que es
// siempre CajaPrincipal + CajaAuxiliar.
type ProfitTotals struct {
	Ganancia      float64 `json:"ganancia"`
	CajaPrincipal float64 `json:"caja_principal"`
	CajaAuxiliar  float64 `json:"caja_auxiliar"`
}

func ProfitSectionDefinitions() []ProfitSectionDefinition {
	return []ProfitSectionDefinition{
		{Code: ProfitSectionMainCash, Title: "Cuadres caja principal", DefaultConcept: "Cuadre caja principal"},
		{Code: ProfitSectionAuxCash, Title: "Cuadres caja auxiliar", DefaultConcept: "Cuadre caja auxiliar"},
	}
}

func ProfitPeriodDefinitions() []ProfitPeriodDefinition {
	return []ProfitPeriodDefinition{
		{Code: ProfitPeriodWeekly, Label: "Semana actual"},
		{Code: ProfitPeriodBiweekly, Label: "Quincena actual"},
		{Code: ProfitPeriodMonthly, Label: "Mes actual"},
		{Code: ProfitPeriodLast7Days, Label: "Últimos 7 días"},
		{Code: ProfitPeriodLast15Days, Label: "Últimos 15 días"},
		{Code: ProfitPeriodLast30Days, Label: "Últimos 30 días"},
	}
}

func IsValidProfitSection(section string) bool {
	switch section {
	case ProfitSectionMainCash, ProfitSectionAuxCash:
		return true
	default:
		return false
	}
}

// CalculateProfitTotals suma las líneas por sección y calcula la ganancia como
// caja principal + caja auxiliar. Las líneas de la antigua sección "ganancia"
// (registros viejos) no se cuentan. No modifica la entrada.
func CalculateProfitTotals(lines []models.AvicolaProfitLine) ProfitTotals {
	var totals ProfitTotals
	for _, line := range lines {
		switch line.Section {
		case ProfitSectionMainCash:
			totals.CajaPrincipal += line.Amount
		case ProfitSectionAuxCash:
			totals.CajaAuxiliar += line.Amount
		}
	}
	totals.Ganancia = totals.CajaPrincipal + totals.CajaAuxiliar
	return totals
}

func AddProfitTotals(current, addition ProfitTotals) ProfitTotals {
	return ProfitTotals{
		Ganancia:      current.Ganancia + addition.Ganancia,
		CajaPrincipal: current.CajaPrincipal + addition.CajaPrincipal,
		CajaAuxiliar:  current.CajaAuxiliar + addition.CajaAuxiliar,
	}
}

// ValidateProfitLines verifica las reglas de negocio de las líneas ya
// normalizadas. Exige al menos un campo en cada una de las dos secciones.
func ValidateProfitLines(lines []models.AvicolaProfitLine) error {
	if len(lines) == 0 {
		return fmt.Errorf("debes registrar al menos un campo")
	}
	if len(lines) > maxProfitLines {
		return fmt.Errorf("no se pueden registrar más de %d campos en un mismo cuadre", maxProfitLines)
	}

	counts := map[string]int{}
	for index, line := range lines {
		if !IsValidProfitSection(line.Section) {
			return fmt.Errorf("la sección %q no es válida", line.Section)
		}
		if strings.TrimSpace(line.ConceptName) == "" {
			return fmt.Errorf("el nombre del campo %d es obligatorio", index+1)
		}
		if len([]rune(line.ConceptName)) > maxProfitConceptLength {
			return fmt.Errorf("el nombre del campo %d no puede pasar de %d caracteres", index+1, maxProfitConceptLength)
		}
		if math.IsNaN(line.Amount) || math.IsInf(line.Amount, 0) {
			return fmt.Errorf("el valor del campo %q no es válido", line.ConceptName)
		}

		counts[line.Section]++
		if counts[line.Section] > maxProfitLinesPerGroup {
			return fmt.Errorf("la sección %q no puede tener más de %d campos", line.Section, maxProfitLinesPerGroup)
		}
	}

	for _, definition := range ProfitSectionDefinitions() {
		if counts[definition.Code] == 0 {
			return fmt.Errorf("debes registrar al menos un campo en %s", definition.Title)
		}
	}

	return nil
}

func NormalizeProfitNotes(notes string) (string, error) {
	trimmed := strings.TrimSpace(notes)
	if len([]rune(trimmed)) > maxProfitNotesLength {
		return "", fmt.Errorf("las observaciones no pueden pasar de %d caracteres", maxProfitNotesLength)
	}
	return trimmed, nil
}

// ResolveProfitLocation usa la zona enviada por el navegador y cae en la zona
// configurada si el nombre no es válido.
func ResolveProfitLocation(timezone string, fallback *time.Location) *time.Location {
	name := strings.TrimSpace(timezone)
	if name == "" {
		return fallback
	}

	location, err := time.LoadLocation(name)
	if err != nil {
		return fallback
	}
	return location
}

// ResolveProfitPeriodRange calcula un rango inclusivo en la zona indicada.
// Semana: lunes a domingo. Quincena: 1-15 o 16-fin de mes. Mes: calendario.
func ResolveProfitPeriodRange(period string, reference time.Time, location *time.Location) (time.Time, time.Time, error) {
	if location == nil {
		return time.Time{}, time.Time{}, fmt.Errorf("falta la zona horaria")
	}

	referenceDate := time.Date(reference.Year(), reference.Month(), reference.Day(), 0, 0, 0, 0, location)

	var start time.Time
	switch period {
	case ProfitPeriodWeekly:
		offset := (int(referenceDate.Weekday()) + 6) % 7 // lunes = 0
		start = referenceDate.AddDate(0, 0, -offset)
	case ProfitPeriodBiweekly:
		if referenceDate.Day() <= 15 {
			start = time.Date(referenceDate.Year(), referenceDate.Month(), 1, 0, 0, 0, 0, location)
		} else {
			start = time.Date(referenceDate.Year(), referenceDate.Month(), 16, 0, 0, 0, 0, location)
		}
	case ProfitPeriodMonthly:
		start = time.Date(referenceDate.Year(), referenceDate.Month(), 1, 0, 0, 0, 0, location)
	case ProfitPeriodLast7Days:
		start = referenceDate.AddDate(0, 0, -6)
	case ProfitPeriodLast15Days:
		start = referenceDate.AddDate(0, 0, -14)
	case ProfitPeriodLast30Days:
		start = referenceDate.AddDate(0, 0, -29)
	default:
		return time.Time{}, time.Time{}, fmt.Errorf("el periodo %q no es válido", period)
	}

	var end time.Time
	switch period {
	case ProfitPeriodWeekly:
		end = start.AddDate(0, 0, 7).Add(-time.Nanosecond)
	case ProfitPeriodBiweekly:
		if start.Day() == 1 {
			end = time.Date(start.Year(), start.Month(), 16, 0, 0, 0, 0, location).Add(-time.Nanosecond)
		} else {
			end = time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, location).
				AddDate(0, 1, 0).
				Add(-time.Nanosecond)
		}
	case ProfitPeriodMonthly:
		end = start.AddDate(0, 1, 0).Add(-time.Nanosecond)
	default:
		end = referenceDate.AddDate(0, 0, 1).Add(-time.Nanosecond)
	}

	return start, end, nil
}

// ResolveProfitDateRange interpreta dos fechas YYYY-MM-DD como días completos
// en la zona indicada. El límite final es inclusivo.
func ResolveProfitDateRange(fromKey, toKey string, location *time.Location) (time.Time, time.Time, error) {
	if location == nil {
		return time.Time{}, time.Time{}, fmt.Errorf("falta la zona horaria")
	}

	from, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(fromKey), location)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("la fecha inicial no es válida")
	}
	to, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(toKey), location)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("la fecha final no es válida")
	}
	if to.Before(from) {
		return time.Time{}, time.Time{}, fmt.Errorf("la fecha final no puede ser anterior a la inicial")
	}

	return from, to.AddDate(0, 0, 1).Add(-time.Nanosecond), nil
}

// FormatProfitRangeDate devuelve la fecha local del rango en formato ISO.
func FormatProfitRangeDate(value time.Time, location *time.Location) string {
	if location == nil {
		return value.Format("2006-01-02")
	}
	return value.In(location).Format("2006-01-02")
}
