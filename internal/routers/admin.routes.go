package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/radifan9/tickitz-ticketing-backend/internal/handlers"
	"github.com/radifan9/tickitz-ticketing-backend/internal/middlewares"
	"github.com/radifan9/tickitz-ticketing-backend/internal/repositories"
)

func RegisterAdminRoutes(v1 *gin.RouterGroup, db *pgxpool.Pool) {
	adminRepo := repositories.NewMovieRepository(db)
	adminHandler := handlers.NewMovieHandler(adminRepo)
	admin := v1.Group("/admin")

	admin.GET("/movies",
		middlewares.VerifyToken,
		middlewares.Access("admin"),
		adminHandler.ListAllMovies)

	admin.DELETE("/movies/:id/archive",
		middlewares.VerifyToken,
		middlewares.Access("admin"),
		adminHandler.ArchiveMovieByID)
}
