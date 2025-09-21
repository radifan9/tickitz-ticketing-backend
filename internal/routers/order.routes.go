package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/radifan9/tickitz-ticketing-backend/internal/handlers"
	"github.com/radifan9/tickitz-ticketing-backend/internal/middlewares"
	"github.com/radifan9/tickitz-ticketing-backend/internal/repositories"
	"github.com/redis/go-redis/v9"
)

func RegisterOrderRoutes(v1 *gin.RouterGroup, db *pgxpool.Pool, rdb *redis.Client) {
	orderRepo := repositories.NewOrderRepository(db, rdb)
	orderHandler := handlers.NewOrderHandler(orderRepo)
	VerifyTokenWithBlacklist := middlewares.VerifyTokenWithBlacklist(rdb)

	orders := v1.Group("/orders")
	orders.Use(VerifyTokenWithBlacklist, middlewares.Access("user"))

	orders.POST("", orderHandler.AddTransaction)
	orders.PATCH("/transactions/:id", orderHandler.PayTransaction)
	orders.GET("/histories", orderHandler.ListTransaction)
}
