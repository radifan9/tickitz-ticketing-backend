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

	movies := v1.Group("/movies")
	{
		movies.GET("/upcoming", movieHandler.ListUpcomingMovies)
		movies.GET("/", movieHandler.ListFilteredMovies)
	}

	// Only Admin can do this
	admin := v1.Group("/admin")
	{
		admin.GET("/movies",
			middlewares.VerifyToken,
			middlewares.Access("admin"),
			movieHandler.GetAllMovies)

		admin.DELETE("/movies/:id/archive",
			middlewares.VerifyToken,
			middlewares.Access("admin"),
			movieHandler.GoArchiveAMovie)
	}

}
