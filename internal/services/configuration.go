package services

import (
	"errors"
	"fmt"

	"appcuadre/internal/models"

	"gorm.io/gorm"
)

// ConfigurationService guarda y lee la configuración de cada usuario.
type ConfigurationService struct {
	db *gorm.DB
}

func NewConfigurationService(db *gorm.DB) (*ConfigurationService, error) {
	if db == nil {
		return nil, errors.New("configuration service: falta la base de datos")
	}
	return &ConfigurationService{db: db}, nil
}

func (s *ConfigurationService) GetSystemConfiguration(userID uint) (models.Configuration, error) {
	if userID == 0 {
		return models.Configuration{}, fmt.Errorf("falta el usuario de la configuración")
	}

	var cfg models.Configuration
	result := s.db.
		Where("user_id = ?", userID).
		Order("id").
		Attrs(models.Configuration{
			UserID:               userID,
			CompanyName:          "Correspondencia",
			DefaultCurrency:      "COP",
			ShowAvicolaDashboard: true,
			AllowCustomFields:    true,
		}).
		FirstOrCreate(&cfg)
	if result.Error != nil {
		return models.Configuration{}, result.Error
	}
	return cfg, nil
}

func (s *ConfigurationService) SaveSystemConfiguration(userID uint, cfg models.Configuration) (models.Configuration, error) {
	current, err := s.GetSystemConfiguration(userID)
	if err != nil {
		return models.Configuration{}, err
	}

	updates := map[string]interface{}{
		"company_name":           cfg.CompanyName,
		"default_currency":       cfg.DefaultCurrency,
		"show_avicola_dashboard": cfg.ShowAvicolaDashboard,
		"allow_custom_fields":    cfg.AllowCustomFields,
	}

	if err := s.db.
		Model(&models.Configuration{}).
		Where("id = ? AND user_id = ?", current.ID, userID).
		Updates(updates).Error; err != nil {
		return models.Configuration{}, err
	}

	return s.GetSystemConfiguration(userID)
}
