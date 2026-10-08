package main

import (
	"automarket/internal/handlers"
	"automarket/internal/middleware"
	"automarket/internal/security"
	"automarket/internal/storage"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	store, err := storage.New("data")
	if err != nil {
		panic(err)
	}

	tokenStore := security.NewTokenStore()

	r.Use(middleware.AccountingMiddleware(store))

	api := r.Group("/api")
	{
		api.POST("/auth/register", handlers.RegisterHandler(store))
		api.POST("/auth/login", handlers.LoginHandler(store, tokenStore))

		api.GET("/vehicles", handlers.GetPublicVehiclesHandler(store))

		venderGroup := api.Group("/vehicles")
		venderGroup.Use(middleware.AuthenticationMiddleware(tokenStore))
		{
			venderGroup.POST("", handlers.CreateVehicleHandler)
			venderGroup.PUT("/:id/sold", handlers.MarkVehicleAsSoldHandler)
		}

		adminGroup := api.Group("/admin")
		adminGroup.Use(middleware.AuthenticationMiddleware(tokenStore))
		adminGroup.Use(middleware.AuthorizationMiddleware(store, "administrador"))
		{
			adminGroup.PUT("/vehicles/:id/approve", handlers.ApproveVehicleHandler(store))
			adminGroup.DELETE("/vehicles/:id", handlers.DeleteVehicleHandler(store))
			adminGroup.DELETE("/users/:id", handlers.DeleteUserHandler(store))
		}
	}
	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}
