package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/radifan9/tickitz-ticketing-backend/internal/handlers"
	"github.com/radifan9/tickitz-ticketing-backend/internal/repositories"
	"github.com/redis/go-redis/v9"
)

func RegisterMovieRoutes(v1 *gin.RouterGroup, db *pgxpool.Pool, rdb *redis.Client) {
	movieRepo := repositories.NewMovieRepository(db, rdb)
	movieHandler := handlers.NewMovieHandler(movieRepo)
	movies := v1.Group("/movies")

	movies.GET("/", movieHandler.ListFilteredMovies)
	movies.GET("/:id", movieHandler.GetMovieDetails)

	// Sub-resources for movies
	movies.GET("/upcoming", movieHandler.ListUpcomingMovies)
	movies.GET("/popular", movieHandler.ListPopularMovies)

}
