package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// SecurityHeaders aplica cabeceras de seguridad conservadoras. Las páginas y
// la API nunca se guardan en caché, porque muestran datos de caja.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.Writer.Header()
		header.Set("X-Content-Type-Options", "nosniff")
		header.Set("X-Frame-Options", "DENY")
		header.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		header.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		header.Set("Content-Security-Policy", "default-src 'self'; base-uri 'self'; form-action 'self'; frame-ancestors 'none'; img-src 'self' data:; style-src 'self'; script-src 'self'; object-src 'none'")
		if strings.HasPrefix(c.Request.URL.Path, "/static/") {
			header.Set("Cache-Control", "no-cache")
		} else {
			header.Set("Cache-Control", "no-store")
		}
		c.Next()
	}
}
