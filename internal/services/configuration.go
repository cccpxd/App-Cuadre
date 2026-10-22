package services

import (
    "appcuadre/internal/database"
    "appcuadre/internal/models"
)

func GetSystemConfiguration() (models.Configuration, error) {
    var cfg models.Configuration
    result := database.DB.FirstOrCreate(&cfg, models.Configuration{CompanyName: "Correspondencia"})
    if result.Error != nil {
        return models.Configuration{}, result.Error
    }
    return cfg, nil
}

func SaveSystemConfiguration(cfg models.Configuration) (models.Configuration, error) {
    if cfg.ID == 0 {
        if err := database.DB.Create(&cfg).Error; err != nil {
            return models.Configuration{}, err
        }
        return cfg, nil
    }

    if err := database.DB.Save(&cfg).Error; err != nil {
        return models.Configuration{}, err
    }
    return cfg, nil
}
