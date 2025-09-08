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

	v1.GET("/order/schedules", orderHandler.ListSchedules)
	v1.POST("/order", orderHandler.AddTransaction)

}
