// Package middleware implementa el marco AAA (RF-01) como middlewares de Gin.
package middleware

import (
	"log"
	"net/http"
	"strings"

	"automarket/internal/httpx"
	"automarket/internal/models"
	"automarket/internal/security"

	"github.com/gin-gonic/gin"
)

type userReader interface {
	GetUserByID(id string) (models.User, error)
}

type auditStore interface {
	userReader
	SaveAuditLog(entry *models.AuditLog) error
}

// AuthenticationMiddleware exige "Authorization: Bearer <token>" válido y
// deja el ID del usuario en el contexto.
func AuthenticationMiddleware(tokens *security.TokenStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			httpx.Abort(c, http.StatusUnauthorized, "token requerido")
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			httpx.Abort(c, http.StatusUnauthorized, "formato de token inválido")
			return
		}

		userID, ok := tokens.GetUserID(parts[1])
		if !ok {
			httpx.Abort(c, http.StatusUnauthorized, "token inválido")
			return
		}

		httpx.SetUserID(c, userID)
		c.Next()
	}
}

// AuthorizationMiddleware exige que el usuario autenticado tenga el rol
// requerido. El rol se lee del almacenamiento en cada petición.
func AuthorizationMiddleware(users userReader, requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := httpx.UserID(c)
		if userID == "" {
			httpx.Abort(c, http.StatusUnauthorized, "usuario no autenticado")
			return
		}

		user, err := users.GetUserByID(userID)
		if err != nil {
			httpx.Abort(c, http.StatusUnauthorized, "usuario no encontrado")
			return
		}
		if user.Role != requiredRole {
			httpx.Abort(c, http.StatusForbidden, "no tienes permisos para realizar esta operación")
			return
		}

		httpx.SetUserRole(c, user.Role)
		c.Next()
	}
}

// AccountingMiddleware registra cada petición, con su resultado, en la cadena
// de auditoría. Si el registro falla, solo se informa en el log del servidor.
func AccountingMiddleware(store auditStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		entry := models.AuditLog{
			ActorID:        "anonimo",
			ActorRole:      models.RoleVisitor,
			Method:         c.Request.Method,
			Endpoint:       c.Request.URL.Path,
			StatusResponse: c.Writer.Status(),
		}
		if userID := httpx.UserID(c); userID != "" {
			entry.ActorID = userID
			entry.ActorRole = resolveRole(c, store, userID)
		}

		if err := store.SaveAuditLog(&entry); err != nil {
			log.Printf("no se pudo guardar el registro de auditoría: %v", err)
		}
	}
}

// resolveRole usa el rol ya autorizado y, si la petición falló antes de
// autorizarse, lo consulta; sin dato conserva el rol de visitante.
func resolveRole(c *gin.Context, users userReader, userID string) string {
	if role := httpx.UserRole(c); role != "" {
		return role
	}
	if user, err := users.GetUserByID(userID); err == nil {
		return user.Role
	}
	return models.RoleVisitor
}
