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
	GetDepartureTime(trainID int) (string, error)
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
func (s *Service) BookSeat(req dto.BookingRequest) (*dto.BookingResponse, error, bool) {

	journeyDate, err := time.Parse("2006-01-02", req.JourneyDate)
	if err != nil || journeyDate.Before(time.Now().Truncate(24*time.Hour)) {
		return nil, fmt.Errorf("invalid or past journey date"), false
	}

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
// Guard order: validate → check departure → write to DB.
// Nothing is mutated before all checks pass.
func (s *Service) CancelBooking(req dto.CancelRequest) (*dto.CancelResponse, error) {
	// 1. Fetch the booking — read-only, no side effects.
	booking, err := s.repo.GetBookingByID(req.BookingID)
	if err != nil {
		return nil, fmt.Errorf(constants.MsgBookingNotFound)
	}

	// 2. Departure check before any writes.
	//    JourneyDate is stored as DATE (YYYY-MM-DD) by Postgres.
	departureTime, err := s.repo.GetDepartureTime(booking.TrainID)
	if err == nil {
		journeyDate, parseErr := time.Parse("2006-01-02", booking.JourneyDate)
		depTime, depErr := time.Parse("15:04:05", departureTime)
		if parseErr == nil && depErr == nil {
			journeyStart := time.Date(
				journeyDate.Year(), journeyDate.Month(), journeyDate.Day(),
				depTime.Hour(), depTime.Minute(), depTime.Second(), 0, time.Local,
			)
			if time.Now().After(journeyStart) {
				return nil, fmt.Errorf("cannot cancel after journey has started")
			}
		}
	}

	// 3. All guards passed — now write.
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