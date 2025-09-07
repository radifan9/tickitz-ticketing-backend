package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/radifan9/tickitz-ticketing-backend/internal/handlers"
	"github.com/radifan9/tickitz-ticketing-backend/internal/models"
	"github.com/radifan9/tickitz-ticketing-backend/internal/repositories"
)

func InitRouter(db *pgxpool.Pool) *gin.Engine {
	router := gin.Default()

	userRepo := repositories.NewUserRepository(db)
	userHandler := handlers.NewUserHandler((*repositories.UserRepository)(userRepo))

	// API Version 1
	v1 := router.Group("/api/v1")
	{
		v1.POST("/register", userHandler.Register)
	}

	// Catch all route
	router.NoRoute(func(ctx *gin.Context) {
		ctx.JSON(http.StatusNotFound, models.Response{
			Message: "Not Found",
			Status:  "Route does not exist",
		})
	})

	return router
}
