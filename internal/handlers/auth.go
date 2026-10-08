package handlers

import (
	"net/http"

	"automarket/internal/models"
	"automarket/internal/security"
	"automarket/internal/storage"

	"github.com/gin-gonic/gin"
)

func RegisterHandler(store *storage.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "JSON inválido",
			})
			return
		}

		if input.Email == "" || input.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "email y password son obligatorios",
			})
			return
		}

		passwordHash, err := security.HashPassword(input.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "no se pudo procesar la contraseña",
			})
			return
		}

		user := models.User{
			Name:         input.Name,
			Email:        input.Email,
			PasswordHash: passwordHash,
			Role:         "vendedor",
		}

		if err := store.CreateUser(&user); err != nil {
			c.JSON(http.StatusConflict, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.Set("userID", user.ID)
		c.Set("userRole", user.Role)

		c.JSON(http.StatusCreated, gin.H{
			"message": "usuario creado",
			"user": gin.H{
				"id":    user.ID,
				"name":  user.Name,
				"email": user.Email,
				"role":  user.Role,
			},
		})
	}
}

func LoginHandler(store *storage.Store, tokenStore *security.TokenStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "JSON inválido",
			})
			return
		}

		user, err := store.GetUserByEmail(input.Email)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "credenciales inválidas",
			})
			return
		}

		valid, err := security.VerifyPassword(
			input.Password,
			user.PasswordHash,
		)

		if err != nil || !valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "credenciales inválidas",
			})
			return
		}

		token, err := tokenStore.Create(user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "no se pudo generar el token",
			})
			return
		}

		c.Set("userID", user.ID)
		c.Set("userRole", user.Role)

		c.JSON(http.StatusOK, gin.H{
			"message": "login correcto",
			"token":   token,
			"user": gin.H{
				"id":    user.ID,
				"name":  user.Name,
				"email": user.Email,
				"role":  user.Role,
			},
		})
	}
}
