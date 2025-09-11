package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/radifan9/tickitz-ticketing-backend/internal/models"
	"github.com/radifan9/tickitz-ticketing-backend/internal/repositories"
	"github.com/radifan9/tickitz-ticketing-backend/internal/utils"
	"github.com/radifan9/tickitz-ticketing-backend/pkg"
)

// or : order repository
type OrderHandler struct {
	or *repositories.OrderRepository
}

func NewOrderHandler(or *repositories.OrderRepository) *OrderHandler {
	return &OrderHandler{or: or}
}

func (o *OrderHandler) ListSchedules(ctx *gin.Context) {
	var queryParams models.ScheduleFilter
	if err := ctx.ShouldBindQuery(&queryParams); err != nil {
		utils.HandleError(ctx, http.StatusBadRequest, err.Error(), "failed to get schedule")
		return
	}

	schedules, err := o.or.FilterSchedule(ctx, queryParams)
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

func (o *OrderHandler) GetSoldSeatsByScheduleID(ctx *gin.Context) {
	scheduleID := ctx.Param("id")

	soldSeats, err := o.or.GetSoldSeatsByScheduleID(ctx, scheduleID)
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

// --- Method used in Payment Page, when user clicked "Check Payment"
func (o *OrderHandler) AddTransaction(ctx *gin.Context) {
	var body models.Transaction
	if err := ctx.ShouldBind(&body); err != nil {
		log.Println("error : ", err.Error())
		utils.HandleResponse(ctx, http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Status:  http.StatusInternalServerError,
			Error:   err.Error(),
		})
		return
	}

	transaction, err := o.or.AddNewTransactionsAndSeatCodes(ctx, body)
	if err != nil {
		log.Println("error : ", err.Error())
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
		Data:    transaction,
	})
}

// --- Method used in profile "Order History"
func (o *OrderHandler) ListTransaction(ctx *gin.Context) {
	// Get the userID from token
	claims, _ := ctx.Get("claims")
	user, ok := claims.(pkg.Claims)
	if !ok {
		utils.HandleError(ctx, http.StatusInternalServerError, "internal server error", "cannot cast into pkg.claims")
		return
	}

	tHistories, err := o.or.ListTransaction(ctx, user.UserId)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, err.Error(), "cannot get list of transaction")
		return
	}

	utils.HandleResponse(ctx, http.StatusOK, models.SuccessResponse{
		Success: true,
		Status:  http.StatusOK,
		Data:    tHistories,
	})
}
