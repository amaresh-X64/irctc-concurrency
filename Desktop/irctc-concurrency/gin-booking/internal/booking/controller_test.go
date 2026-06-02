package booking

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gin-booking/internal/booking/dto"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockBookingService struct {
	bookSeatFn             func(req dto.BookingRequest) (*dto.BookingResponse, error, bool)
	cancelBookingFn        func(req dto.CancelRequest) (*dto.CancelResponse, error)
	getUserBookingsFn      func(userID int) ([]Booking, error)
	isSeatLockedFn         func(trainID, seatID int, journeyDate string) bool
	isSeatAvailableForDate func(seatID int, journeyDate string) bool
}

func (m *mockBookingService) BookSeat(req dto.BookingRequest) (*dto.BookingResponse, error, bool) {
	return m.bookSeatFn(req)
}
func (m *mockBookingService) CancelBooking(req dto.CancelRequest) (*dto.CancelResponse, error) {
	return m.cancelBookingFn(req)
}
func (m *mockBookingService) GetUserBookings(userID int) ([]Booking, error) {
	return m.getUserBookingsFn(userID)
}
func (m *mockBookingService) IsSeatLocked(trainID, seatID int, journeyDate string) bool {
	return m.isSeatLockedFn(trainID, seatID, journeyDate)
}
func (m *mockBookingService) IsSeatAvailableForDate(seatID int, journeyDate string) bool {
	return m.isSeatAvailableForDate(seatID, journeyDate)
}

func setupRouter(svc BookingService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	ctrl := NewController(svc)
	ctrl.RegisterRoutes(r.Group("/api/v1"))
	return r
}

func toJSON(v any) *bytes.Buffer {
	b, _ := json.Marshal(v)
	return bytes.NewBuffer(b)
}

func TestCreateBooking_Returns201_WhenSuccessful(t *testing.T) {
	svc := &mockBookingService{
		bookSeatFn: func(req dto.BookingRequest) (*dto.BookingResponse, error, bool) {
			return &dto.BookingResponse{BookingID: 1, SeatNumber: "A1"}, nil, false
		},
	}
	r := setupRouter(svc)

	body := toJSON(dto.BookingRequest{UserID: 1, TrainID: 1, SeatID: 1, JourneyDate: "2028-12-25"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/bookings/", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCreateBooking_Returns409_WhenSeatTaken(t *testing.T) {
	svc := &mockBookingService{
		bookSeatFn: func(req dto.BookingRequest) (*dto.BookingResponse, error, bool) {
			return nil, errors.New("seat taken"), true
		},
	}
	r := setupRouter(svc)

	body := toJSON(dto.BookingRequest{UserID: 1, TrainID: 1, SeatID: 1, JourneyDate: "2028-12-25"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/bookings/", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestCreateBooking_Returns500_WhenServiceErrors(t *testing.T) {
	svc := &mockBookingService{
		bookSeatFn: func(req dto.BookingRequest) (*dto.BookingResponse, error, bool) {
			return nil, errors.New("db error"), false
		},
	}
	r := setupRouter(svc)

	body := toJSON(dto.BookingRequest{UserID: 1, TrainID: 1, SeatID: 1, JourneyDate: "2028-12-25"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/bookings/", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCreateBooking_Returns400_WhenBadJSON(t *testing.T) {
	svc := &mockBookingService{}
	r := setupRouter(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/bookings/", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCancelBooking_Returns200_WhenSuccessful(t *testing.T) {
	svc := &mockBookingService{
		cancelBookingFn: func(req dto.CancelRequest) (*dto.CancelResponse, error) {
			return &dto.CancelResponse{BookingID: 1, Status: "CANCELLED"}, nil
		},
	}
	r := setupRouter(svc)

	body := toJSON(dto.CancelRequest{BookingID: 1, UserID: 1})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/bookings/cancel", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCancelBooking_Returns404_WhenBookingNotFound(t *testing.T) {
	svc := &mockBookingService{
		cancelBookingFn: func(req dto.CancelRequest) (*dto.CancelResponse, error) {
			return nil, errors.New("booking not found")
		},
	}
	r := setupRouter(svc)

	body := toJSON(dto.CancelRequest{BookingID: 99, UserID: 1})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/bookings/cancel", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCancelBooking_Returns400_WhenBadJSON(t *testing.T) {
	svc := &mockBookingService{}
	r := setupRouter(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/bookings/cancel", bytes.NewBufferString("bad"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetUserBookings_Returns200_WithBookings(t *testing.T) {
	svc := &mockBookingService{
		getUserBookingsFn: func(userID int) ([]Booking, error) {
			return []Booking{{ID: 1, UserID: userID}}, nil
		},
	}
	r := setupRouter(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/bookings/user/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetUserBookings_Returns500_WhenServiceErrors(t *testing.T) {
	svc := &mockBookingService{
		getUserBookingsFn: func(userID int) ([]Booking, error) {
			return nil, errors.New("db error")
		},
	}
	r := setupRouter(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/bookings/user/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetUserBookings_Returns400_WhenUserIDInvalid(t *testing.T) {
	svc := &mockBookingService{}
	r := setupRouter(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/bookings/user/abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCheckSeatStatus_Returns200_WithLockedAndAvailableStatus(t *testing.T) {
	svc := &mockBookingService{
		isSeatLockedFn: func(trainID, seatID int, journeyDate string) bool {
			return true
		},
		isSeatAvailableForDate: func(seatID int, journeyDate string) bool {
			return false
		},
	}
	r := setupRouter(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet,
		"/api/v1/bookings/seat-status?train_id=1&seat_id=2&journey_date=2028-12-25", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]any)
	assert.Equal(t, true, data["is_locked"])
	assert.Equal(t, false, data["is_available"])
}

func TestCheckSeatStatus_Returns200_WhenSeatFree(t *testing.T) {
	svc := &mockBookingService{
		isSeatLockedFn:         func(trainID, seatID int, journeyDate string) bool { return false },
		isSeatAvailableForDate: func(seatID int, journeyDate string) bool { return true },
	}
	r := setupRouter(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet,
		"/api/v1/bookings/seat-status?train_id=1&seat_id=3&journey_date=2028-12-25", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
