package handlers

import (
	"net/http"

	"automarket/internal/models"
	"automarket/internal/storage"

	"github.com/gin-gonic/gin"
)

// publicVehicle evita exponer campos internos del vendedor, como PasswordHash.
type publicVehicle struct {
	models.Vehicle
	Seller publicSeller `json:"seller"`
}

type publicSeller struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// GetPublicVehiclesHandler muestra solo vehículos publicados y el contacto de su vendedor.
func GetPublicVehiclesHandler(store *storage.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		vehicles := store.ListVehiclesByStatus("publicada")
		response := make([]publicVehicle, 0, len(vehicles))

		for _, vehicle := range vehicles {
			seller, err := store.GetUserByID(vehicle.SellerID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo obtener el contacto del vendedor"})
				return
			}

			response = append(response, publicVehicle{
				Vehicle: vehicle,
				Seller: publicSeller{
					Name:  seller.Name,
					Email: seller.Email,
				},
			})
		}

		c.JSON(http.StatusOK, response)
	}
}

//CreateVehicleHandler permite registrar una publicación nueva como "pendiente" y asociarla al vendedor autenticado.
func CreateVehicleHandler(store *storage.Store) gin.HandlerFunc {
	return func(c *gin.Context){
		var input struct {
			Title	string  `json:"title"`
			Brand	string	`json:"brand"`
			Model	string	`json:"model"`
			Price	float64	`json:"price"`
			Description	string	`json:"description"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "el JSON no es válido",
			})
			return
		}

		if input.Title == "" || input.Brand == "" || input.Model == "" || input.Price <= 0 {	
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "título, marca, modelo y precio (mayor a cero) son obligatorios",
			})
			return
		}

		vehicle := models.Vehicle{
			SellerID: c.GetString("userID"),
			Title: input.Title,
			Brand: input.Brand,
			Model: input.Model,
			Price: input.Price,
			Description: input.Description,
		}

		if err := store.CreateVehicle(&vehicle); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "no se pudo crear la publicación",
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H {
			"message": "publicación creada, pendiente de aprobación",
			"vehicle": vehicle,
		})
	}
}

// MarkVehicleAsSoldHandler permite al vendedor dueño reportar la venta de una publicación que ya está disponible publicamente
func MarkVehicleAsSoldHandler(store *storage.Store) gin.HandlerFunc {
	return func (c *gin.Context) {
		vehicle, err := store.GetVehicle(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "publicación no encontrada",
			})
			return
		}

		if vehicle.SellerID != c.GetString("userID") {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "no tienes permiso para marcar esta publicación como vendida",
			})
			return
		}

		if vehicle.Status != "publicada" {
			c.JSON(http.StatusConflict, gin.H {
				"error": "solo se pueden marcar como vendidas las publicaciones que estén publicadas",
			})
			return
		}

		vehicle.Status = "vendido"
		if err := store.UpdateVehicle(vehicle); err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "no se pudo marcar la venta",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "se marcó la publicación como vendida",
			"vehicle": vehicle,
		})
	}
}

