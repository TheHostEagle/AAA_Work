package handlers

import (
	"errors"
	"net/http"

	"automarket/internal/httpx"
	"automarket/internal/service"
	"automarket/internal/storage"

	"github.com/gin-gonic/gin"
)

type errorMessages struct {
	NotFound  string
	Forbidden string
	Conflict  string
	Internal  string
}

func respondServiceError(c *gin.Context, err error, messages errorMessages) {
	switch {
	case errors.Is(err, service.ErrInvalidCredentials):
		httpx.Error(c, http.StatusUnauthorized, "credenciales inválidas")
	case errors.Is(err, service.ErrForbidden):
		httpx.Error(c, http.StatusForbidden, messages.Forbidden)
	case errors.Is(err, storage.ErrNotFound):
		httpx.Error(c, http.StatusNotFound, messages.NotFound)
	case errors.Is(err, storage.ErrStatusMismatch):
		httpx.Error(c, http.StatusConflict, messages.Conflict)
	case errors.Is(err, storage.ErrEmailTaken):
		httpx.Error(c, http.StatusConflict, storage.ErrEmailTaken.Error())
	default:
		httpx.Error(c, http.StatusInternalServerError, messages.Internal)
	}
}
