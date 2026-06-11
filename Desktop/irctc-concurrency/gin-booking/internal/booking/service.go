package booking

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"gin-booking/internal/booking/dto"
	"gin-booking/internal/constants"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// tracer is the package-level tracer for the booking domain.
// Name matches the service so spans group correctly in ES.
var tracer = otel.Tracer("gin-booking/booking")

type RepositoryStore interface {
	GetSeatByID(seatID int) (*Seat, error)
	LockSeat(seatID int, tx *sql.Tx) error
	DecrementAvailableSeats(trainID int, tx *sql.Tx) error
	IncrementAvailableSeats(trainID int, tx *sql.Tx) error
	CreateBooking(userID, trainID, seatID int, journeyDate, status string, tx *sql.Tx) (int, error)
	GetBookingByID(bookingID int) (*Booking, error)
	CancelBooking(bookingID int) error
	UnlockSeat(seatID int) error
	UnlockSeatTx(seatID int, tx *sql.Tx) error
	DeleteBooking(bookingID int, tx *sql.Tx) error
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

func (s *Service) BookSeat(req dto.BookingRequest) (*dto.BookingResponse, error, bool) {
	// ── Span: BookSeat ──────────────────────────────────────────────────────
	// This is the most important span in the whole system.
	// It captures whether the in-memory locker fired, DB contention,
	// and final booking outcome. The anomaly detector learns from these.
	ctx, span := tracer.Start(context.Background(), "BookSeat")
	defer span.End()

	span.SetAttributes(
		attribute.Int("booking.train_id", req.TrainID),
		attribute.Int("booking.seat_id", req.SeatID),
		attribute.String("booking.journey_date", req.JourneyDate),
		attribute.Int("booking.user_id", req.UserID),
	)

	lockKey := fmt.Sprintf("%s%d:%d:%s", constants.SeatLockPrefix, req.TrainID, req.SeatID, req.JourneyDate)

	// ── Hybrid mapping demo ──────────────────────────────────────────────
	// meta.* attributes land in Attributes.meta, mapped as `flattened` —
	// arbitrary key-value pairs queryable via exists/term but not given
	// individual field mappings. lock_key is high-cardinality and would
	// otherwise risk mapping explosion if explicitly typed.
	span.SetAttributes(
		attribute.String("meta.lock_key", lockKey),
		attribute.String("meta.client_request_id", fmt.Sprintf("req-%d-%d", req.UserID, time.Now().UnixNano())),
	)

	journeyDate, err := time.Parse("2006-01-02", req.JourneyDate)
	if err != nil || journeyDate.Before(time.Now().Truncate(24*time.Hour)) {
		span.SetStatus(codes.Error, "invalid_journey_date")
		span.SetAttributes(attribute.String("booking.outcome", "invalid_date"))
		return nil, fmt.Errorf("invalid or past journey date"), false
	}

	// ── Sub-span: in-memory lock attempt ───────────────────────────────────
	// This is what makes the concurrency story visible in traces.
	// A surge of locker_acquired=false in significant_terms = contention spike.
	_, lockSpan := tracer.Start(ctx, "SeatLocker.TryLock")
	acquired := s.locker.TryLock(lockKey)
	lockSpan.SetAttributes(
		attribute.String("locker.key", lockKey),
		attribute.Bool("locker.acquired", acquired),
	)
	lockSpan.End()

	if !acquired {
		span.SetAttributes(
			attribute.String("booking.outcome", "lock_contention"),
			attribute.Bool("booking.locker_hit", false),
			// Mirrored onto the parent span so percolator rules can match
			// on the full booking context in one document.
			attribute.Bool("locker.acquired", false),
		)
		span.SetStatus(codes.Error, "seat_lock_contention")
		return nil, fmt.Errorf(constants.MsgSeatTaken), true
	}
	defer s.locker.Unlock(lockKey)

	span.SetAttributes(
		attribute.Bool("booking.locker_hit", true),
		attribute.Bool("locker.acquired", true),
	)

	// ── Sub-span: DB transaction ───────────────────────────────────────────
	_, txSpan := tracer.Start(ctx, "BookSeat.DBTransaction")

	tx, err := s.db.Begin()
	if err != nil {
		txSpan.RecordError(err)
		txSpan.End()
		span.SetStatus(codes.Error, "db_begin_failed")
		return nil, err, false
	}

	available, err := s.repo.IsSeatAvailableForDate(req.SeatID, req.JourneyDate)
	if err != nil || !available {
		tx.Rollback()
		txSpan.SetAttributes(attribute.Bool("db.seat_available", false))
		txSpan.End()
		span.SetAttributes(
			attribute.String("booking.outcome", "seat_unavailable"),
			// Mirrored onto the parent span so percolator rules can match
			// on the full booking context in one document.
			attribute.Bool("db.seat_available", false),
		)
		span.SetStatus(codes.Error, "seat_unavailable")
		return nil, fmt.Errorf(constants.MsgSeatTaken), true
	}

	txSpan.SetAttributes(attribute.Bool("db.seat_available", true))
	span.SetAttributes(attribute.Bool("db.seat_available", true))

	seat, err := s.repo.GetSeatByID(req.SeatID)
	if err != nil {
		tx.Rollback()
		txSpan.RecordError(err)
		txSpan.End()
		span.SetStatus(codes.Error, "seat_fetch_failed")
		return nil, fmt.Errorf(constants.MsgSeatTaken), true
	}

	if err := s.repo.LockSeat(req.SeatID, tx); err != nil {
		tx.Rollback()
		txSpan.RecordError(err)
		txSpan.End()
		return nil, err, false
	}

	if err := s.repo.DecrementAvailableSeats(req.TrainID, tx); err != nil {
		tx.Rollback()
		txSpan.RecordError(err)
		txSpan.End()
		return nil, err, false
	}

	bookingID, err := s.repo.CreateBooking(req.UserID, req.TrainID, req.SeatID, req.JourneyDate, constants.StatusConfirmed, tx)
	if err != nil {
		tx.Rollback()
		txSpan.RecordError(err)
		txSpan.End()
		return nil, err, false
	}

	if err := tx.Commit(); err != nil {
		txSpan.RecordError(err)
		txSpan.End()
		return nil, err, false
	}

	txSpan.SetAttributes(
		attribute.Int("db.booking_id", bookingID),
		attribute.String("db.status", constants.StatusConfirmed),
	)
	txSpan.End()

	// ── Final span attributes: the anomaly detector keys on these ──────────
	span.SetAttributes(
		attribute.String("booking.outcome", "confirmed"),
		attribute.String("booking.status", constants.StatusConfirmed),
		attribute.Int("booking.booking_id", bookingID),
		attribute.String("booking.seat_number", seat.SeatNumber),
		attribute.String("booking.seat_type", seat.SeatType),
	)
	span.SetStatus(codes.Ok, "")

	log.Printf("Seat %d booked by user %d — booking %d", req.SeatID, req.UserID, bookingID)
	go s.ScheduleExpiryCheck(bookingID, req.TrainID, req.SeatID)
	go s.syncSeatsToES(req.TrainID)

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

func (s *Service) CancelBooking(req dto.CancelRequest) (*dto.CancelResponse, error) {
	_, span := tracer.Start(context.Background(), "CancelBooking")
	defer span.End()

	span.SetAttributes(attribute.Int("booking.booking_id", req.BookingID))

	booking, err := s.repo.GetBookingByID(req.BookingID)
	if err != nil {
		span.SetStatus(codes.Error, "booking_not_found")
		return nil, fmt.Errorf(constants.MsgBookingNotFound)
	}

	span.SetAttributes(
		attribute.Int("booking.train_id", booking.TrainID),
		attribute.Int("booking.seat_id", booking.SeatID),
		attribute.String("booking.journey_date", booking.JourneyDate),
	)

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
				span.SetAttributes(attribute.String("cancel.outcome", "post_departure_rejected"))
				span.SetStatus(codes.Error, "post_departure")
				return nil, fmt.Errorf("cannot cancel after journey has started")
			}
		}
	}

	if err := s.repo.UnlockSeat(booking.SeatID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "unlock_failed")
		return nil, fmt.Errorf("failed to unlock seat: %w", err)
	}

	if err := s.repo.CancelBooking(req.BookingID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "cancel_db_failed")
		return nil, fmt.Errorf("failed to cancel booking: %w", err)
	}

	span.SetAttributes(attribute.String("cancel.outcome", "cancelled"))
	span.SetStatus(codes.Ok, "")

	go s.confirmNextWaitlist(booking.TrainID, booking.JourneyDate)
	go s.syncSeatsToES(booking.TrainID)

	return &dto.CancelResponse{
		BookingID: req.BookingID,
		Status:    constants.StatusCancelled,
		Message:   constants.MsgCancelSuccess,
	}, nil
}

func (s *Service) syncSeatsToES(trainID int) {
	fastapiURL := os.Getenv("FASTAPI_URL")
	if fastapiURL == "" {
		fastapiURL = "http://fastapi-trains:8001"
	}
	repo, ok := s.repo.(*Repository)
	if !ok {
		return
	}
	available, err := repo.GetAvailableSeats(trainID)
	if err != nil {
		log.Printf("syncSeatsToES: could not read available_seats for train %d: %v", trainID, err)
		return
	}

	url := fmt.Sprintf("%s/internal/trains/%d/seats", fastapiURL, trainID)
	body := fmt.Sprintf(`{"available_seats":%d}`, available)
	req, err := http.NewRequest(http.MethodPatch, url, strings.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("syncSeatsToES: PATCH failed for train %d: %v", trainID, err)
		return
	}
	defer resp.Body.Close()
	log.Printf("syncSeatsToES: train %d ES seat count updated (available=%d)", trainID, available)
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