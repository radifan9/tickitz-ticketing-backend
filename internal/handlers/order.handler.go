package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/radifan9/tickitz-ticketing-backend/internal/models"
	"github.com/radifan9/tickitz-ticketing-backend/internal/repositories"
	"github.com/radifan9/tickitz-ticketing-backend/internal/utils"
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
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
	}

	schedules, err := o.or.FilterSchedule(ctx, queryParams)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"sucess": false,
			"error":  err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    schedules,
	})
}

func (o *OrderHandler) AddTransaction(ctx *gin.Context) {
	var body models.Transaction
	if err := ctx.ShouldBind(&body); err != nil {
		log.Println("Internal server error.\nCause: ", err.Error())
		utils.HandleResponse(ctx, http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Status:  http.StatusInternalServerError,
			Error:   err.Error(),
		})
		return
	}

	log.Println(body.UserID)
	log.Println(body.PaymentID)
	log.Println(body.TotalPayment)
	log.Println(body.Seats)

	transaction, err := o.or.AddNewTransactionsAndSeatCodes(ctx, body)
	if err != nil {
		log.Println("Error : ", err.Error())
		utils.HandleResponse(ctx, http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Status:  http.StatusInternalServerError,
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success":     true,
		"transaction": transaction,
	})
}
