package handlers

import (
	"net/http"

	"automarket/internal/service"

	"github.com/gin-gonic/gin"
)

func ApproveVehicleHandler(admin *service.AdminService) gin.HandlerFunc {
	return func(c *gin.Context) {
		vehicle, err := admin.ApproveVehicle(c.Param("id"))
		if err != nil {
			respondServiceError(c, err, errorMessages{
				NotFound: "vehículo no encontrado",
				Conflict: "el vehículo no está pendiente de aprobación",
				Internal: "no se pudo aprobar el vehículo",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "vehículo aprobado", "vehicle": vehicle})
	}
}

func DeleteVehicleHandler(admin *service.AdminService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := admin.DeleteVehicle(c.Param("id")); err != nil {
			respondServiceError(c, err, errorMessages{
				NotFound: "vehículo no encontrado",
				Internal: "no se pudo eliminar el vehículo",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "vehículo eliminado"})
	}
}

func DeleteUserHandler(admin *service.AdminService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := admin.DeleteUser(c.Param("id")); err != nil {
			respondServiceError(c, err, errorMessages{
				NotFound: "usuario no encontrado",
				Internal: "no se pudo eliminar el usuario",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "usuario eliminado"})
	}
}
