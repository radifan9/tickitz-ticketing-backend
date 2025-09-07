package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/radifan9/tickitz-ticketing-backend/internal/repositories"
)

// mr : movie repository
type MovieHandler struct {
	mr *repositories.MovieRepository
}

func NewMovieHandler(mr *repositories.MovieRepository) *MovieHandler {
	return &MovieHandler{mr: mr}
}

func (m *MovieHandler) ListUpcomingMovies(ctx *gin.Context) {
	upcomingMovies, err := m.mr.ListUpcomingMovie(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"body":    upcomingMovies,
	})
}
