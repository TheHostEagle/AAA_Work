package middleware

import (
	"log"
	"net/http"
	"strings"

	"automarket/internal/httpx"
	"automarket/internal/models"
	"automarket/internal/security"
	"automarket/internal/storage"

	"github.com/gin-gonic/gin"
)

// AuthenticationMiddleware valida que la petición tenga
// un token válido.
func AuthenticationMiddleware(tokenStore *security.TokenStore) gin.HandlerFunc {
	return func(c *gin.Context) {

		// 1. Obtener el header Authorization.
		authHeader := c.GetHeader("Authorization")

		// 2. Si no existe, rechazamos la petición.
		if authHeader == "" {
			httpx.Abort(c, http.StatusUnauthorized, "token requerido")
			return
		}

		// 3. El formato esperado es:
		// Authorization: Bearer <token>
		parts := strings.SplitN(authHeader, " ", 2)

		if len(parts) != 2 || parts[0] != "Bearer" {
			httpx.Abort(c, http.StatusUnauthorized, "formato de token inválido")
			return
		}

		token := parts[1]

		// 4. Buscar el usuario asociado al token.
		userID, ok := tokenStore.GetUserID(token)

		if !ok {
			httpx.Abort(c, http.StatusUnauthorized, "token inválido")
			return
		}

		// 5. Guardamos el ID del usuario en el contexto.
		c.Set("userID", userID)

		// 6. Continuamos hacia el siguiente middleware/handler.
		c.Next()
	}
}

// AuthorizationMiddleware valida que el usuario tenga
// el rol requerido.
func AuthorizationMiddleware(
	store *storage.Store,
	requiredRole string,
) gin.HandlerFunc {
	return func(c *gin.Context) {

		// 1. Obtener el userID que dejó AuthenticationMiddleware.
		userIDValue, exists := c.Get("userID")

		if !exists {
			httpx.Abort(c, http.StatusUnauthorized, "usuario no autenticado")
			return
		}

		// 2. Convertir el valor obtenido a string.
		userID, ok := userIDValue.(string)

		if !ok {
			httpx.Abort(c, http.StatusUnauthorized, "identidad inválida")
			return
		}

		// 3. Buscar al usuario en el almacenamiento.
		user, err := store.GetUserByID(userID)

		if err != nil {
			httpx.Abort(c, http.StatusUnauthorized, "usuario no encontrado")
			return
		}

		// 4. Comprobar si el usuario tiene el rol requerido.
		if user.Role != requiredRole {
			httpx.Abort(c, http.StatusForbidden, "no tienes permisos para realizar esta operación")
			return
		}

		// 5. El usuario tiene permiso.
		c.Set("userRole", user.Role)
		c.Next()
	}
}

// AccountingMiddleware registra cada petición después de conocer su resultado.
func AccountingMiddleware(store *storage.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		entry := models.AuditLog{
			ActorID:        "anonimo",
			ActorRole:      models.RoleVisitor,
			Method:         c.Request.Method,
			Endpoint:       c.Request.URL.Path,
			StatusResponse: c.Writer.Status(),
		}

		if userIDValue, exists := c.Get("userID"); exists {
			if userID, ok := userIDValue.(string); ok {
				entry.ActorID = userID
				entry.ActorRole = models.RoleVisitor
				if roleValue, hasRole := c.Get("userRole"); hasRole {
					if role, ok := roleValue.(string); ok {
						entry.ActorRole = role
					}
				} else if user, err := store.GetUserByID(userID); err == nil {
					entry.ActorRole = user.Role
				}
			}
		}

		if err := store.SaveAuditLog(&entry); err != nil {
			log.Printf("no se pudo guardar el registro de auditoría: %v", err)
		}
	}
}
