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
	pageParam := ctx.Query("page")

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

	// Convert page
	// If no page in param, then page = 0
	page, _ := strconv.Atoi(pageParam)

	// Set sensible defaults
	if page <= 0 {
		page = 1
	}

	// Calculate Limit & Offset based on page
	limit := 20
	offset := (page - 1) * limit

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
		utils.HandleError(ctx, http.StatusInternalServerError, err.Error(), "cannot get filtered movies")
		return
	}

	utils.HandleResponse(ctx, http.StatusOK, models.SuccessResponse{
		Success: true,
		Status:  http.StatusOK,
		Data:    movies,
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
