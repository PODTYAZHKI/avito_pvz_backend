package routes

import (
	"github.com/gin-gonic/gin"

	"avito-pvz/internal/config"
	"avito-pvz/internal/handler"
)

func SetupRoutes(
	cfg *config.Config,
	r *gin.Engine,
	userHandler handler.UserHandler,
	pvzHandler handler.PvzHandler,
	receptionHandler handler.ReceptionHandler,
	productHandler handler.ProductHandler,
	authMiddleware gin.HandlerFunc,
	moderatorOnly gin.HandlerFunc,
	employeeOnly gin.HandlerFunc,
) {
	r.POST("/dummyLogin", userHandler.DummyLogin)
	r.POST("/register", userHandler.Register)
	r.POST("/login", userHandler.Login)

	pvzGroup := r.Group("/pvz")
	{
		pvzGroup.POST("", authMiddleware, moderatorOnly, pvzHandler.CreatePvz)
		pvzGroup.GET("", authMiddleware, pvzHandler.GetPvzs)
		pvzGroup.POST("/:pvzId/close_latest_reception", authMiddleware, employeeOnly, receptionHandler.CloseActiveReception)
		pvzGroup.POST("/:pvzId/delete_last_product", authMiddleware, employeeOnly, productHandler.DeleteLastProduct)
	}

	receptionGroup := r.Group("/receptions")
	{
		receptionGroup.POST("", authMiddleware, employeeOnly, receptionHandler.CreateReception)
	}
	productsGroup := r.Group("/products")
	{
		productsGroup.POST("", authMiddleware, employeeOnly, productHandler.AddProduct)
	}
}
