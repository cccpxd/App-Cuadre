package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"appcuadre/internal/services"

	"github.com/gin-gonic/gin"
)

const maxLoginBodyBytes = 4 << 10

// Authenticator es lo que el handler necesita del servicio de autenticación.
type Authenticator interface {
	Login(ctx context.Context, email, password string) (services.LoginResult, error)
	Logout(ctx context.Context, rawToken string) error
}

// CookieConfig describe la cookie de sesión.
type CookieConfig struct {
	Name     string
	TTL      time.Duration
	Secure   bool
	SameSite http.SameSite
}

type AuthHandler struct {
	auth   Authenticator
	logger *slog.Logger
	cookie CookieConfig
}

func NewAuthHandler(auth Authenticator, logger *slog.Logger, cookie CookieConfig) *AuthHandler {
	return &AuthHandler{auth: auth, logger: logger, cookie: cookie}
}

func (h *AuthHandler) LoginPage(c *gin.Context) {
	h.renderLogin(c, http.StatusOK, "", "")
}

func (h *AuthHandler) Login(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxLoginBodyBytes)

	email := strings.TrimSpace(c.PostForm("email"))
	password := c.PostForm("password")
	if email == "" || password == "" {
		h.renderLogin(c, http.StatusBadRequest, "Ingresa tu email y tu contraseña.", email)
		return
	}

	result, err := h.auth.Login(c.Request.Context(), email, password)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			h.renderLogin(c, http.StatusUnauthorized, "Credenciales inválidas.", email)
			return
		}
		h.logger.ErrorContext(c.Request.Context(), "login_failed", "error", err)
		h.renderLogin(c, http.StatusInternalServerError, "No se pudo iniciar sesión. Intenta de nuevo.", email)
		return
	}

	h.setSessionCookie(c, result.Token)
	c.Redirect(http.StatusSeeOther, "/")
}

func (h *AuthHandler) Logout(c *gin.Context) {
	token, err := c.Cookie(h.cookie.Name)
	if err == nil && token != "" {
		if err := h.auth.Logout(c.Request.Context(), token); err != nil {
			h.logger.ErrorContext(c.Request.Context(), "logout_failed", "error", err)
		}
	}

	h.clearSessionCookie(c)
	c.Redirect(http.StatusSeeOther, "/login")
}

func (h *AuthHandler) renderLogin(c *gin.Context, status int, message, email string) {
	c.HTML(status, "login.html", gin.H{
		"title": "Iniciar sesión",
		"error": message,
		"email": email,
	})
}

func (h *AuthHandler) setSessionCookie(c *gin.Context, token string) {
	c.SetSameSite(h.cookie.SameSite)
	c.SetCookie(h.cookie.Name, token, int(h.cookie.TTL.Seconds()), "/", "", h.cookie.Secure, true)
}

func (h *AuthHandler) clearSessionCookie(c *gin.Context) {
	c.SetSameSite(h.cookie.SameSite)
	c.SetCookie(h.cookie.Name, "", -1, "/", "", h.cookie.Secure, true)
}
