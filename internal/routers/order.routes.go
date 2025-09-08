package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/radifan9/tickitz-ticketing-backend/internal/handlers"
	"github.com/radifan9/tickitz-ticketing-backend/internal/repositories"
)

func RegisterOrderRoutes(v1 *gin.RouterGroup, db *pgxpool.Pool) {
	orderRepo := repositories.NewOrderRepository(db)
	orderHandler := handlers.NewOrderHandler(orderRepo)

	order := v1.Group("/order")
	{
		order.GET("/schedules", orderHandler.ListSchedules)
		order.POST("/", orderHandler.AddTransaction)

	}

}
