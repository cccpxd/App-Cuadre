package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"appcuadre/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// AuthError es un error de autenticación definido como constante, para no
// usar variables globales. errors.Is funciona igual con estos valores.
type AuthError string

func (e AuthError) Error() string { return string(e) }

const (
	ErrInvalidCredentials AuthError = "credenciales inválidas"
	ErrInvalidSession     AuthError = "sesión inválida"
	ErrAdminEmailInvalid  AuthError = "ADMIN_EMAIL no es un email válido"
	ErrAdminPasswordShort AuthError = "ADMIN_PASSWORD debe tener al menos 10 caracteres"
	ErrAdminPasswordLong  AuthError = "ADMIN_PASSWORD no puede pasar de 72 bytes"
)

const (
	bcryptCost          = 12
	minAdminPasswordLen = 10
	maxPasswordBytes    = 72 // límite real de bcrypt
	sessionTokenBytes   = 32

	// dummyPasswordHash se compara cuando el email no existe, para que
	// responder "credenciales inválidas" tarde lo mismo exista o no el usuario.
	// Debe tener el mismo costo que bcryptCost.
	dummyPasswordHash = "$2a$12$JRj4Gz.2RRt7xRGAzzkBzODz87dF193kAVhtqOXkM8KAsY/EYEl6W"
)

// LoginResult trae lo único que la capa HTTP necesita después de un login.
type LoginResult struct {
	Token     string
	ExpiresAt time.Time
}

// AuthService contiene las reglas de autenticación. No sabe nada de HTTP,
// cookies ni plantillas.
type AuthService struct {
	db         *gorm.DB
	sessionTTL time.Duration
	now        func() time.Time
}

func NewAuthService(db *gorm.DB, sessionTTL time.Duration) (*AuthService, error) {
	if db == nil {
		return nil, errors.New("auth service: falta la base de datos")
	}
	if sessionTTL <= 0 {
		return nil, errors.New("auth service: la duración de la sesión debe ser positiva")
	}
	return &AuthService{db: db, sessionTTL: sessionTTL, now: time.Now}, nil
}

// EnsureAdmin crea el administrador si todavía no existe ninguno. Es
// idempotente: si ya hay un administrador no cambia nada (ni el email ni la
// contraseña). Devuelve true solo cuando lo creó.
func (s *AuthService) EnsureAdmin(ctx context.Context, email, password string) (bool, error) {
	var count int64
	if err := s.db.WithContext(ctx).Model(&models.AdminUser{}).Count(&count).Error; err != nil {
		return false, fmt.Errorf("contar administradores: %w", err)
	}
	if count > 0 {
		return false, nil
	}

	normalizedEmail := normalizeEmail(email)
	if err := validateAdminCredentials(normalizedEmail, password); err != nil {
		return false, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return false, fmt.Errorf("crear hash de la contraseña: %w", err)
	}

	admin := models.AdminUser{
		Email:        normalizedEmail,
		PasswordHash: string(hash),
		CreatedAt:    s.now().UTC(),
	}

	// Si otra instancia lo creó al mismo tiempo, no es un error.
	result := s.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&admin)
	if result.Error != nil {
		return false, fmt.Errorf("crear administrador: %w", result.Error)
	}
	return result.RowsAffected == 1, nil
}

// Login verifica las credenciales y crea una sesión en el servidor.
func (s *AuthService) Login(ctx context.Context, email, password string) (LoginResult, error) {
	normalizedEmail := normalizeEmail(email)
	if normalizedEmail == "" || password == "" || len(password) > maxPasswordBytes {
		s.verifyDummyPassword(password)
		return LoginResult{}, ErrInvalidCredentials
	}

	var admin models.AdminUser
	err := s.db.WithContext(ctx).Where("email = ?", normalizedEmail).First(&admin).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		s.verifyDummyPassword(password)
		return LoginResult{}, ErrInvalidCredentials
	}
	if err != nil {
		return LoginResult{}, fmt.Errorf("buscar administrador: %w", err)
	}

	if bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)) != nil {
		return LoginResult{}, ErrInvalidCredentials
	}

	return s.startSession(ctx, admin)
}

// Authenticate convierte un token en el administrador dueño de la sesión.
func (s *AuthService) Authenticate(ctx context.Context, rawToken string) (models.AdminUser, error) {
	if rawToken == "" {
		return models.AdminUser{}, ErrInvalidSession
	}

	var session models.Session
	err := s.db.WithContext(ctx).
		Preload("AdminUser").
		Where("token_hash = ? AND expires_at > ?", fingerprint(rawToken), s.now().UTC()).
		First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.AdminUser{}, ErrInvalidSession
	}
	if err != nil {
		return models.AdminUser{}, fmt.Errorf("buscar sesión: %w", err)
	}
	return session.AdminUser, nil
}

// Logout invalida la sesión asociada al token.
func (s *AuthService) Logout(ctx context.Context, rawToken string) error {
	if rawToken == "" {
		return nil
	}
	err := s.db.WithContext(ctx).Where("token_hash = ?", fingerprint(rawToken)).Delete(&models.Session{}).Error
	if err != nil {
		return fmt.Errorf("borrar sesión: %w", err)
	}
	return nil
}

func (s *AuthService) startSession(ctx context.Context, admin models.AdminUser) (LoginResult, error) {
	rawToken, err := newSessionToken()
	if err != nil {
		return LoginResult{}, fmt.Errorf("crear token de sesión: %w", err)
	}

	now := s.now().UTC()
	session := models.Session{
		TokenHash:   fingerprint(rawToken),
		AdminUserID: admin.ID,
		CreatedAt:   now,
		ExpiresAt:   now.Add(s.sessionTTL),
	}

	// Las sesiones vencidas ya no sirven. Limpiarlas es opcional y no debe
	// impedir un login correcto.
	_ = s.db.WithContext(ctx).Where("expires_at <= ?", now).Delete(&models.Session{}).Error

	if err := s.db.WithContext(ctx).Create(&session).Error; err != nil {
		return LoginResult{}, fmt.Errorf("guardar sesión: %w", err)
	}
	return LoginResult{Token: rawToken, ExpiresAt: session.ExpiresAt}, nil
}

func (s *AuthService) verifyDummyPassword(password string) {
	_ = bcrypt.CompareHashAndPassword([]byte(dummyPasswordHash), []byte(password))
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func validateAdminCredentials(email, password string) error {
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		return ErrAdminEmailInvalid
	}
	if len(password) < minAdminPasswordLen {
		return ErrAdminPasswordShort
	}
	if len(password) > maxPasswordBytes {
		return ErrAdminPasswordLong
	}
	return nil
}

func newSessionToken() (string, error) {
	buffer := make([]byte, sessionTokenBytes)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func fingerprint(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}
