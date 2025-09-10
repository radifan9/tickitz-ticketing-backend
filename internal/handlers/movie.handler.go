package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/radifan9/tickitz-ticketing-backend/internal/models"
	"github.com/radifan9/tickitz-ticketing-backend/internal/repositories"
	"github.com/radifan9/tickitz-ticketing-backend/internal/utils"
)

// mr : movie repository
type MovieHandler struct {
	mr *repositories.MovieRepository
}

func NewMovieHandler(mr *repositories.MovieRepository) *MovieHandler {
	return &MovieHandler{mr: mr}
}

func (m *MovieHandler) ListUpcomingMovies(ctx *gin.Context) {
	upcomingMovies, err := m.mr.ListUpcomingMovies(ctx)
	if err != nil {
		utils.HandleError(ctx, http.StatusOK, err.Error(), "failed to list upcoming movies")
		return
	}

	utils.HandleResponse(ctx, http.StatusOK, models.SuccessResponse{
		Success: true,
		Status:  http.StatusOK,
		Data:    upcomingMovies,
	})
}

func (m *MovieHandler) ListPopularMovies(ctx *gin.Context) {
	popularMovies, err := m.mr.ListPopularMovies(ctx)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, err.Error(), "failed to list popular movies")
		return
	}

	utils.HandleResponse(ctx, http.StatusOK, models.SuccessResponse{
		Success: true,
		Status:  http.StatusOK,
		Data:    popularMovies,
	})
}

func (m *MovieHandler) ListFilteredMovies(ctx *gin.Context) {
	keywordParam := ctx.Query("keywords")
	genreParam := ctx.Query("genres")
	offsetParam := ctx.Query("offset")
	limitParam := ctx.Query("limit")

	// Convert to []string
	keywords := []string{}
	if keywordParam != "" {
		keywords = strings.Split(keywordParam, ",")
	}

	// Convert to []int
	genres := []int{}
	if genreParam != "" {
		genreStrings := strings.Split(genreParam, ",")
		for _, g := range genreStrings {
			if id, err := strconv.Atoi(g); err == nil {
				genres = append(genres, id)
			}
		}
	}

	// Convert offset & limit
	offset, _ := strconv.Atoi(offsetParam)
	limit, _ := strconv.Atoi(limitParam)

	// Set sensible defaults
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 20
	}

	log.Println("keywords : ", keywords)
	log.Println("genres : ", genres)
	log.Println("offset : ", offset)
	log.Println("limit : ", limit)

	// Build filter struct
	filter := models.MovieFilter{
		Keywords: keywords,
		Genres:   genres,
		Offset:   offset,
		Limit:    limit,
	}

	// Call repo
	movies, err := m.mr.ListMovieFiltered(ctx, filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"body":    movies,
	})
}

func (m *MovieHandler) ArchiveMovieByID(ctx *gin.Context) {
	movieId := ctx.Param("id")

	archievedMovieId, err := m.mr.ArchiveMovieByID(ctx, movieId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"body":    0,
		})
	}

	utils.HandleResponse(ctx, http.StatusOK, models.SuccessResponse{
		Success: true,
		Status:  http.StatusOK,
		Data:    archievedMovieId,
	})
}

// --- Get Movie Details
func (m *MovieHandler) GetMovieDetails(ctx *gin.Context) {
	movieID := ctx.Param("id")

	movie, err := m.mr.GetMovieDetails(ctx, movieID)
	if err != nil {
		log.Println("error : ", err)
		utils.HandleResponse(ctx, http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Status:  http.StatusInternalServerError,
			Error:   err.Error(),
		})
		return
	}

	utils.HandleResponse(ctx, http.StatusOK, models.SuccessResponse{
		Success: true,
		Status:  http.StatusOK,
		Data:    movie,
	})
}

// (Admin) List All Movies
func (m *MovieHandler) ListAllMovies(ctx *gin.Context) {
	allMovies, err := m.mr.ListAllMovies(ctx)
	if err != nil {
		utils.HandleResponse(ctx, http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Status:  http.StatusInternalServerError,
			Error:   err.Error(),
		})
		return
	}

	utils.HandleResponse(ctx, http.StatusOK, models.SuccessResponse{
		Success: true,
		Status:  http.StatusOK,
		Data:    allMovies,
	})
}
