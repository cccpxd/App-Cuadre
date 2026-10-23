package models

import "time"

// CashBoxEntry representa un registro de valor monetario dentro del cuadre.
type CashBoxEntry struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	CashBoxRecordID uint      `gorm:"not null;index" json:"cash_box_record_id"`
	Name            string    `gorm:"size:100;not null" json:"name"`
	Value           float64   `gorm:"not null;default:0" json:"value"`
	IsBase          bool      `gorm:"default:false" json:"is_base"`
	OrderIndex      int       `gorm:"not null;default:0" json:"order_index"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// CashBoxRecord representa la información consolidada de un cuadre de caja.
type CashBoxRecord struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	RecordDate time.Time      `gorm:"not null" json:"record_date"`
	BaseValue  float64        `gorm:"not null;default:0" json:"base_value"`
	TotalValue float64        `gorm:"not null;default:0" json:"total_value"`
	Difference float64        `gorm:"not null;default:0" json:"difference"`
	Notes      string         `gorm:"type:text" json:"notes"`
	PreparedBy string         `gorm:"size:100;not null;default:''" json:"prepared_by"`
	Entries    []CashBoxEntry `gorm:"foreignKey:CashBoxRecordID" json:"entries"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}
