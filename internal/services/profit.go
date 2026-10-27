package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"appcuadre/internal/models"

	"gorm.io/gorm"
)

type ProfitLineInput struct {
	Section     string  `json:"section"`
	ConceptName string  `json:"concept_name"`
	Amount      float64 `json:"amount"`
	OrderIndex  int     `json:"order_index"`
}

type ProfitRecordInput struct {
	RecordedAt time.Time         `json:"recorded_at"`
	Notes      string            `json:"notes"`
	Lines      []ProfitLineInput `json:"lines"`
}

// ProfitService contiene las reglas y la persistencia de las ganancias
// avícolas. Recibe sus dependencias por constructor para no depender de estado
// global ni de HTTP.
type ProfitService struct {
	db       *gorm.DB
	location *time.Location
	now      func() time.Time
}

func NewProfitService(db *gorm.DB, location *time.Location, now func() time.Time) (*ProfitService, error) {
	if db == nil {
		return nil, fmt.Errorf("profit service: falta la base de datos")
	}
	if location == nil {
		return nil, fmt.Errorf("profit service: falta la zona horaria")
	}
	if now == nil {
		now = time.Now
	}
	return &ProfitService{db: db, location: location, now: now}, nil
}

func (s *ProfitService) Location() *time.Location {
	return s.location
}

func (s *ProfitService) Now() time.Time {
	return s.now().In(s.location)
}

func (s *ProfitService) CreateRecord(ctx context.Context, userID uint, input ProfitRecordInput) (models.AvicolaProfitRecord, error) {
	if userID == 0 {
		return models.AvicolaProfitRecord{}, fmt.Errorf("falta el usuario del registro")
	}

	// Estos errores son de datos inválidos del usuario (se responden con 400).
	// Los de la base de datos de más abajo son internos y no se muestran.
	notes, err := NormalizeProfitNotes(input.Notes)
	if err != nil {
		return models.AvicolaProfitRecord{}, ValidationError(err.Error())
	}

	lines, err := buildProfitLines(input.Lines)
	if err != nil {
		return models.AvicolaProfitRecord{}, ValidationError(err.Error())
	}
	if err := ValidateProfitLines(lines); err != nil {
		return models.AvicolaProfitRecord{}, ValidationError(err.Error())
	}

	recordedAt := input.RecordedAt
	if recordedAt.IsZero() {
		recordedAt = s.now()
	}

	record := models.AvicolaProfitRecord{
		UserID:     userID,
		RecordedAt: recordedAt.UTC(),
		Notes:      notes,
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&record).Error; err != nil {
			return fmt.Errorf("guardar cabecera de ganancia: %w", err)
		}

		for index := range lines {
			lines[index].AvicolaProfitRecordID = record.ID
		}
		if len(lines) > 0 {
			if err := tx.Create(&lines).Error; err != nil {
				return fmt.Errorf("guardar campos de ganancia: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return models.AvicolaProfitRecord{}, err
	}

	record.Lines = lines
	return record, nil
}

func (s *ProfitService) ListRecords(ctx context.Context, userID uint, from, to time.Time) ([]models.AvicolaProfitRecord, error) {
	if userID == 0 {
		return nil, fmt.Errorf("falta el usuario de la consulta")
	}
	if to.Before(from) {
		return nil, fmt.Errorf("el rango de fechas no es válido")
	}

	var records []models.AvicolaProfitRecord
	err := s.db.WithContext(ctx).
		Preload("Lines", func(query *gorm.DB) *gorm.DB {
			return query.Order("section asc, order_index asc, id asc")
		}).
		Where("user_id = ? AND recorded_at >= ? AND recorded_at <= ?", userID, from.UTC(), to.UTC()).
		Order("recorded_at desc, id desc").
		Find(&records).Error
	if err != nil {
		return nil, err
	}
	return records, nil
}

func buildProfitLines(inputs []ProfitLineInput) ([]models.AvicolaProfitLine, error) {
	if len(inputs) == 0 {
		return nil, fmt.Errorf("debes registrar al menos un campo")
	}

	sectionOrder := make(map[string]int, len(ProfitSectionDefinitions()))
	lines := make([]models.AvicolaProfitLine, 0, len(inputs))

	for _, input := range inputs {
		section := strings.TrimSpace(input.Section)
		concept := strings.TrimSpace(input.ConceptName)

		orderIndex := sectionOrder[section]
		sectionOrder[section] = orderIndex + 1

		lines = append(lines, models.AvicolaProfitLine{
			Section:     section,
			ConceptName: concept,
			Amount:      input.Amount,
			OrderIndex:  orderIndex,
		})
	}

	return lines, nil
}

func SummarizeProfitRecords(records []models.AvicolaProfitRecord) ProfitTotals {
	var totals ProfitTotals
	for _, record := range records {
		totals = AddProfitTotals(totals, CalculateProfitTotals(record.Lines))
	}
	return totals
}
