package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/radifan9/tickitz-ticketing-backend/internal/handlers"
	"github.com/radifan9/tickitz-ticketing-backend/internal/middlewares"
	"github.com/radifan9/tickitz-ticketing-backend/internal/repositories"
)

func RegisterMovieRoutes(v1 *gin.RouterGroup, db *pgxpool.Pool) {
	movieRepo := repositories.NewMovieRepository(db)
	movieHandler := handlers.NewMovieHandler(movieRepo)

	v1.GET("/movies/upcoming", movieHandler.ListUpcomingMovies)
	v1.GET("/movies/", movieHandler.ListFilteredMovies)

	// Only Admin can do this
	v1.PATCH("/movies/:id/archive", middlewares.VerifyToken, middlewares.Access("admin"), movieHandler.GoArchiveAMovie)
}
