package models

import "time"

// AvicolaProfitRecord es la cabecera de un registro de ganancias avícolas.
// Es independiente del módulo AvicolaLedger: aquí se guardan ganancias y
// cuadres de caja propios de esta funcionalidad.
type AvicolaProfitRecord struct {
	ID         uint                `gorm:"primaryKey" json:"id"`
	UserID     uint                `gorm:"index;not null" json:"-"`
	User       AdminUser           `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	RecordedAt time.Time           `gorm:"type:timestamptz;not null;index" json:"recorded_at"`
	Notes      string              `gorm:"size:500" json:"notes"`
	Lines      []AvicolaProfitLine `gorm:"foreignKey:AvicolaProfitRecordID;constraint:OnDelete:CASCADE" json:"lines"`
	CreatedAt  time.Time           `json:"created_at"`
	UpdatedAt  time.Time           `json:"updated_at"`
}

// AvicolaProfitLine es un campo monetario dentro de una sección del registro.
// La sección permite reutilizar una sola tabla para ganancia, caja principal y
// caja auxiliar sin duplicar estructuras.
type AvicolaProfitLine struct {
	ID                    uint      `gorm:"primaryKey" json:"id"`
	AvicolaProfitRecordID uint      `gorm:"index;not null" json:"-"`
	Section               string    `gorm:"size:30;not null;index" json:"section"`
	ConceptName           string    `gorm:"size:120;not null" json:"concept_name"`
	Amount                float64   `gorm:"not null;default:0" json:"amount"`
	OrderIndex            int       `gorm:"not null;default:0" json:"order_index"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}
