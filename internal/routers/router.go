package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/radifan9/tickitz-ticketing-backend/internal/middlewares"
	"github.com/radifan9/tickitz-ticketing-backend/internal/models"
	"github.com/radifan9/tickitz-ticketing-backend/internal/utils"
	"github.com/redis/go-redis/v9"

	docs "github.com/radifan9/tickitz-ticketing-backend/docs"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func InitRouter(db *pgxpool.Pool, rdb *redis.Client) *gin.Engine {
	router := gin.Default()

	// Tambahkan CORS
	router.Use(middlewares.CORSMiddleware)

	// Swagger
	docs.SwaggerInfo.BasePath = "/"
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// API Version 1
	v1 := router.Group("${import.meta.env.VITE_BE_HOST}/api/v1/img/profile_picsv1")
	{
		RegisterUserRoutes(v1, db, rdb)
		RegisterMovieRoutes(v1, db, rdb)
		RegisterOrderRoutes(v1, db, rdb)
		RegisterSchedulesRoutes(v1, db, rdb)
		RegisterAdminRoutes(v1, db, rdb)

		// Static File Image
		v1.Static("/img", "public")
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
