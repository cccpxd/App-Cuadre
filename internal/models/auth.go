package models

import "time"

// AdminUser es la cuenta del administrador. La app es privada: no hay registro
// público, el administrador se crea desde variables de entorno.
type AdminUser struct {
	ID           uint      `gorm:"primaryKey"`
	Email        string    `gorm:"size:255;uniqueIndex;not null"`
	PasswordHash string    `gorm:"size:100;not null"`
	CreatedAt    time.Time `gorm:"not null"`
}

// Session es una sesión iniciada. Solo se guarda el hash del token, nunca el
// token real, así que una fuga de la base de datos no permite usar sesiones.
type Session struct {
	ID          uint      `gorm:"primaryKey"`
	TokenHash   string    `gorm:"size:64;uniqueIndex;not null"`
	AdminUserID uint      `gorm:"index;not null"`
	AdminUser   AdminUser `gorm:"constraint:OnDelete:CASCADE"`
	CreatedAt   time.Time `gorm:"not null"`
	ExpiresAt   time.Time `gorm:"index;not null"`
}
