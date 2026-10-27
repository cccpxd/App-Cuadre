package database

import (
	"fmt"

	"gorm.io/gorm"
)

// BackfillUserOwnership asigna al usuario indicado los registros históricos
// que quedaron sin dueño al agregar la columna user_id. Es idempotente: solo
// toca filas con user_id nulo o cero.
func BackfillUserOwnership(db *gorm.DB, userID uint) (map[string]int64, error) {
	if db == nil {
		return nil, fmt.Errorf("backfill: falta la base de datos")
	}
	if userID == 0 {
		return nil, fmt.Errorf("backfill: falta el usuario dueño")
	}

	tables := []string{"cash_box_records", "avicola_ledgers", "configurations"}
	updated := make(map[string]int64, len(tables))

	for _, table := range tables {
		result := db.Table(table).
			Where("user_id IS NULL OR user_id = 0").
			Update("user_id", userID)
		if result.Error != nil {
			return updated, fmt.Errorf("asignar dueño en %s: %w", table, result.Error)
		}
		updated[table] = result.RowsAffected
	}

	return updated, nil
}
