package handlers

import (
	"net/http"

	"automarket/internal/httpx"
	"automarket/internal/models"
	"automarket/internal/service"

	"github.com/gin-gonic/gin"
)

func GetPublicVehiclesHandler(vehicles *service.VehicleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		catalog, err := vehicles.PublicCatalog()
		if err != nil {
			respondServiceError(c, err, errorMessages{Internal: "no se pudo obtener el contacto del vendedor"})
			return
		}
		c.JSON(http.StatusOK, catalog)
	}
}

func CreateVehicleHandler(vehicles *service.VehicleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Title       string  `json:"title"`
			Brand       string  `json:"brand"`
			Model       string  `json:"model"`
			Price       float64 `json:"price"`
			Description string  `json:"description"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			httpx.Error(c, http.StatusBadRequest, "el JSON no es válido")
			return
		}
		if input.Title == "" || input.Brand == "" || input.Model == "" || input.Price <= 0 {
			httpx.Error(c, http.StatusBadRequest, "título, marca, modelo y precio (mayor a cero) son obligatorios")
			return
		}

		vehicle, err := vehicles.CreateForSeller(c.GetString("userID"), models.Vehicle{
			Title: input.Title, Brand: input.Brand, Model: input.Model,
			Price: input.Price, Description: input.Description,
		})
		if err != nil {
			respondServiceError(c, err, errorMessages{Internal: "no se pudo crear la publicación"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{
			"message": "publicación creada, pendiente de aprobación",
			"vehicle": vehicle,
		})
	}
}

func MarkVehicleAsSoldHandler(vehicles *service.VehicleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		vehicle, err := vehicles.ReportSold(c.GetString("userID"), c.Param("id"))
		if err != nil {
			respondServiceError(c, err, errorMessages{
				NotFound:  "publicación no encontrada",
				Forbidden: "no tienes permiso para marcar esta publicación como vendida",
				Conflict:  "solo se pueden marcar como vendidas las publicaciones que estén publicadas",
				Internal:  "no se pudo marcar la venta",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "se marcó la publicación como vendida",
			"vehicle": vehicle,
		})
	}
}
