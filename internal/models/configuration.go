package models

import "time"

// Configuration guarda la configuración operativa del sistema de cuadre.
type Configuration struct {
    ID                     uint      `gorm:"primaryKey" json:"id"`
    CompanyName            string    `gorm:"size:200;not null;default:'Correspondencia'" json:"company_name"`
    DefaultCurrency        string    `gorm:"size:10;not null;default:'COP'" json:"default_currency"`
    ShowAvicolaDashboard   bool      `gorm:"default:true" json:"show_avicola_dashboard"`
    AllowCustomFields      bool      `gorm:"default:true" json:"allow_custom_fields"`
    CreatedAt              time.Time `json:"created_at"`
    UpdatedAt              time.Time `json:"updated_at"`
}
