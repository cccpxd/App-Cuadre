package handlers

import (
	"net/http"

	"appcuadre/internal/middleware"
	"appcuadre/internal/models"

	"github.com/gin-gonic/gin"
)

func currentUser(c *gin.Context) (models.AdminUser, bool) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "no autorizado"})
		return models.AdminUser{}, false
	}
	return user, true
}

func pageData(c *gin.Context, title string, extra gin.H) gin.H {
	data := gin.H{
		"title":            title,
		"currentUserEmail": "",
	}
	if user, ok := middleware.CurrentUser(c); ok {
		data["currentUserEmail"] = user.Email
	}
	for key, value := range extra {
		data[key] = value
	}
	return data
}
