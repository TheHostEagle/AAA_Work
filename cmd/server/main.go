package main

import (
	"automarket/internal/handlers"
	"automarket/internal/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Aplicar el middleware de Auditoría (Accounting) de manera global para cumplir con el RF-01
	r.Use(middleware.AccountingMiddleware())

	api := r.Group("/api")
	{
		// Rutas públicas (Sin autenticación requerida)
		api.POST("/auth/register", handlers.RegisterHandler)
		api.POST("/auth/login", handlers.LoginHandler)
		api.GET("/vehicles", handlers.GetPublicVehiclesHandler) // HU-VIS-01 (Visitante, Vendedor, Admin)

		// Rutas protegidas para Vendedores
		venderGroup := api.Group("/vehicles")
		venderGroup.Use(middleware.AuthenticationMiddleware())
		{
			venderGroup.POST("", handlers.CreateVehicleHandler)           // HU-VEN-01
			venderGroup.PUT("/:id/sold", handlers.MarkVehicleAsSoldHandler) // HU-VEN-01
		}

		// Rutas protegidas para Administradores
		adminGroup := api.Group("/admin")
		adminGroup.Use(middleware.AuthenticationMiddleware())
		{
			adminGroup.PUT("/vehicles/:id/approve", handlers.ApproveVehicleHandler) // HU-ADM-01
			adminGroup.DELETE("/vehicles/:id", handlers.DeleteVehicleHandler)       // HU-ADM-01
			adminGroup.DELETE("/users/:id", handlers.DeleteUserHandler)             // HU-ADM-01
		}
	}

	// Iniciar servidor HTTP en el puerto 8080 (Restricción: solo HTTP sin TLS)
	r.Run(":8080")
}