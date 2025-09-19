package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/radifan9/tickitz-ticketing-backend/internal/models"
	"github.com/radifan9/tickitz-ticketing-backend/internal/repositories"
	"github.com/radifan9/tickitz-ticketing-backend/internal/utils"
)

type ScheduleHandler struct {
	sr *repositories.ScheduleRepository
}

func NewScheduleHandler(sr *repositories.ScheduleRepository) *ScheduleHandler {
	return &ScheduleHandler{sr: sr}
}

func (s *ScheduleHandler) ListCinemas(ctx *gin.Context) {
	cinemas, err := s.sr.ListCinemas(ctx)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "status internal error", err.Error())
		return
	}

	utils.HandleResponse(ctx, http.StatusOK, models.SuccessResponse{
		Success: true,
		Status:  http.StatusOK,
		Data:    cinemas,
	})
}

func (s *ScheduleHandler) ListSchedules(ctx *gin.Context) {
	var queryParams models.ScheduleFilter
	if err := ctx.ShouldBindQuery(&queryParams); err != nil {
		utils.HandleError(ctx, http.StatusBadRequest, err.Error(), "failed to get schedule")
		return
	}

	schedules, err := s.sr.FilterSchedule(ctx, queryParams)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, err.Error(), "failed to get schedule")
		return
	}

	utils.HandleResponse(ctx, http.StatusOK, models.SuccessResponse{
		Success: true,
		Status:  http.StatusOK,
		Data:    schedules,
	})
}

func (s *ScheduleHandler) GetSoldSeatsByScheduleID(ctx *gin.Context) {
	scheduleID := ctx.Param("id")

	soldSeats, err := s.sr.GetSoldSeatsByScheduleID(ctx, scheduleID)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, err.Error(), "failed to get sold seats by schedule_id")
		return
	}

	utils.HandleResponse(ctx, http.StatusOK, models.SuccessResponse{
		Success: true,
		Status:  http.StatusOK,
		Data:    soldSeats,
	})
}
