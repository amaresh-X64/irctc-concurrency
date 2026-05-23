package booking

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"gin-booking/internal/booking/dto"
	"gin-booking/internal/constants"
	redisClient "gin-booking/pkg/redis"

	"github.com/redis/go-redis/v9"
)

type RepositoryStore interface {
	GetSeatByID(seatID int) (*Seat, error)
	LockSeat(seatID int, tx *sql.Tx) error
	DecrementAvailableSeats(trainID int, tx *sql.Tx) error
	CreateBooking(userID, trainID, seatID int, journeyDate, status string, tx *sql.Tx) (int, error)
	GetBookingByID(bookingID int) (*Booking, error)
	CancelBooking(bookingID int) error
	UnlockSeat(seatID int) error
	GetBookingsByUser(userID int) ([]Booking, error)
	IsSeatAvailableForDate(seatID int, journeyDate string) (bool, error)
}

type Service struct {
	repo RepositoryStore
	db   *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{
		repo: NewRepository(db),
		db:   db,
	}
}

func NewServiceWithRepo(repo RepositoryStore, db *sql.DB) *Service {
	return &Service{
		repo: repo,
		db:   db,
	}
}

func (s *Service) BookSeat(req dto.BookingRequest) (*dto.BookingResponse, error, bool) {

	lockKey := fmt.Sprintf("%s%d:%d:%s", constants.SeatLockPrefix, req.TrainID, req.SeatID, req.JourneyDate)

	locked, err := redisClient.Client.SetNX(redisClient.Ctx, lockKey, req.UserID, time.Duration(constants.SeatLockTTL)*time.Second).Result()
	if err != nil {
		log.Printf(" Redis error: %v", err)
		return nil, fmt.Errorf("redis error"), false
	}
	if !locked {
		return nil, fmt.Errorf(constants.MsgSeatTaken), true
	}
	tx, err := s.db.Begin()

	if err != nil {
		redisClient.Client.Del(redisClient.Ctx, lockKey)
		return nil, err, false
	}
	available, err := s.repo.IsSeatAvailableForDate(req.SeatID, req.JourneyDate)
	if err != nil || !available {
		tx.Rollback()
		redisClient.Client.Del(redisClient.Ctx, lockKey)
		return nil, fmt.Errorf(constants.MsgSeatTaken), true
	}
	seat, err := s.repo.GetSeatByID(req.SeatID)
	if err != nil {
		tx.Rollback()
		redisClient.Client.Del(redisClient.Ctx, lockKey)
		return nil, fmt.Errorf(constants.MsgSeatTaken), true
	}

	if err := s.repo.LockSeat(req.SeatID, tx); err != nil {
		tx.Rollback()
		redisClient.Client.Del(redisClient.Ctx, lockKey)
		return nil, err, false
	}

	if err := s.repo.DecrementAvailableSeats(req.TrainID, tx); err != nil {
		tx.Rollback()
		redisClient.Client.Del(redisClient.Ctx, lockKey)
		return nil, err, false
	}

	bookingID, err := s.repo.CreateBooking(req.UserID, req.TrainID, req.SeatID, req.JourneyDate, constants.StatusConfirmed, tx)
	if err != nil {
		tx.Rollback()
		redisClient.Client.Del(redisClient.Ctx, lockKey)
		return nil, err, false
	}

	if err := tx.Commit(); err != nil {
		redisClient.Client.Del(redisClient.Ctx, lockKey)
		return nil, err, false
	}

	log.Printf("Seat %d booked by user %d — booking %d", req.SeatID, req.UserID, bookingID)

	return &dto.BookingResponse{BookingID: bookingID, UserID: req.UserID, TrainID: req.TrainID, SeatID: req.SeatID, SeatNumber: seat.SeatNumber,
		JourneyDate: req.JourneyDate, Status: constants.StatusConfirmed, BookedAt: time.Now()}, nil, false
}

func (s *Service) CancelBooking(req dto.CancelRequest) (*dto.CancelResponse, error) {
	booking, err := s.repo.GetBookingByID(req.BookingID)
	if err != nil {
		return nil, fmt.Errorf(constants.MsgBookingNotFound)
	}

	s.repo.UnlockSeat(booking.SeatID)
	s.repo.CancelBooking(req.BookingID)

	lockKey := fmt.Sprintf("%s%d:%d:%s", constants.SeatLockPrefix, booking.TrainID, booking.SeatID, booking.JourneyDate)
	redisClient.Client.Del(redisClient.Ctx, lockKey)

	go s.confirmNextWaitlist(booking.TrainID, booking.JourneyDate)

	return &dto.CancelResponse{
		BookingID: req.BookingID, Status: constants.StatusCancelled, Message: constants.MsgCancelSuccess,
	}, nil
}

func (s *Service) confirmNextWaitlist(trainID int, journeyDate string) {
	fastapiURL := os.Getenv("FASTAPI_URL")
	if fastapiURL == "" {
		fastapiURL = "http://fastapi-trains:8001"
	}
	url := fmt.Sprintf("%s/api/v1/waitlist/confirm-next?train_id=%d&journey_date=%s", fastapiURL, trainID, journeyDate)
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		log.Printf(" Waitlist confirm failed: %v", err)
		return
	}
	defer resp.Body.Close()
	log.Printf("Waitlist confirm-next called for train %d", trainID)
}

func (s *Service) GetUserBookings(userID int) ([]Booking, error) {
	return s.repo.GetBookingsByUser(userID)
}

func (s *Service) IsSeatLocked(trainID, seatID int, journeyDate string) bool {
	lockKey := fmt.Sprintf("%s%d:%d:%s", constants.SeatLockPrefix, trainID, seatID, journeyDate)
	val, err := redisClient.Client.Get(redisClient.Ctx, lockKey).Result()
	return err != redis.Nil && val != ""
}

func (s *Service) IsSeatAvailableForDate(seatID int, journeyDate string) bool {
	available, err := s.repo.IsSeatAvailableForDate(seatID, journeyDate)
	if err != nil {
		return false
	}
	return available
}
