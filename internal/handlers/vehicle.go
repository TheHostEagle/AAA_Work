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
func CreateVehicleHandler(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok", "message": "Create vehicle endpoint stub"})
}

func MarkVehicleAsSoldHandler(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok", "message": "Mark vehicle as sold endpoint stub"})
}
