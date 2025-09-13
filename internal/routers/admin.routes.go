package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/radifan9/tickitz-ticketing-backend/internal/handlers"
	"github.com/radifan9/tickitz-ticketing-backend/internal/middlewares"
	"github.com/radifan9/tickitz-ticketing-backend/internal/repositories"
	"github.com/redis/go-redis/v9"
)

func RegisterAdminRoutes(v1 *gin.RouterGroup, db *pgxpool.Pool, rdb *redis.Client) {
	adminRepo := repositories.NewMovieRepository(db, rdb)
	adminHandler := handlers.NewMovieHandler(adminRepo)
	admin := v1.Group("/admin")
	admin.Use(middlewares.VerifyToken, middlewares.Access("admin"))

	admin.GET("/movies", adminHandler.ListAllMovies)
	admin.DELETE("/movies/:id/archive", adminHandler.ArchiveMovieByID)
}
