package services

import (
	"errors"
	"fmt"
	"time"

	"appcuadre/internal/models"

	"gorm.io/gorm"
)

// CashBoxService guarda y consulta los cuadres de caja.
type CashBoxService struct {
	db *gorm.DB
}

func NewCashBoxService(db *gorm.DB) (*CashBoxService, error) {
	if db == nil {
		return nil, errors.New("cashbox service: falta la base de datos")
	}
	return &CashBoxService{db: db}, nil
}

func CalculateCashBox(entries []models.CashBoxEntry, baseValue float64) (float64, float64) {
	total := 0.0
	for _, entry := range entries {
		total += entry.Value
	}
	return total, total - baseValue
}

func BuildDefaultCashBoxEntries() []models.CashBoxEntry {
	names := []string{"Bancolombia", "BEMovil", "Punto Red", "Giros", "PTM", "PTM 2", "Banco de Bogotá", "Avícola", "Monedas", "Efectivo"}
	entries := make([]models.CashBoxEntry, 0, len(names))
	for index, name := range names {
		entries = append(entries, models.CashBoxEntry{
			Name:       name,
			Value:      0,
			IsBase:     index < len(names),
			OrderIndex: index,
		})
	}
	return entries
}

func (s *CashBoxService) CreateCashBoxRecord(userID uint, record models.CashBoxRecord) (models.CashBoxRecord, error) {
	if userID == 0 {
		return models.CashBoxRecord{}, fmt.Errorf("falta el usuario del cuadre")
	}
	if record.RecordDate.IsZero() {
		record.RecordDate = time.Now()
	}

	record.UserID = userID
	entries := record.Entries
	record.Entries = nil
	record.TotalValue, record.Difference = CalculateCashBox(entries, record.BaseValue)

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&record).Error; err != nil {
			return err
		}

		for i := range entries {
			entries[i].CashBoxRecordID = record.ID
			entries[i].OrderIndex = i
		}

		if len(entries) > 0 {
			if err := tx.Create(&entries).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return models.CashBoxRecord{}, err
	}

	record.Entries = entries
	return record, nil
}

func (s *CashBoxService) ListCashBoxRecords(userID uint) ([]models.CashBoxRecord, error) {
	var records []models.CashBoxRecord
	result := s.db.
		Preload("Entries").
		Where("user_id = ?", userID).
		Order("record_date desc, id desc").
		Find(&records)
	if result.Error != nil {
		return nil, result.Error
	}
	return records, nil
}

func (s *CashBoxService) GetCashBoxRecordByID(userID, id uint) (models.CashBoxRecord, error) {
	var record models.CashBoxRecord
	result := s.db.
		Preload("Entries").
		Where("user_id = ?", userID).
		First(&record, id)
	if result.Error != nil {
		return models.CashBoxRecord{}, result.Error
	}
	return record, nil
}

func (s *CashBoxService) UpdateCashBoxRecord(userID uint, record models.CashBoxRecord) error {
	if userID == 0 {
		return fmt.Errorf("falta el usuario del cuadre")
	}
	if record.RecordDate.IsZero() {
		record.RecordDate = time.Now()
	}

	entries := record.Entries
	record.Entries = nil
	record.TotalValue, record.Difference = CalculateCashBox(entries, record.BaseValue)

	tx := s.db.Begin()
	updateResult := tx.Model(&models.CashBoxRecord{}).
		Where("id = ? AND user_id = ?", record.ID, userID).
		Updates(map[string]interface{}{
			"record_date": record.RecordDate,
			"base_value":  record.BaseValue,
			"total_value": record.TotalValue,
			"difference":  record.Difference,
			"notes":       record.Notes,
		})
	if updateResult.Error != nil {
		tx.Rollback()
		return updateResult.Error
	}
	if updateResult.RowsAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("cuadre no encontrado")
	}

	if err := tx.Where("cash_box_record_id = ?", record.ID).Delete(&models.CashBoxEntry{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	for i := range entries {
		entries[i].CashBoxRecordID = record.ID
		entries[i].OrderIndex = i
	}

	if len(entries) > 0 {
		if err := tx.Create(&entries).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func ValidateCashBoxEntry(entry models.CashBoxEntry) error {
	if entry.Name == "" {
		return fmt.Errorf("el nombre de la aplicación es obligatorio")
	}
	if entry.Value < 0 {
		return fmt.Errorf("el valor no puede ser negativo")
	}
	return nil
}
