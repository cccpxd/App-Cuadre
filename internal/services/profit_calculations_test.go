package services

import (
	"math"
	"testing"
	"time"

	"appcuadre/internal/models"
)

func testLocation() *time.Location {
	return time.FixedZone("UTC-5", -5*60*60)
}

func testLines() []models.AvicolaProfitLine {
	return []models.AvicolaProfitLine{
		{Section: ProfitSectionMainCash, ConceptName: "Cuadre caja principal", Amount: 500},
		{Section: ProfitSectionAuxCash, ConceptName: "Cuadre caja auxiliar", Amount: 250},
	}
}

func TestCalculateProfitTotals(t *testing.T) {
	lines := append(testLines(),
		models.AvicolaProfitLine{Section: ProfitSectionMainCash, ConceptName: "Otro campo", Amount: -100},
		// Línea de la antigua sección "ganancia": ya no se cuenta.
		models.AvicolaProfitLine{Section: "ganancia", ConceptName: "Ganancia vieja", Amount: 9999},
	)

	totals := CalculateProfitTotals(lines)

	if totals.Ganancia != 650 {
		t.Fatalf("Ganancia = %v, se esperaba 650 (principal + auxiliar)", totals.Ganancia)
	}
	if totals.CajaPrincipal != 400 {
		t.Fatalf("CajaPrincipal = %v, se esperaba 400", totals.CajaPrincipal)
	}
	if totals.CajaAuxiliar != 250 {
		t.Fatalf("CajaAuxiliar = %v, se esperaba 250", totals.CajaAuxiliar)
	}
}

func TestSummarizeProfitRecords(t *testing.T) {
	records := []models.AvicolaProfitRecord{
		{Lines: testLines()},
		{Lines: []models.AvicolaProfitLine{
			{Section: ProfitSectionMainCash, ConceptName: "Otra caja", Amount: 100},
			{Section: ProfitSectionAuxCash, ConceptName: "Otra auxiliar", Amount: 50},
		}},
	}

	totals := SummarizeProfitRecords(records)

	if totals.Ganancia != 900 || totals.CajaPrincipal != 600 || totals.CajaAuxiliar != 300 {
		t.Fatalf("totales inesperados: %+v", totals)
	}
}

func TestValidateProfitLines(t *testing.T) {
	if err := ValidateProfitLines(testLines()); err != nil {
		t.Fatalf("las líneas válidas fueron rechazadas: %v", err)
	}

	missingSection := testLines()[:1]
	if err := ValidateProfitLines(missingSection); err == nil {
		t.Fatal("se esperaba un error al faltar una sección")
	}

	blankConcept := testLines()
	blankConcept[0].ConceptName = "   "
	if err := ValidateProfitLines(blankConcept); err == nil {
		t.Fatal("se esperaba un error por nombre vacío")
	}

	invalidAmount := testLines()
	invalidAmount[0].Amount = math.NaN()
	if err := ValidateProfitLines(invalidAmount); err == nil {
		t.Fatal("se esperaba un error por valor no finito")
	}

	invalidSection := testLines()
	invalidSection[0].Section = "otra"
	if err := ValidateProfitLines(invalidSection); err == nil {
		t.Fatal("se esperaba un error por sección inválida")
	}

	oldGainSection := testLines()
	oldGainSection[0].Section = "ganancia"
	if err := ValidateProfitLines(oldGainSection); err == nil {
		t.Fatal("la sección ganancia ya no debe aceptarse: es la suma de las dos cajas")
	}
}

func TestResolveProfitPeriodRange(t *testing.T) {
	location := testLocation()
	reference := time.Date(2026, time.October, 27, 15, 30, 0, 0, location)

	tests := []struct {
		name      string
		period    string
		wantStart time.Time
		wantEnd   time.Time
	}{
		{
			name:      "semana actual",
			period:    ProfitPeriodWeekly,
			wantStart: time.Date(2026, time.October, 26, 0, 0, 0, 0, location),
			wantEnd:   time.Date(2026, time.November, 1, 23, 59, 59, 999999999, location),
		},
		{
			name:      "quincena actual",
			period:    ProfitPeriodBiweekly,
			wantStart: time.Date(2026, time.October, 16, 0, 0, 0, 0, location),
			wantEnd:   time.Date(2026, time.October, 31, 23, 59, 59, 999999999, location),
		},
		{
			name:      "mes actual",
			period:    ProfitPeriodMonthly,
			wantStart: time.Date(2026, time.October, 1, 0, 0, 0, 0, location),
			wantEnd:   time.Date(2026, time.October, 31, 23, 59, 59, 999999999, location),
		},
		{
			name:      "últimos 7 días",
			period:    ProfitPeriodLast7Days,
			wantStart: time.Date(2026, time.October, 21, 0, 0, 0, 0, location),
			wantEnd:   time.Date(2026, time.October, 27, 23, 59, 59, 999999999, location),
		},
		{
			name:      "últimos 15 días",
			period:    ProfitPeriodLast15Days,
			wantStart: time.Date(2026, time.October, 13, 0, 0, 0, 0, location),
			wantEnd:   time.Date(2026, time.October, 27, 23, 59, 59, 999999999, location),
		},
		{
			name:      "últimos 30 días",
			period:    ProfitPeriodLast30Days,
			wantStart: time.Date(2026, time.September, 28, 0, 0, 0, 0, location),
			wantEnd:   time.Date(2026, time.October, 27, 23, 59, 59, 999999999, location),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			start, end, err := ResolveProfitPeriodRange(test.period, reference, location)
			if err != nil {
				t.Fatalf("error inesperado: %v", err)
			}
			if !start.Equal(test.wantStart) {
				t.Fatalf("inicio = %v, se esperaba %v", start, test.wantStart)
			}
			if !end.Equal(test.wantEnd) {
				t.Fatalf("fin = %v, se esperaba %v", end, test.wantEnd)
			}
		})
	}

	if _, _, err := ResolveProfitPeriodRange("periodo-invalido", reference, location); err == nil {
		t.Fatal("se esperaba un error con un periodo inválido")
	}
}

func TestResolveProfitDateRange(t *testing.T) {
	location := testLocation()

	from, to, err := ResolveProfitDateRange("2026-10-01", "2026-10-05", location)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if !from.Equal(time.Date(2026, time.October, 1, 0, 0, 0, 0, location)) {
		t.Fatalf("inicio inesperado: %v", from)
	}
	if !to.Equal(time.Date(2026, time.October, 5, 23, 59, 59, 999999999, location)) {
		t.Fatalf("fin inesperado: %v", to)
	}

	if _, _, err := ResolveProfitDateRange("2026-10-05", "2026-10-01", location); err == nil {
		t.Fatal("se esperaba un error cuando la fecha final es anterior")
	}
}

func TestResolveProfitLocation(t *testing.T) {
	fallback := testLocation()

	if location := ResolveProfitLocation("UTC", fallback); location.String() != "UTC" {
		t.Fatalf("se esperaba UTC, se obtuvo %s", location)
	}
	if location := ResolveProfitLocation("zona-que-no-existe", fallback); location != fallback {
		t.Fatalf("se esperaba la zona de respaldo, se obtuvo %s", location)
	}
}
