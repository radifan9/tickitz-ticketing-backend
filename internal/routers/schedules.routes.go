package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/radifan9/tickitz-ticketing-backend/internal/handlers"
	"github.com/radifan9/tickitz-ticketing-backend/internal/repositories"
)

func RegisterSchedulesRoutes(v1 *gin.RouterGroup, db *pgxpool.Pool) {
	scheduleRepo := repositories.NewScheduleRepository(db)
	scheduleHandler := handlers.NewScheduleHandler(scheduleRepo)
	schedules := v1.Group("/schedules")

	schedules.GET("", scheduleHandler.ListSchedules)
	schedules.GET("/cinemas", scheduleHandler.ListCinemas)
	schedules.GET("/:id/sold-seats", scheduleHandler.GetSoldSeatsByScheduleID)

}
