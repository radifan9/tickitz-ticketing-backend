package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/radifan9/tickitz-ticketing-backend/internal/handlers"
	"github.com/radifan9/tickitz-ticketing-backend/internal/middlewares"
	"github.com/radifan9/tickitz-ticketing-backend/internal/repositories"
	"github.com/redis/go-redis/v9"
)

func RegisterUserRoutes(v1 *gin.RouterGroup, db *pgxpool.Pool, rdb *redis.Client) {
	userRepo := repositories.NewUserRepository(db, rdb)
	userHandler := handlers.NewUserHandler(userRepo, rdb)
	verifyTokenWithBlacklist := middlewares.VerifyTokenWithBlacklist(rdb) // Create middleware instance with redis client

	// Authentication routes (no auth required)
	auth := v1.Group("/auth")
	{
		auth.POST("/register", userHandler.Register) // POST /api/v1/auth/register
		auth.POST("/login", userHandler.Login)       // POST /api/v1/auth/login
		auth.DELETE("/logout", verifyTokenWithBlacklist, userHandler.Logout)
	}

	// User profile routes (auth required)
	users := v1.Group("/users")
	users.Use(verifyTokenWithBlacklist, middlewares.Access("admin", "user"))
	{
		users.GET("/profile", userHandler.GetProfile)        // GET /api/v1/users/profile
		users.PATCH("/profile", userHandler.EditProfile)     // PATCH /api/v1/users/profile
		users.PATCH("/password", userHandler.ChangePassword) // PATCH /api/v1/users/password
	}
}
