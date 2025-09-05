package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/radifan9/tickitz-ticketing-backend/internal/models"
)

func InitRouter(db *pgxpool.Pool) *gin.Engine {
	router := gin.Default()

	// Setup routing

	// Catch all route
	router.NoRoute(func(ctx *gin.Context) {
		ctx.JSON(http.StatusNotFound, models.Response{
			Message: "Not Found",
			Status:  "Route does not exist",
		})
	})

	return router
}
