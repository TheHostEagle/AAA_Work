package handlers

import (
	"net/http"

	"automarket/internal/httpx"
	"automarket/internal/service"

	"github.com/gin-gonic/gin"
)

// RegisterHandler validates registration input and delegates account creation.
func RegisterHandler(auth *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			httpx.Error(c, http.StatusBadRequest, "JSON inválido")
			return
		}
		if input.Name == "" || input.Email == "" || input.Password == "" {
			httpx.Error(c, http.StatusBadRequest, "todos los campos son obligatorios")
			return
		}

		user, err := auth.Register(input.Name, input.Email, input.Password)
		if err != nil {
			respondServiceError(c, err, errorMessages{Internal: "no se pudo crear el usuario"})
			return
		}
		c.Set("userID", user.ID)
		c.Set("userRole", user.Role)
		c.JSON(http.StatusCreated, gin.H{
			"message": "usuario creado",
			"user": gin.H{
				"id": user.ID, "name": user.Name, "email": user.Email, "role": user.Role,
			},
		})
	}
}

// LoginHandler authenticates credentials and returns an opaque bearer token.
func LoginHandler(auth *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			httpx.Error(c, http.StatusBadRequest, "JSON inválido")
			return
		}

		user, token, err := auth.Login(input.Email, input.Password)
		if err != nil {
			respondServiceError(c, err, errorMessages{Internal: "no se pudo generar el token"})
			return
		}
		c.Set("userID", user.ID)
		c.Set("userRole", user.Role)
		c.JSON(http.StatusOK, gin.H{
			"message": "login correcto",
			"token":   token,
			"user": gin.H{
				"id": user.ID, "name": user.Name, "email": user.Email, "role": user.Role,
			},
		})
	}
}
