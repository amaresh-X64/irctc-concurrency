package booking

import (
	"net/http"
	"strconv"

	"gin-booking/internal/booking/dto"
	"gin-booking/internal/constants"
	"gin-booking/internal/helpers"

	"github.com/gin-gonic/gin"
)

type BookingService interface {
	BookSeat(req dto.BookingRequest) (*dto.BookingResponse, error, bool)
	CancelBooking(req dto.CancelRequest) (*dto.CancelResponse, error)
	GetUserBookings(userID int) ([]Booking, error)
	IsSeatLocked(trainID, seatID int, journeyDate string) bool
	IsSeatAvailableForDate(seatID int, journeyDate string) bool
}

type Controller struct {
	service BookingService
}

func NewController(service BookingService) *Controller {
	return &Controller{service: service}
}

func (ctrl *Controller) RegisterRoutes(rg *gin.RouterGroup) {
	booking := rg.Group("/bookings")
	{
		booking.POST("/", ctrl.CreateBooking)
		booking.DELETE("/cancel", ctrl.CancelBooking)
		booking.GET("/user/:user_id", ctrl.GetUserBookings)
		booking.GET("/seat-status", ctrl.CheckSeatStatus)
	}
}

func (ctrl *Controller) CreateBooking(c *gin.Context) {
	var req dto.BookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, constants.MsgInvalidRequest, nil)
		return
	}
	response, err, isSeatTaken := ctrl.service.BookSeat(req)
	if isSeatTaken {
		helpers.ErrorResponse(c, http.StatusConflict, constants.MsgSeatTaken, nil)
		return
	}
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, constants.MsgServerError, nil)
		return
	}
	helpers.SuccessResponse(c, http.StatusCreated, constants.MsgBookingSuccess, response)
}

func (ctrl *Controller) CancelBooking(c *gin.Context) {
	var req dto.CancelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, constants.MsgInvalidRequest, nil)
		return
	}

	response, err := ctrl.service.CancelBooking(req)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusNotFound, err.Error(), nil)
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, constants.MsgCancelSuccess, response)
}

func (ctrl *Controller) GetUserBookings(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}

	bookings, err := ctrl.service.GetUserBookings(userID)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError,
			constants.MsgServerError, nil)
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, "Bookings fetched", bookings)
}

func (ctrl *Controller) CheckSeatStatus(c *gin.Context) {
	trainID, _ := strconv.Atoi(c.Query("train_id"))
	seatID, _ := strconv.Atoi(c.Query("seat_id"))
	journeyDate := c.Query("journey_date")

	isLocked := ctrl.service.IsSeatLocked(trainID, seatID, journeyDate)
	isAvailable := ctrl.service.IsSeatAvailableForDate(seatID, journeyDate)

	helpers.SuccessResponse(c, http.StatusOK, "Seat status", gin.H{
		"seat_id":      seatID,
		"is_locked":    isLocked,
		"is_available": isAvailable,
	})
}
