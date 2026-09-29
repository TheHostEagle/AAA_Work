package handlers

import "github.com/gin-gonic/gin"

func RegisterHandler(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok", "message": "Register endpoint stub"})
}

func LoginHandler(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok", "message": "Login endpoint stub"})
}