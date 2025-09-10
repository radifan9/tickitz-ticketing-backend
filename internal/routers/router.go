package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/radifan9/tickitz-ticketing-backend/internal/middlewares"
	"github.com/radifan9/tickitz-ticketing-backend/internal/models"
	"github.com/radifan9/tickitz-ticketing-backend/internal/utils"

	docs "github.com/radifan9/tickitz-ticketing-backend/docs"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func InitRouter(db *pgxpool.Pool) *gin.Engine {
	router := gin.Default()

	// Tambahkan cors
	router.Use(middlewares.CORSMiddleware)

	// Swagger
	docs.SwaggerInfo.BasePath = "/"
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// API Version 1
	v1 := router.Group("/api/v1")
	{
		RegisterUserRoutes(v1, db)
		RegisterMovieRoutes(v1, db)
		RegisterOrderRoutes(v1, db)
		RegisterAdminRoutes(v1, db)
	}

	// Catch all route
	router.NoRoute(func(ctx *gin.Context) {
		utils.HandleResponse(ctx, http.StatusNotFound, models.ErrorResponse{
			Success: false,
			Status:  http.StatusNotFound,
			Error:   "Route does not exist",
		})
	})

	return router
}
