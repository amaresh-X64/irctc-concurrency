package booking

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"gin-booking/internal/booking/dto"
	"gin-booking/internal/constants"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

// ─── No TestMain needed — zero external dependencies ──────────────────────────

// ─── Mock Repository ──────────────────────────────────────────────────────────

type MockRepository struct {
	getSeatByIDFn             func(seatID int) (*Seat, error)
	lockSeatFn                func(seatID int, tx *sql.Tx) error
	decrementAvailableSeatsFn func(trainID int, tx *sql.Tx) error
	createBookingFn           func(userID, trainID, seatID int, journeyDate, status string, tx *sql.Tx) (int, error)
	getBookingByIDFn          func(bookingID int) (*Booking, error)
	cancelBookingFn           func(bookingID int) error
	unlockSeatFn              func(seatID int) error
	getBookingsByUserFn       func(userID int) ([]Booking, error)
	isSeatAvailableForDateFn  func(seatID int, journeyDate string) (bool, error)
}

func (m *MockRepository) GetSeatByID(seatID int) (*Seat, error) {
	return m.getSeatByIDFn(seatID)
}
func (m *MockRepository) LockSeat(seatID int, tx *sql.Tx) error {
	return m.lockSeatFn(seatID, tx)
}
func (m *MockRepository) DecrementAvailableSeats(trainID int, tx *sql.Tx) error {
	return m.decrementAvailableSeatsFn(trainID, tx)
}
func (m *MockRepository) CreateBooking(userID, trainID, seatID int, journeyDate, status string, tx *sql.Tx) (int, error) {
	return m.createBookingFn(userID, trainID, seatID, journeyDate, status, tx)
}
func (m *MockRepository) GetBookingByID(bookingID int) (*Booking, error) {
	return m.getBookingByIDFn(bookingID)
}
func (m *MockRepository) CancelBooking(bookingID int) error {
	return m.cancelBookingFn(bookingID)
}
func (m *MockRepository) UnlockSeat(seatID int) error {
	return m.unlockSeatFn(seatID)
}
func (m *MockRepository) GetBookingsByUser(userID int) ([]Booking, error) {
	return m.getBookingsByUserFn(userID)
}
func (m *MockRepository) IsSeatAvailableForDate(seatID int, journeyDate string) (bool, error) {
	return m.isSeatAvailableForDateFn(seatID, journeyDate)
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func makeValidBookingRequest() dto.BookingRequest {
	return dto.BookingRequest{UserID: 1, TrainID: 1, SeatID: 1, JourneyDate: "2024-12-25"}
}

func makeMockSeat() *Seat {
	return &Seat{ID: 1, SeatNumber: "A1", IsAvailable: true}
}

func makeMockBooking() *Booking {
	return &Booking{
		ID: 1, UserID: 1, TrainID: 1, SeatID: 1,
		JourneyDate: "2024-12-25", Status: constants.StatusConfirmed, BookedAt: time.Now(),
	}
}

func newMockDB(t *testing.T, expectCommit bool) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	mock.ExpectBegin()
	if expectCommit {
		mock.ExpectCommit()
	} else {
		mock.ExpectRollback()
	}
	return db, mock
}

// successMockRepo returns a MockRepository wired for the full happy path.
func successMockRepo() *MockRepository {
	return &MockRepository{
		isSeatAvailableForDateFn: func(seatID int, journeyDate string) (bool, error) { return true, nil },
		getSeatByIDFn:            func(seatID int) (*Seat, error) { return makeMockSeat(), nil },
		lockSeatFn:               func(seatID int, tx *sql.Tx) error { return nil },
		decrementAvailableSeatsFn: func(trainID int, tx *sql.Tx) error { return nil },
		createBookingFn: func(userID, trainID, seatID int, journeyDate, status string, tx *sql.Tx) (int, error) {
			return 42, nil
		},
	}
}

// ─── GetUserBookings ──────────────────────────────────────────────────────────

func TestGetUserBookings_ReturnsBookings(t *testing.T) {
	mock := &MockRepository{
		getBookingsByUserFn: func(userID int) ([]Booking, error) {
			return []Booking{{ID: 1, UserID: userID}}, nil
		},
	}
	svc := NewServiceWithRepo(mock, nil)
	result, err := svc.GetUserBookings(1)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, 1, result[0].UserID)
}

func TestGetUserBookings_ReturnsEmptyList(t *testing.T) {
	mock := &MockRepository{
		getBookingsByUserFn: func(userID int) ([]Booking, error) {
			return []Booking{}, nil
		},
	}
	svc := NewServiceWithRepo(mock, nil)
	result, err := svc.GetUserBookings(1)
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestGetUserBookings_ReturnsError_WhenRepositoryFails(t *testing.T) {
	mock := &MockRepository{
		getBookingsByUserFn: func(userID int) ([]Booking, error) {
			return nil, errors.New("db error")
		},
	}
	svc := NewServiceWithRepo(mock, nil)
	result, err := svc.GetUserBookings(1)
	assert.Error(t, err)
	assert.Nil(t, result)
}

// ─── CancelBooking ────────────────────────────────────────────────────────────

func TestCancelBooking_ReturnsError_WhenBookingNotFound(t *testing.T) {
	mock := &MockRepository{
		getBookingByIDFn: func(bookingID int) (*Booking, error) {
			return nil, errors.New("not found")
		},
	}
	svc := NewServiceWithRepo(mock, nil)
	result, err := svc.CancelBooking(dto.CancelRequest{BookingID: 999})
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, constants.MsgBookingNotFound, err.Error())
}

func TestCancelBooking_ReturnsCancelledStatus(t *testing.T) {
	mock := &MockRepository{
		getBookingByIDFn: func(bookingID int) (*Booking, error) { return makeMockBooking(), nil },
		unlockSeatFn:     func(seatID int) error { return nil },
		cancelBookingFn:  func(bookingID int) error { return nil },
	}
	svc := NewServiceWithRepo(mock, nil)
	result, err := svc.CancelBooking(dto.CancelRequest{BookingID: 1})
	assert.NoError(t, err)
	assert.Equal(t, constants.StatusCancelled, result.Status)
	assert.Equal(t, 1, result.BookingID)
}

func TestCancelBooking_ReturnsError_WhenUnlockSeatFails(t *testing.T) {
	mock := &MockRepository{
		getBookingByIDFn: func(bookingID int) (*Booking, error) { return makeMockBooking(), nil },
		unlockSeatFn:     func(seatID int) error { return errors.New("db error") },
	}
	svc := NewServiceWithRepo(mock, nil)
	result, err := svc.CancelBooking(dto.CancelRequest{BookingID: 1})
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to unlock seat")
}

func TestCancelBooking_ReturnsError_WhenCancelBookingFails(t *testing.T) {
	mock := &MockRepository{
		getBookingByIDFn: func(bookingID int) (*Booking, error) { return makeMockBooking(), nil },
		unlockSeatFn:     func(seatID int) error { return nil },
		cancelBookingFn:  func(bookingID int) error { return errors.New("db error") },
	}
	svc := NewServiceWithRepo(mock, nil)
	result, err := svc.CancelBooking(dto.CancelRequest{BookingID: 1})
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to cancel booking")
}

// ─── IsSeatAvailableForDate ───────────────────────────────────────────────────

func TestIsSeatAvailableForDate_ReturnsTrue(t *testing.T) {
	mock := &MockRepository{
		isSeatAvailableForDateFn: func(seatID int, journeyDate string) (bool, error) { return true, nil },
	}
	svc := NewServiceWithRepo(mock, nil)
	assert.True(t, svc.IsSeatAvailableForDate(1, "2024-12-25"))
}

func TestIsSeatAvailableForDate_ReturnsFalse_WhenBooked(t *testing.T) {
	mock := &MockRepository{
		isSeatAvailableForDateFn: func(seatID int, journeyDate string) (bool, error) { return false, nil },
	}
	svc := NewServiceWithRepo(mock, nil)
	assert.False(t, svc.IsSeatAvailableForDate(1, "2024-12-25"))
}

func TestIsSeatAvailableForDate_ReturnsFalse_WhenRepoErrors(t *testing.T) {
	mock := &MockRepository{
		isSeatAvailableForDateFn: func(seatID int, journeyDate string) (bool, error) {
			return false, errors.New("db error")
		},
	}
	svc := NewServiceWithRepo(mock, nil)
	assert.False(t, svc.IsSeatAvailableForDate(1, "2024-12-25"))
}

// ─── IsSeatLocked — now purely in-process, no Redis ──────────────────────────

func TestIsSeatLocked_ReturnsFalse_WhenNotLocked(t *testing.T) {
	svc := NewServiceWithRepo(&MockRepository{}, nil)
	assert.False(t, svc.IsSeatLocked(1, 1, "2024-12-25"))
}

func TestIsSeatLocked_ReturnsTrue_WhenLocked(t *testing.T) {
	svc := NewServiceWithRepo(&MockRepository{}, nil)
	key := "seat_lock:1:10:2024-12-25"
	svc.locker.TryLock(key)
	defer svc.locker.Unlock(key)
	assert.True(t, svc.IsSeatLocked(1, 10, "2024-12-25"))
}

// ─── BookSeat ─────────────────────────────────────────────────────────────────

func TestBookSeat_ReturnsConflict_WhenSeatAlreadyLocked(t *testing.T) {
	svc := NewServiceWithRepo(&MockRepository{}, nil)
	req := makeValidBookingRequest()

	// pre-acquire the lock to simulate a concurrent holder
	lockKey := "seat_lock:1:1:2024-12-25"
	svc.locker.TryLock(lockKey)
	defer svc.locker.Unlock(lockKey)

	result, err, isSeatTaken := svc.BookSeat(req)
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.True(t, isSeatTaken)
}

func TestBookSeat_ReturnsError_WhenDBBeginFails(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	db.Close() // closed DB forces Begin() to fail

	mock := &MockRepository{
		isSeatAvailableForDateFn: func(seatID int, journeyDate string) (bool, error) { return true, nil },
	}
	svc := NewServiceWithRepo(mock, db)

	result, err, isSeatTaken := svc.BookSeat(dto.BookingRequest{UserID: 1, TrainID: 1, SeatID: 5, JourneyDate: "2024-12-25"})
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.False(t, isSeatTaken)
}

func TestBookSeat_ReturnsConflict_WhenSeatNotAvailableForDate(t *testing.T) {
	db, sqlMock, _ := sqlmock.New()
	defer db.Close()
	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	mock := &MockRepository{
		isSeatAvailableForDateFn: func(seatID int, journeyDate string) (bool, error) { return false, nil },
	}
	svc := NewServiceWithRepo(mock, db)

	result, err, isSeatTaken := svc.BookSeat(dto.BookingRequest{UserID: 1, TrainID: 1, SeatID: 2, JourneyDate: "2024-12-25"})
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.True(t, isSeatTaken)
}

func TestBookSeat_ReturnsConflict_WhenGetSeatFails(t *testing.T) {
	db, sqlMock, _ := sqlmock.New()
	defer db.Close()
	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	mock := &MockRepository{
		isSeatAvailableForDateFn: func(seatID int, journeyDate string) (bool, error) { return true, nil },
		getSeatByIDFn:            func(seatID int) (*Seat, error) { return nil, errors.New("seat not found") },
	}
	svc := NewServiceWithRepo(mock, db)

	result, err, isSeatTaken := svc.BookSeat(dto.BookingRequest{UserID: 1, TrainID: 1, SeatID: 3, JourneyDate: "2024-12-25"})
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.True(t, isSeatTaken)
}

func TestBookSeat_ReturnsError_WhenLockSeatFails(t *testing.T) {
	db, sqlMock, _ := sqlmock.New()
	defer db.Close()
	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	mock := &MockRepository{
		isSeatAvailableForDateFn: func(seatID int, journeyDate string) (bool, error) { return true, nil },
		getSeatByIDFn:            func(seatID int) (*Seat, error) { return makeMockSeat(), nil },
		lockSeatFn:               func(seatID int, tx *sql.Tx) error { return errors.New("lock failed") },
	}
	svc := NewServiceWithRepo(mock, db)

	result, err, isSeatTaken := svc.BookSeat(dto.BookingRequest{UserID: 1, TrainID: 1, SeatID: 6, JourneyDate: "2024-12-25"})
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.False(t, isSeatTaken)
}

func TestBookSeat_ReturnsError_WhenDecrementFails(t *testing.T) {
	db, sqlMock, _ := sqlmock.New()
	defer db.Close()
	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	mock := &MockRepository{
		isSeatAvailableForDateFn:  func(seatID int, journeyDate string) (bool, error) { return true, nil },
		getSeatByIDFn:             func(seatID int) (*Seat, error) { return makeMockSeat(), nil },
		lockSeatFn:                func(seatID int, tx *sql.Tx) error { return nil },
		decrementAvailableSeatsFn: func(trainID int, tx *sql.Tx) error { return errors.New("decrement failed") },
	}
	svc := NewServiceWithRepo(mock, db)

	result, err, isSeatTaken := svc.BookSeat(dto.BookingRequest{UserID: 1, TrainID: 1, SeatID: 7, JourneyDate: "2024-12-25"})
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.False(t, isSeatTaken)
}

func TestBookSeat_ReturnsError_WhenCreateBookingFails(t *testing.T) {
	db, sqlMock, _ := sqlmock.New()
	defer db.Close()
	sqlMock.ExpectBegin()
	sqlMock.ExpectRollback()

	mock := &MockRepository{
		isSeatAvailableForDateFn:  func(seatID int, journeyDate string) (bool, error) { return true, nil },
		getSeatByIDFn:             func(seatID int) (*Seat, error) { return makeMockSeat(), nil },
		lockSeatFn:                func(seatID int, tx *sql.Tx) error { return nil },
		decrementAvailableSeatsFn: func(trainID int, tx *sql.Tx) error { return nil },
		createBookingFn: func(userID, trainID, seatID int, journeyDate, status string, tx *sql.Tx) (int, error) {
			return 0, errors.New("insert failed")
		},
	}
	svc := NewServiceWithRepo(mock, db)

	result, err, isSeatTaken := svc.BookSeat(dto.BookingRequest{UserID: 1, TrainID: 1, SeatID: 8, JourneyDate: "2024-12-25"})
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.False(t, isSeatTaken)
}

func TestBookSeat_ReturnsSuccess_WhenAllStepsPass(t *testing.T) {
	db, sqlMock, _ := sqlmock.New()
	defer db.Close()
	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	svc := NewServiceWithRepo(successMockRepo(), db)

	result, err, isSeatTaken := svc.BookSeat(dto.BookingRequest{UserID: 1, TrainID: 1, SeatID: 4, JourneyDate: "2024-12-25"})
	assert.NoError(t, err)
	assert.False(t, isSeatTaken)
	assert.NotNil(t, result)
	assert.Equal(t, 42, result.BookingID)
	assert.Equal(t, constants.StatusConfirmed, result.Status)
	assert.Equal(t, "A1", result.SeatNumber)
}

func TestBookSeat_ReturnsError_WhenCommitFails(t *testing.T) {
	db, sqlMock, _ := sqlmock.New()
	defer db.Close()
	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit().WillReturnError(errors.New("commit failed"))

	svc := NewServiceWithRepo(successMockRepo(), db)

	result, err, isSeatTaken := svc.BookSeat(dto.BookingRequest{UserID: 1, TrainID: 1, SeatID: 9, JourneyDate: "2024-12-25"})
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.False(t, isSeatTaken)
}

func TestNewService_ReturnsServiceWithRepo(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()
	svc := NewService(db)
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.repo)
	assert.NotNil(t, svc.locker)
	assert.Equal(t, db, svc.db)
}

// ─── Concurrency test — the one that was impossible with Redis ────────────────
// 20 goroutines race to book the exact same seat on the same date.
// Exactly 1 must succeed; the other 19 must get a seat-taken conflict.
// No external services, no Docker, pure Go.
func TestBookSeat_Concurrent_OnlyOneSucceeds(t *testing.T) {
	const goroutines = 20

	// Each goroutine gets its own sqlmock db because sql.Tx cannot be shared.
	// We model the service with a shared MockRepository and shared SeatLocker.
	// The locker is the only shared state under test — which is the point.
	successCount := 0
	conflictCount := 0
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Shared service — one locker, one mock repo that always says "available".
	// In production the DB transaction provides the second gate; here we trust
	// the locker alone since that is the unit under test.
	repo := &MockRepository{
		isSeatAvailableForDateFn: func(seatID int, journeyDate string) (bool, error) { return true, nil },
		getSeatByIDFn:            func(seatID int) (*Seat, error) { return makeMockSeat(), nil },
		lockSeatFn:               func(seatID int, tx *sql.Tx) error { return nil },
		decrementAvailableSeatsFn: func(trainID int, tx *sql.Tx) error { return nil },
		createBookingFn: func(userID, trainID, seatID int, journeyDate, status string, tx *sql.Tx) (int, error) {
			return 1, nil
		},
	}
	svc := NewServiceWithRepo(repo, nil)

	results := make([]struct {
		resp       *dto.BookingResponse
		err        error
		isTaken    bool
	}, goroutines)

	// We need a real DB for Begin() — one per goroutine.
	dbs := make([]*sql.DB, goroutines)
	for i := range dbs {
		db, mock, _ := sqlmock.New()
		mock.ExpectBegin()
		mock.ExpectCommit()
		dbs[i] = db
	}
	defer func() {
		for _, db := range dbs {
			db.Close()
		}
	}()

	// Override the service's db per-goroutine by calling a local helper
	// that injects the per-goroutine db only for Begin().
	// Because SeatLocker is shared on svc, only the first goroutine to
	// TryLock will proceed; the rest return isSeatTaken=true immediately
	// without ever calling db.Begin().
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		i := i
		go func() {
			defer wg.Done()
			// Build a local service that shares the same locker via svc.locker
			// but has its own db.
			local := &Service{repo: repo, db: dbs[i], locker: svc.locker}
			req := dto.BookingRequest{UserID: i + 1, TrainID: 1, SeatID: 1, JourneyDate: "2024-12-25"}
			resp, err, isTaken := local.BookSeat(req)
			mu.Lock()
			results[i] = struct {
				resp    *dto.BookingResponse
				err     error
				isTaken bool
			}{resp, err, isTaken}
			mu.Unlock()
		}()
	}
	wg.Wait()

	for _, r := range results {
		if r.err == nil && r.resp != nil {
			successCount++
		} else if r.isTaken {
			conflictCount++
		}
	}

	assert.Equal(t, 1, successCount, "exactly one booking should succeed")
	assert.Equal(t, goroutines-1, conflictCount, "all others should get seat-taken conflict")
}

// ─── confirmNextWaitlist ──────────────────────────────────────────────────────

func TestConfirmNextWaitlist_SuccessfulHTTPCall(t *testing.T) {
	called := make(chan struct{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called <- struct{}{}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	t.Setenv("FASTAPI_URL", server.URL)

	svc := NewServiceWithRepo(&MockRepository{}, nil)
	svc.confirmNextWaitlist(1, "2024-12-25")

	select {
	case <-called:
	case <-time.After(2 * time.Second):
		t.Fatal("waitlist endpoint was not called")
	}
}

func TestConfirmNextWaitlist_HandlesHTTPError(t *testing.T) {
	t.Setenv("FASTAPI_URL", "http://localhost:1")
	svc := NewServiceWithRepo(&MockRepository{}, nil)
	assert.NotPanics(t, func() { svc.confirmNextWaitlist(1, "2024-12-25") })
}

func TestConfirmNextWaitlist_UsesFallbackURL_WhenEnvNotSet(t *testing.T) {
	os.Unsetenv("FASTAPI_URL")
	svc := NewServiceWithRepo(&MockRepository{}, nil)
	assert.NotPanics(t, func() { svc.confirmNextWaitlist(1, "2024-12-25") })
}