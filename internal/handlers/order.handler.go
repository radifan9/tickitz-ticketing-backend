package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/radifan9/tickitz-ticketing-backend/internal/models"
	"github.com/radifan9/tickitz-ticketing-backend/internal/repositories"
)

// or : order repository
type OrderHandler struct {
	or *repositories.OrderRepository
}

func NewOrderHandler(or *repositories.OrderRepository) *OrderHandler {
	return &OrderHandler{or: or}
}

func (o *OrderHandler) ListSchedules(ctx *gin.Context) {

	movieID := ctx.Query("movie_id")
	cityID := ctx.Query("city_id")
	showTimeID := ctx.Query("show_time_id")

	log.Println("movieId: ", movieID)
	log.Println("cityId: ", cityID)
	log.Println("showTimeID: ", showTimeID)

	schedules, err := o.or.FilterSchedule(ctx, movieID, cityID, showTimeID)
	if err != nil {
		ctx.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(200, schedules)
}

func (o *OrderHandler) AddTransaction(ctx *gin.Context) {
	var body models.Transaction
	if err := ctx.ShouldBind(&body); err != nil {
		log.Println("Internal server error.\nCause: ", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Internal server error",
		})
		return
	}

	log.Println(body.UserID)
	log.Println(body.PaymentID)
	log.Println(body.TotalPayment)

	transaction, err := o.or.AddNewTransactions(ctx, body)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success":     true,
		"transaction": transaction,
	})
}
