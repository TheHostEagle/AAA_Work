package main

import (
	"log"
	"net/http"

	"automarket/internal/handlers"
	"automarket/internal/middleware"
	"automarket/internal/models"
	"automarket/internal/security"
	"automarket/internal/service"
	"automarket/internal/storage"

	"github.com/gin-gonic/gin"
)

func main() {
	store, err := storage.New("data")
	if err != nil {
		panic(err)
	}
	if err := store.VerifyAuditChain(); err != nil {
		log.Fatalf("verificando cadena de auditoría: %v", err)
	}
	if err := seedAdmin(store); err != nil {
		log.Fatalf("creando administrador inicial: %v", err)
	}

	tokenStore := security.NewTokenStore()
	authService := service.NewAuthService(store, tokenStore)
	vehicleService := service.NewVehicleService(store)
	adminService := service.NewAdminService(store)

	r := gin.New()
	r.Use(gin.Logger(), middleware.AccountingMiddleware(store), gin.Recovery())
	r.LoadHTMLFiles("web/templates/index.html")
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	api := r.Group("/api")
	{
		api.POST("/auth/register", handlers.RegisterHandler(authService))
		api.POST("/auth/login", handlers.LoginHandler(authService))

		api.GET("/vehicles", handlers.GetPublicVehiclesHandler(vehicleService))

		sellerGroup := api.Group("/vehicles")
		sellerGroup.Use(middleware.AuthenticationMiddleware(tokenStore))
		sellerGroup.Use(middleware.AuthorizationMiddleware(store, models.RoleSeller))
		{
			sellerGroup.POST("", handlers.CreateVehicleHandler(vehicleService))
			sellerGroup.PUT("/:id/sold", handlers.MarkVehicleAsSoldHandler(vehicleService))
		}

		adminGroup := api.Group("/admin")
		adminGroup.Use(middleware.AuthenticationMiddleware(tokenStore))
		adminGroup.Use(middleware.AuthorizationMiddleware(store, models.RoleAdmin))
		{
			adminGroup.PUT("/vehicles/:id/approve", handlers.ApproveVehicleHandler(adminService))
			adminGroup.DELETE("/vehicles/:id", handlers.DeleteVehicleHandler(adminService))
			adminGroup.DELETE("/users/:id", handlers.DeleteUserHandler(adminService))
		}
	}
	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}
