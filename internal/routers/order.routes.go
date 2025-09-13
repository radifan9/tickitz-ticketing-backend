package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/radifan9/tickitz-ticketing-backend/internal/handlers"
	"github.com/radifan9/tickitz-ticketing-backend/internal/middlewares"
	"github.com/radifan9/tickitz-ticketing-backend/internal/repositories"
)

func RegisterOrderRoutes(v1 *gin.RouterGroup, db *pgxpool.Pool) {
	orderRepo := repositories.NewOrderRepository(db)
	orderHandler := handlers.NewOrderHandler(orderRepo)
	orders := v1.Group("/orders")
	orders.Use(middlewares.VerifyToken, middlewares.Access("user"))

	orders.POST("/", orderHandler.AddTransaction)
	orders.GET("/histories", orderHandler.ListTransaction)
}
