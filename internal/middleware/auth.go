package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"appcuadre/internal/models"
	"appcuadre/internal/services"

	"github.com/gin-gonic/gin"
)

type contextKey string

const currentUserKey contextKey = "currentUser"

// SessionValidator convierte un token de sesión en el usuario dueño.
type SessionValidator interface {
	Authenticate(ctx context.Context, rawToken string) (models.AdminUser, error)
}

// RequireAuth deja pasar solo las peticiones con una sesión válida.
// Las páginas redirigen a /login y la API responde 401 en JSON.
func RequireAuth(auth SessionValidator, cookieName string, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(cookieName)
		if err != nil || token == "" {
			deny(c)
			return
		}

		user, err := auth.Authenticate(c.Request.Context(), token)
		switch {
		case errors.Is(err, services.ErrInvalidSession):
			deny(c)
		case err != nil:
			logger.ErrorContext(c.Request.Context(), "session_lookup_failed", "error", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
		default:
			c.Set(string(currentUserKey), user)
			c.Next()
		}
	}
}

// CurrentUser devuelve el usuario autenticado que RequireAuth guardó en el
// contexto. Los handlers nunca deben leer un user_id enviado por el cliente.
func CurrentUser(c *gin.Context) (models.AdminUser, bool) {
	value, exists := c.Get(string(currentUserKey))
	if !exists {
		return models.AdminUser{}, false
	}
	user, ok := value.(models.AdminUser)
	return user, ok
}

func deny(c *gin.Context) {
	if strings.HasPrefix(c.Request.URL.Path, "/api/") {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "no autorizado"})
		return
	}
	c.Redirect(http.StatusFound, "/login")
	c.Abort()
}
