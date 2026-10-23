package services

import (
	"fmt"
	"math"
	"strings"
	"time"

	"appcuadre/internal/database"
	"appcuadre/internal/models"
)

func getAvicolaBalanceForDate(date time.Time) (float64, error) {
	dateKey := date.Format("2006-01-02")

	var current models.AvicolaLedger
	if err := database.DB.Where("DATE(entry_date) = ?", dateKey).Order("entry_date desc, id desc").First(&current).Error; err == nil {
		return current.Balance, nil
	}

	var previous models.AvicolaLedger
	if err := database.DB.Where("DATE(entry_date) < ?", dateKey).Order("entry_date desc, id desc").First(&previous).Error; err == nil {
		return previous.Balance, nil
	}

	return 0, nil
}

func CreateAvicolaMovement(movement models.AvicolaLedger) (models.AvicolaLedger, error) {
	if movement.Description == "" {
		return models.AvicolaLedger{}, fmt.Errorf("la descripción es obligatoria")
	}
	if movement.Amount <= 0 {
		return models.AvicolaLedger{}, fmt.Errorf("el valor debe ser mayor a cero")
	}
	if movement.MovementType == "" {
		movement.MovementType = "ingreso"
	}
	if movement.EntryDate.IsZero() {
		movement.EntryDate = time.Now()
	}

	currentBalance, err := getAvicolaBalanceForDate(movement.EntryDate)
	if err != nil {
		return models.AvicolaLedger{}, err
	}

	if movement.MovementType == "ingreso" {
		movement.Balance = currentBalance + movement.Amount
	} else if movement.MovementType == "egreso" {
		movement.Balance = currentBalance - movement.Amount
	} else {
		movement.Balance = currentBalance
	}

	if math.IsNaN(movement.Balance) || math.IsInf(movement.Balance, 0) {
		return models.AvicolaLedger{}, fmt.Errorf("el saldo calculado no es válido")
	}

	if err := database.DB.Create(&movement).Error; err != nil {
		return models.AvicolaLedger{}, err
	}
	return movement, nil
}

func ListAvicolaMovements(dateFilter string) ([]models.AvicolaLedger, error) {
	var movements []models.AvicolaLedger
	query := database.DB.Order("entry_date desc, id desc")

	filteredDate := strings.TrimSpace(dateFilter)
	if filteredDate != "" {
		query = query.Where("DATE(entry_date) = ?", filteredDate)
	}

	result := query.Find(&movements)
	if result.Error != nil {
		return nil, result.Error
	}
	return movements, nil
}

func GetAvicolaBalance(dateFilter string) (float64, error) {
	filteredDate := strings.TrimSpace(dateFilter)
	if filteredDate == "" {
		filteredDate = time.Now().Format("2006-01-02")
	}

	parsedDate, err := time.Parse("2006-01-02", filteredDate)
	if err != nil {
		return 0, fmt.Errorf("fecha inválida: %s", filteredDate)
	}

	return getAvicolaBalanceForDate(parsedDate)
}
