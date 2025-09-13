package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/radifan9/tickitz-ticketing-backend/internal/handlers"
	"github.com/radifan9/tickitz-ticketing-backend/internal/middlewares"
	"github.com/radifan9/tickitz-ticketing-backend/internal/repositories"
)

func RegisterUserRoutes(v1 *gin.RouterGroup, db *pgxpool.Pool) {
	userRepo := repositories.NewUserRepository(db)
	userHandler := handlers.NewUserHandler(userRepo)

	v1.POST("/register", userHandler.Register)
	v1.POST("/login", userHandler.Login)
	v1.GET("/profile", middlewares.VerifyToken,
		middlewares.Access("admin", "user"), userHandler.GetProfile)
	v1.PATCH("/profile", middlewares.VerifyToken,
		middlewares.Access("admin", "user"), userHandler.EditProfile)

}
