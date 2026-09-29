package handlers

import "github.com/gin-gonic/gin"

func GetPublicVehiclesHandler(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok", "message": "Catalog public endpoint stub"})
}

func CreateVehicleHandler(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok", "message": "Create vehicle endpoint stub"})
}

func MarkVehicleAsSoldHandler(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok", "message": "Mark vehicle as sold endpoint stub"})
}