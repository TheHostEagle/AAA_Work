package handlers

import (
	"net/http"
	"strings"

	"automarket/internal/httpx"
	"automarket/internal/service"

	"github.com/gin-gonic/gin"
)

// RegisterHandler valida los datos de registro y delega la creación de la cuenta.
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
		input.Name = strings.TrimSpace(input.Name)
		input.Email = strings.TrimSpace(input.Email)
		if input.Name == "" || input.Email == "" || input.Password == "" {
			httpx.Error(c, http.StatusBadRequest, "todos los campos son obligatorios")
			return
		}

		user, err := auth.Register(input.Name, input.Email, input.Password)
		if err != nil {
			respondServiceError(c, err, errorMessages{Internal: "no se pudo crear el usuario"})
			return
		}
		httpx.SetActor(c, user.ID, user.Role)
		c.JSON(http.StatusCreated, gin.H{
			"message": "usuario creado",
			"user": gin.H{
				"id": user.ID, "name": user.Name, "email": user.Email, "role": user.Role,
			},
		})
	}
}

// LoginHandler autentica las credenciales y devuelve un token Bearer opaco.
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

		user, token, err := auth.Login(strings.TrimSpace(input.Email), input.Password)
		if err != nil {
			respondServiceError(c, err, errorMessages{Internal: "no se pudo generar el token"})
			return
		}
		httpx.SetActor(c, user.ID, user.Role)
		c.JSON(http.StatusOK, gin.H{
			"message": "login correcto",
			"token":   token,
			"user": gin.H{
				"id": user.ID, "name": user.Name, "email": user.Email, "role": user.Role,
			},
		})
	}
}
