package services

import (
	"fmt"
	"time"

	"appcuadre/internal/database"
	"appcuadre/internal/models"
)

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

func CreateCashBoxRecord(record models.CashBoxRecord) (models.CashBoxRecord, error) {
	if record.RecordDate.IsZero() {
		record.RecordDate = time.Now()
	}

	entries := record.Entries
	record.Entries = nil
	record.TotalValue, record.Difference = CalculateCashBox(entries, record.BaseValue)

	result := database.DB.Create(&record)
	if result.Error != nil {
		return models.CashBoxRecord{}, result.Error
	}

	for i := range entries {
		entries[i].CashBoxRecordID = record.ID
		entries[i].OrderIndex = i
	}

	if len(entries) > 0 {
		if err := database.DB.Create(&entries).Error; err != nil {
			return models.CashBoxRecord{}, err
		}
	}

	record.Entries = entries
	return record, nil
}

func ListCashBoxRecords() ([]models.CashBoxRecord, error) {
	var records []models.CashBoxRecord
	result := database.DB.Preload("Entries").Order("record_date desc, id desc").Find(&records)
	if result.Error != nil {
		return nil, result.Error
	}
	return records, nil
}

func GetCashBoxRecordByID(id uint) (models.CashBoxRecord, error) {
	var record models.CashBoxRecord
	result := database.DB.Preload("Entries").First(&record, id)
	if result.Error != nil {
		return models.CashBoxRecord{}, result.Error
	}
	return record, nil
}

func UpdateCashBoxRecord(record models.CashBoxRecord) error {
	if record.RecordDate.IsZero() {
		record.RecordDate = time.Now()
	}

	entries := record.Entries
	record.Entries = nil
	record.TotalValue, record.Difference = CalculateCashBox(entries, record.BaseValue)

	tx := database.DB.Begin()
	if err := tx.Model(&models.CashBoxRecord{}).Where("id = ?", record.ID).Updates(map[string]interface{}{
		"record_date": record.RecordDate,
		"base_value":  record.BaseValue,
		"total_value": record.TotalValue,
		"difference":  record.Difference,
		"notes":       record.Notes,
	}).Error; err != nil {
		tx.Rollback()
		return err
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
