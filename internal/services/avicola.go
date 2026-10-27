package services

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"appcuadre/internal/models"

	"gorm.io/gorm"
)

const (
	avicolaIncome              = "ingreso"
	avicolaExpense             = "egreso"
	maxAvicolaDescriptionRunes = 200
	avicolaDateLayout          = "2006-01-02"
)

// AvicolaService maneja los movimientos y el saldo del módulo avícola.
type AvicolaService struct {
	db *gorm.DB
}

func NewAvicolaService(db *gorm.DB) (*AvicolaService, error) {
	if db == nil {
		return nil, errors.New("avicola service: falta la base de datos")
	}
	return &AvicolaService{db: db}, nil
}

// latestBalance devuelve el saldo del último movimiento que cumpla la
// condición. found es false cuando no hay ninguno. Un error de base de datos
// se devuelve como error: antes se confundía con "no hay movimientos" y el
// saldo salía en 0 sin avisar.
func (s *AvicolaService) latestBalance(condition string, args ...interface{}) (balance float64, found bool, err error) {
	var movement models.AvicolaLedger
	err = s.db.Where(condition, args...).Order("entry_date desc, id desc").First(&movement).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("buscar saldo avícola: %w", err)
	}
	return movement.Balance, true, nil
}

func (s *AvicolaService) getAvicolaBalanceForDate(userID uint, date time.Time) (float64, error) {
	dateKey := date.Format(avicolaDateLayout)

	balance, found, err := s.latestBalance("user_id = ? AND DATE(entry_date) = ?", userID, dateKey)
	if err != nil || found {
		return balance, err
	}

	balance, _, err = s.latestBalance("user_id = ? AND DATE(entry_date) < ?", userID, dateKey)
	return balance, err
}

// avicolaDay deja solo el día (en UTC) de una fecha, para comparar días.
func avicolaDay(value time.Time) time.Time {
	year, month, day := value.UTC().Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func (s *AvicolaService) CreateAvicolaMovement(userID uint, movement models.AvicolaLedger) (models.AvicolaLedger, error) {
	if userID == 0 {
		return models.AvicolaLedger{}, fmt.Errorf("falta el usuario del movimiento")
	}

	movement.Description = strings.TrimSpace(movement.Description)
	if movement.Description == "" {
		return models.AvicolaLedger{}, ValidationError("la descripción es obligatoria")
	}
	if len([]rune(movement.Description)) > maxAvicolaDescriptionRunes {
		return models.AvicolaLedger{}, ValidationError(fmt.Sprintf("la descripción no puede pasar de %d caracteres", maxAvicolaDescriptionRunes))
	}
	if movement.Amount <= 0 || math.IsNaN(movement.Amount) || math.IsInf(movement.Amount, 0) {
		return models.AvicolaLedger{}, ValidationError("el valor debe ser mayor a cero")
	}

	switch movement.MovementType {
	case "":
		movement.MovementType = avicolaIncome
	case avicolaIncome, avicolaExpense:
		// Tipo válido.
	default:
		return models.AvicolaLedger{}, ValidationError("el tipo de movimiento debe ser ingreso o egreso")
	}

	if movement.EntryDate.IsZero() {
		movement.EntryDate = time.Now()
	}

	// Nunca se confía en el id ni en el dueño que mande el cliente.
	movement.ID = 0
	movement.UserID = userID

	// Todo el cálculo del saldo y el guardado van en una transacción con un
	// candado por usuario: dos peticiones al tiempo no pueden leer el mismo
	// saldo anterior y dejar un saldo equivocado.
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", int64(userID)).Error; err != nil {
			return fmt.Errorf("bloquear movimientos del usuario: %w", err)
		}

		var last models.AvicolaLedger
		currentBalance := 0.0
		err := tx.Where("user_id = ?", userID).Order("entry_date desc, id desc").First(&last).Error
		switch {
		case err == nil:
			// Un movimiento con fecha anterior dejaría desactualizados los
			// saldos de los movimientos que ya existen después de esa fecha.
			if avicolaDay(movement.EntryDate).Before(avicolaDay(last.EntryDate)) {
				return ValidationError("no se puede registrar un movimiento con fecha anterior al último movimiento")
			}
			// Mismo día pero una hora anterior (por ejemplo, relojes
			// desajustados): se alinea con el último para mantener el orden.
			if movement.EntryDate.Before(last.EntryDate) {
				movement.EntryDate = last.EntryDate
			}
			currentBalance = last.Balance
		case !errors.Is(err, gorm.ErrRecordNotFound):
			return fmt.Errorf("buscar el último movimiento: %w", err)
		}

		if movement.MovementType == avicolaIncome {
			movement.Balance = currentBalance + movement.Amount
		} else {
			movement.Balance = currentBalance - movement.Amount
		}

		if math.IsNaN(movement.Balance) || math.IsInf(movement.Balance, 0) {
			return fmt.Errorf("el saldo calculado no es válido")
		}

		return tx.Create(&movement).Error
	})
	if err != nil {
		return models.AvicolaLedger{}, err
	}
	return movement, nil
}

func (s *AvicolaService) ListAvicolaMovements(userID uint, dateFilter string) ([]models.AvicolaLedger, error) {
	var movements []models.AvicolaLedger
	query := s.db.
		Where("user_id = ?", userID).
		Order("entry_date desc, id desc")

	filteredDate := strings.TrimSpace(dateFilter)
	if filteredDate != "" {
		if _, err := time.Parse(avicolaDateLayout, filteredDate); err != nil {
			return nil, ValidationError("fecha inválida")
		}
		query = query.Where("DATE(entry_date) = ?", filteredDate)
	}

	result := query.Find(&movements)
	if result.Error != nil {
		return nil, result.Error
	}
	return movements, nil
}

func (s *AvicolaService) GetAvicolaBalance(userID uint, dateFilter string) (float64, error) {
	filteredDate := strings.TrimSpace(dateFilter)
	if filteredDate == "" {
		filteredDate = time.Now().Format(avicolaDateLayout)
	}

	parsedDate, err := time.Parse(avicolaDateLayout, filteredDate)
	if err != nil {
		return 0, ValidationError("fecha inválida")
	}

	return s.getAvicolaBalanceForDate(userID, parsedDate)
}
