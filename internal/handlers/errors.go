package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"appcuadre/internal/services"

	"github.com/gin-gonic/gin"
)

const invalidBodyMessage = "los datos enviados no son válidos"

// respondServiceError decide qué ve el usuario cuando un servicio falla.
// Si el error es de datos inválidos, se muestra el mensaje (400). Si es
// cualquier otro (base de datos, red...), se guarda en el log y el usuario
// solo recibe un mensaje genérico (500), para no filtrar detalles internos.
func respondServiceError(c *gin.Context, logger *slog.Logger, err error) {
	var validation services.ValidationError
	if errors.As(err, &validation) {
		c.JSON(http.StatusBadRequest, gin.H{"error": validation.Error()})
		return
	}

	logger.ErrorContext(c.Request.Context(), "request_failed",
		"method", c.Request.Method,
		"path", c.FullPath(),
		"error", err,
	)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno del servidor"})
}
