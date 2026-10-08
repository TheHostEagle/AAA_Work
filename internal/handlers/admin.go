package handlers

import (
	"net/http"

	"automarket/internal/storage"

	"github.com/gin-gonic/gin"
)

// ApproveVehicleHandler cambia a publicada una publicación pendiente.
func ApproveVehicleHandler(store *storage.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		vehicle, err := store.GetVehicle(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "vehículo no encontrado"})
			return
		}
		if vehicle.Status != "pendiente" {
			c.JSON(http.StatusConflict, gin.H{"error": "el vehículo no está pendiente de aprobación"})
			return
		}

		vehicle.Status = "publicada"
		if err := store.UpdateVehicle(vehicle); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo aprobar el vehículo"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "vehículo aprobado", "vehicle": vehicle})
	}
}

// DeleteVehicleHandler elimina una publicación por su ID.
func DeleteVehicleHandler(store *storage.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := store.DeleteVehicle(c.Param("id")); err != nil {
			if err == storage.ErrNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "vehículo no encontrado"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo eliminar el vehículo"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "vehículo eliminado"})
	}
}

// DeleteUserHandler elimina una cuenta por su ID.
func DeleteUserHandler(store *storage.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := store.DeleteUser(c.Param("id")); err != nil {
			if err == storage.ErrNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "usuario no encontrado"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo eliminar el usuario"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "usuario eliminado"})
	}
}
