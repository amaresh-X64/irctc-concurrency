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
	repo   RepositoryStore
	db     *sql.DB
	locker *SeatLocker
}

func NewService(db *sql.DB) *Service {
	return &Service{
		repo:   NewRepository(db),
		db:     db,
		locker: NewSeatLocker(),
	}
}

func NewServiceWithRepo(repo RepositoryStore, db *sql.DB) *Service {
	return &Service{
		repo:   repo,
		db:     db,
		locker: NewSeatLocker(),
	}
}

// ─── BookSeat ─────────────────────────────────────────────────────────────────
// Concurrency strategy:
//   1. In-process SeatLocker (sync.Map + sync.Mutex) — first gate, no I/O cost.
//   2. DB transaction with IsSeatAvailableForDate — second gate, serialised by
//      Postgres row locking, so concurrent requests from different pods are also
//      safe once a proper SELECT FOR UPDATE is in place.
func (s *Service) BookSeat(req dto.BookingRequest) (*dto.BookingResponse, error, bool) {
	lockKey := fmt.Sprintf("%s%d:%d:%s", constants.SeatLockPrefix, req.TrainID, req.SeatID, req.JourneyDate)

	if !s.locker.TryLock(lockKey) {
		return nil, fmt.Errorf(constants.MsgSeatTaken), true
	}
	defer s.locker.Unlock(lockKey)

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err, false
	}

	available, err := s.repo.IsSeatAvailableForDate(req.SeatID, req.JourneyDate)
	if err != nil || !available {
		tx.Rollback()
		return nil, fmt.Errorf(constants.MsgSeatTaken), true
	}

	seat, err := s.repo.GetSeatByID(req.SeatID)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf(constants.MsgSeatTaken), true
	}

	if err := s.repo.LockSeat(req.SeatID, tx); err != nil {
		tx.Rollback()
		return nil, err, false
	}

	if err := s.repo.DecrementAvailableSeats(req.TrainID, tx); err != nil {
		tx.Rollback()
		return nil, err, false
	}

	bookingID, err := s.repo.CreateBooking(req.UserID, req.TrainID, req.SeatID, req.JourneyDate, constants.StatusConfirmed, tx)
	if err != nil {
		tx.Rollback()
		return nil, err, false
	}

	if err := tx.Commit(); err != nil {
		return nil, err, false
	}

	log.Printf("Seat %d booked by user %d — booking %d", req.SeatID, req.UserID, bookingID)

	return &dto.BookingResponse{
		BookingID:   bookingID,
		UserID:      req.UserID,
		TrainID:     req.TrainID,
		SeatID:      req.SeatID,
		SeatNumber:  seat.SeatNumber,
		JourneyDate: req.JourneyDate,
		Status:      constants.StatusConfirmed,
		BookedAt:    time.Now(),
	}, nil, false
}

// ─── CancelBooking ────────────────────────────────────────────────────────────
// Errors from UnlockSeat and CancelBooking are now propagated — previously they
// were silently swallowed, which could leave the seat in a permanently broken state.
func (s *Service) CancelBooking(req dto.CancelRequest) (*dto.CancelResponse, error) {
	booking, err := s.repo.GetBookingByID(req.BookingID)
	if err != nil {
		return nil, fmt.Errorf(constants.MsgBookingNotFound)
	}

	if err := s.repo.UnlockSeat(booking.SeatID); err != nil {
		return nil, fmt.Errorf("failed to unlock seat: %w", err)
	}

	if err := s.repo.CancelBooking(req.BookingID); err != nil {
		return nil, fmt.Errorf("failed to cancel booking: %w", err)
	}

	go s.confirmNextWaitlist(booking.TrainID, booking.JourneyDate)

	return &dto.CancelResponse{
		BookingID: req.BookingID,
		Status:    constants.StatusCancelled,
		Message:   constants.MsgCancelSuccess,
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
		log.Printf("Waitlist confirm failed: %v", err)
		return
	}
	defer resp.Body.Close()
	log.Printf("Waitlist confirm-next called for train %d", trainID)
}

func (s *Service) GetUserBookings(userID int) ([]Booking, error) {
	return s.repo.GetBookingsByUser(userID)
}

// IsSeatLocked checks the in-process locker — replaces the Redis GET.
func (s *Service) IsSeatLocked(trainID, seatID int, journeyDate string) bool {
	lockKey := fmt.Sprintf("%s%d:%d:%s", constants.SeatLockPrefix, trainID, seatID, journeyDate)
	return s.locker.IsLocked(lockKey)
}

func (s *Service) IsSeatAvailableForDate(seatID int, journeyDate string) bool {
	available, err := s.repo.IsSeatAvailableForDate(seatID, journeyDate)
	if err != nil {
		return false
	}
	return available
}