package httpx

import "github.com/gin-gonic/gin"

const (
	ctxUserID   = "userID"
	ctxUserRole = "userRole"
)

// SetUserID guarda en el contexto el usuario autenticado de la petición.
func SetUserID(c *gin.Context, userID string) { c.Set(ctxUserID, userID) }

// SetUserRole guarda en el contexto el rol autorizado de la petición.
func SetUserRole(c *gin.Context, role string) { c.Set(ctxUserRole, role) }

// SetActor guarda identidad y rol juntos; lo usan registro y login, que
// autentican dentro del handler.
func SetActor(c *gin.Context, userID, role string) {
	SetUserID(c, userID)
	SetUserRole(c, role)
}

// UserID devuelve el usuario autenticado, o "" si la petición es anónima.
func UserID(c *gin.Context) string { return c.GetString(ctxUserID) }

// UserRole devuelve el rol ya autorizado, o "" si aún no se resolvió.
func UserRole(c *gin.Context) string { return c.GetString(ctxUserRole) }
