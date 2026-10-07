package middleware

import (
	"net/http"
	"strings"

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
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "token requerido",
			})
			c.Abort()
			return
		}

		// 3. El formato esperado es:
		// Authorization: Bearer <token>
		parts := strings.SplitN(authHeader, " ", 2)

		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "formato de token inválido",
			})
			c.Abort()
			return
		}

		token := parts[1]

		// 4. Buscar el usuario asociado al token.
		userID, ok := tokenStore.GetUserID(token)

		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "token inválido",
			})
			c.Abort()
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
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "usuario no autenticado",
			})
			c.Abort()
			return
		}

		// 2. Convertir el valor obtenido a string.
		userID, ok := userIDValue.(string)

		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "identidad inválida",
			})
			c.Abort()
			return
		}

		// 3. Buscar al usuario en el almacenamiento.
		user, err := store.GetUserByID(userID)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "usuario no encontrado",
			})
			c.Abort()
			return
		}

		// 4. Comprobar si el usuario tiene el rol requerido.
		if user.Role != requiredRole {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "no tienes permisos para realizar esta operación",
			})
			c.Abort()
			return
		}

		// 5. El usuario tiene permiso.
		c.Next()
	}
}

// AccountingMiddleware registrará de forma trazable
// las operaciones.
func AccountingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// La auditoría la implementaremos después.
	}
}
