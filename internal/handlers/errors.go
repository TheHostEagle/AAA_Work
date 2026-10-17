package handlers

import (
	"errors"
	"net/http"

	"automarket/internal/httpx"
	"automarket/internal/models"
	"automarket/internal/service"

	"github.com/gin-gonic/gin"
)

// errorMessages permite a cada handler personalizar los textos por tipo de error.
type errorMessages struct {
	NotFound  string
	Forbidden string
	Conflict  string
	Internal  string
}

// respondServiceError traduce los errores de servicio y dominio a un código
// HTTP y a un cuerpo de error único para toda la API.
func respondServiceError(c *gin.Context, err error, messages errorMessages) {
	switch {
	case errors.Is(err, service.ErrInvalidCredentials):
		httpx.Error(c, http.StatusUnauthorized, "credenciales inválidas")
	case errors.Is(err, service.ErrForbidden):
		httpx.Error(c, http.StatusForbidden, messages.Forbidden)
	case errors.Is(err, models.ErrNotFound):
		httpx.Error(c, http.StatusNotFound, messages.NotFound)
	case errors.Is(err, models.ErrStatusMismatch):
		httpx.Error(c, http.StatusConflict, messages.Conflict)
	case errors.Is(err, models.ErrEmailTaken):
		httpx.Error(c, http.StatusConflict, models.ErrEmailTaken.Error())
	default:
		httpx.Error(c, http.StatusInternalServerError, messages.Internal)
	}
}
