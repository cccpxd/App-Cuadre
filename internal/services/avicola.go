package services

import (
    "fmt"
    "math"
    "time"

    "appcuadre/internal/database"
    "appcuadre/internal/models"
)

func CreateAvicolaMovement(movement models.AvicolaLedger) (models.AvicolaLedger, error) {
    if movement.Description == "" {
        return models.AvicolaLedger{}, fmt.Errorf("la descripción es obligatoria")
    }
    if movement.MovementType == "" {
        movement.MovementType = "ingreso"
    }
    if movement.EntryDate.IsZero() {
        movement.EntryDate = time.Now()
    }

    var latestBalance float64
    var latest models.AvicolaLedger
    if err := database.DB.Order("entry_date desc, id desc").First(&latest).Error; err == nil {
        latestBalance = latest.Balance
    }

    if movement.MovementType == "ingreso" {
        movement.Balance = latestBalance + movement.Amount
    } else if movement.MovementType == "egreso" {
        movement.Balance = latestBalance - movement.Amount
    } else {
        movement.Balance = latestBalance
    }

    if math.IsNaN(movement.Balance) || math.IsInf(movement.Balance, 0) {
        return models.AvicolaLedger{}, fmt.Errorf("el saldo calculado no es válido")
    }

    if err := database.DB.Create(&movement).Error; err != nil {
        return models.AvicolaLedger{}, err
    }
    return movement, nil
}

func ListAvicolaMovements() ([]models.AvicolaLedger, error) {
    var movements []models.AvicolaLedger
    result := database.DB.Order("entry_date desc").Find(&movements)
    if result.Error != nil {
        return nil, result.Error
    }
    return movements, nil
}

func GetAvicolaBalance() (float64, error) {
    var movement models.AvicolaLedger
    result := database.DB.Order("entry_date desc").First(&movement)
    if result.Error != nil {
        return 0, result.Error
    }
    return movement.Balance, nil
}
