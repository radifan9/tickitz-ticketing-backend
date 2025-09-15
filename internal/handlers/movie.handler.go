package handlers

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

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

// @Summary Get upcoming movies
// @Tags    Movies
// @Produce json
// @Success 200 {object} models.SuccessResponse
// @Router  /api/v1/movies/upcoming [get]
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

// @Summary Get popular movies
// @Tags    Movies
// @Produce json
// @Success 200 {object} models.SuccessResponse
// @Router  /api/v1/movies/popular [get]
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

// @Summary Get filtered movies
// @Tags    Movies
// @Produce json
// @Param   keywords query string false "Comma-separated keywords"
// @Param   genres   query string false "Comma-separated genre IDs"
// @Param   page     query int    false "Page number"
// @Success 200 {object} models.SuccessResponse
// @Router  /api/v1/movies/ [get]
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

// @Summary Get movie details
// @Tags    Movies
// @Produce json
// @Param   id path string true "Movie ID"
// @Success 200 {object} models.SuccessResponse
// @Router  /api/v1/movies/{id} [get]
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

// @Summary List all movies (admin)
// @Tags    Admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.SuccessResponse
// @Router  /api/v1/admin/movies [get]
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

// @Summary Archive movie by ID (admin)
// @Tags    Admin
// @Produce json
// @Security BearerAuth
// @Param   id path string true "Movie ID"
// @Success 200 {object} models.SuccessResponse
// @Router  /api/v1/admin/movies/{id}/archive [delete]
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

// @Summary Create a new movie (admin)
// @Tags    Admin
// @Accept  json
// @Produce json
// @Security BearerAuth
// @Param   body body models.CreateMovie true "Movie data"
// @Success 200 {object} models.SuccessResponse
// @Router  /api/v1/admin/movies [post]
func (m *MovieHandler) CreateMovie(ctx *gin.Context) {
	var body models.CreateMovie
	if err := ctx.ShouldBind(&body); err != nil {
		utils.HandleError(ctx, http.StatusBadRequest, "bad request", err.Error())
		return
	}

	// From postman must upload a new poster
	filePoster := body.PosterImg
	var locationPoster string

	if filePoster != nil {
		ext := filepath.Ext(filePoster.Filename)
		re := regexp.MustCompile(`(?i)\.(png|jpg|jpeg|webp)$`)
		if !re.MatchString(ext) {
			utils.HandleError(ctx, http.StatusBadRequest, "invalid file type", "only png, jpg, jpeg, webp allowed")
			return
		}

		filenamePoster := fmt.Sprintf("%d_images_%s%s", time.Now().UnixNano(), "userID", ext) // replace "userID"
		locationPoster = filepath.Join("public", filenamePoster)

		if err := ctx.SaveUploadedFile(filePoster, locationPoster); err != nil {
			utils.HandleError(ctx, http.StatusBadRequest, err.Error(), "failed to upload")
			return
		}
	}

	fileBackdrop := body.BackdropImg
	var locationBackdrop string

	if fileBackdrop != nil {
		ext := filepath.Ext(fileBackdrop.Filename)
		re := regexp.MustCompile(`(?i)\.(png|jpg|jpeg|webp)$`)
		if !re.MatchString(ext) {
			utils.HandleError(ctx, http.StatusBadRequest, "invalid file type", "only png, jpg, jpeg, webp allowed")
			return
		}

		filenamePoster := fmt.Sprintf("%d_images_%s%s", time.Now().UnixNano(), "userID", ext) // replace "userID"
		locationBackdrop = filepath.Join("public", filenamePoster)

		if err := ctx.SaveUploadedFile(fileBackdrop, locationBackdrop); err != nil {
			utils.HandleError(ctx, http.StatusBadRequest, err.Error(), "failed to upload")
			return
		}
	}

	newM, err := m.mr.CreateMovie(ctx, body, locationPoster, locationBackdrop)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "status internal server error", err.Error())
		return
	}

	log.Println("Newly created movie : ", newM)
	utils.HandleResponse(ctx, http.StatusOK, models.SuccessResponse{
		Success: true,
		Status:  http.StatusOK,
		Data:    newM,
	})
}
