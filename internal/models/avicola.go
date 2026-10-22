package models

import "time"

// AvicolaLedger representa el historial de saldos del módulo avícola.
type AvicolaLedger struct {
    ID          uint      `gorm:"primaryKey" json:"id"`
    EntryDate   time.Time `gorm:"not null" json:"entry_date"`
    Description string    `gorm:"size:200;not null" json:"description"`
    Amount      float64   `gorm:"not null;default:0" json:"amount"`
    MovementType string    `gorm:"size:20;not null" json:"movement_type"`
    Balance     float64   `gorm:"not null;default:0" json:"balance"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
