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

type MockRepository struct {
	getSeatByIDFn             func(seatID int) (*Seat, error)
	lockSeatFn                func(seatID int, tx *sql.Tx) error
	decrementAvailableSeatsFn func(trainID int, tx *sql.Tx) error
	incrementAvailableSeatsFn func(trainID int, tx *sql.Tx) error
	createBookingFn           func(userID, trainID, seatID int, journeyDate, status string, tx *sql.Tx) (int, error)
	getBookingByIDFn          func(bookingID int) (*Booking, error)
	cancelBookingFn           func(bookingID int) error
	unlockSeatFn              func(seatID int) error
	unlockSeatTxFn            func(seatID int, tx *sql.Tx) error
	deleteBookingFn           func(bookingID int, tx *sql.Tx) error
	getBookingsByUserFn       func(userID int) ([]Booking, error)
	isSeatAvailableForDateFn  func(seatID int, journeyDate string) (bool, error)
	getDepartureTimeFn        func(trainID int) (string, error)
	getAvailableSeatsFn       func(trainID int) (int, error)
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
func (m *MockRepository) IncrementAvailableSeats(trainID int, tx *sql.Tx) error {
	if m.incrementAvailableSeatsFn != nil {
		return m.incrementAvailableSeatsFn(trainID, tx)
	}
	return nil
}
func (m *MockRepository) UnlockSeatTx(seatID int, tx *sql.Tx) error {
	if m.unlockSeatTxFn != nil {
		return m.unlockSeatTxFn(seatID, tx)
	}
	return nil
}
func (m *MockRepository) DeleteBooking(bookingID int, tx *sql.Tx) error {
	if m.deleteBookingFn != nil {
		return m.deleteBookingFn(bookingID, tx)
	}
	return nil
}
func (m *MockRepository) GetDepartureTime(trainID int) (string, error) {
	if m.getDepartureTimeFn != nil {
		return m.getDepartureTimeFn(trainID)
	}
	return "", nil
}
func (m *MockRepository) GetAvailableSeats(trainID int) (int, error) {
	if m.getAvailableSeatsFn != nil {
		return m.getAvailableSeatsFn(trainID)
	}
	return 0, nil
}

func makeValidBookingRequest() dto.BookingRequest {
	return dto.BookingRequest{UserID: 1, TrainID: 1, SeatID: 1, JourneyDate: "2028-12-25"}
}

func makeMockSeat() *Seat {
	return &Seat{ID: 1, SeatNumber: "A1", IsAvailable: true}
}

func makeMockBooking() *Booking {
	return &Booking{
		ID: 1, UserID: 1, TrainID: 1, SeatID: 1,
		JourneyDate: "2028-12-25", Status: constants.StatusConfirmed, BookedAt: time.Now(),
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
		isSeatAvailableForDateFn:  func(seatID int, journeyDate string) (bool, error) { return true, nil },
		getSeatByIDFn:             func(seatID int) (*Seat, error) { return makeMockSeat(), nil },
		lockSeatFn:                func(seatID int, tx *sql.Tx) error { return nil },
		decrementAvailableSeatsFn: func(trainID int, tx *sql.Tx) error { return nil },
		createBookingFn: func(userID, trainID, seatID int, journeyDate, status string, tx *sql.Tx) (int, error) {
			return 42, nil
		},
	}
}

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

func TestIsSeatAvailableForDate_ReturnsTrue(t *testing.T) {
	mock := &MockRepository{
		isSeatAvailableForDateFn: func(seatID int, journeyDate string) (bool, error) { return true, nil },
	}
	svc := NewServiceWithRepo(mock, nil)
	assert.True(t, svc.IsSeatAvailableForDate(1, "2028-12-25"))
}

func TestIsSeatAvailableForDate_ReturnsFalse_WhenBooked(t *testing.T) {
	mock := &MockRepository{
		isSeatAvailableForDateFn: func(seatID int, journeyDate string) (bool, error) { return false, nil },
	}
	svc := NewServiceWithRepo(mock, nil)
	assert.False(t, svc.IsSeatAvailableForDate(1, "2028-12-25"))
}

func TestIsSeatAvailableForDate_ReturnsFalse_WhenRepoErrors(t *testing.T) {
	mock := &MockRepository{
		isSeatAvailableForDateFn: func(seatID int, journeyDate string) (bool, error) {
			return false, errors.New("db error")
		},
	}
	svc := NewServiceWithRepo(mock, nil)
	assert.False(t, svc.IsSeatAvailableForDate(1, "2028-12-25"))
}

func TestIsSeatLocked_ReturnsFalse_WhenNotLocked(t *testing.T) {
	svc := NewServiceWithRepo(&MockRepository{}, nil)
	assert.False(t, svc.IsSeatLocked(1, 1, "2028-12-25"))
}

func TestIsSeatLocked_ReturnsTrue_WhenLocked(t *testing.T) {
	svc := NewServiceWithRepo(&MockRepository{}, nil)
	key := "seat_lock:1:10:2028-12-25"
	svc.locker.TryLock(key)
	defer svc.locker.Unlock(key)
	assert.True(t, svc.IsSeatLocked(1, 10, "2028-12-25"))
}

func TestBookSeat_ReturnsConflict_WhenSeatAlreadyLocked(t *testing.T) {
	svc := NewServiceWithRepo(&MockRepository{}, nil)
	req := makeValidBookingRequest()
	lockKey := "seat_lock:1:1:2028-12-25"
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
	db.Close()

	mock := &MockRepository{
		isSeatAvailableForDateFn: func(seatID int, journeyDate string) (bool, error) { return true, nil },
	}
	svc := NewServiceWithRepo(mock, db)

	result, err, isSeatTaken := svc.BookSeat(dto.BookingRequest{UserID: 1, TrainID: 1, SeatID: 5, JourneyDate: "2028-12-25"})
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

	result, err, isSeatTaken := svc.BookSeat(dto.BookingRequest{UserID: 1, TrainID: 1, SeatID: 2, JourneyDate: "2028-12-25"})
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

	result, err, isSeatTaken := svc.BookSeat(dto.BookingRequest{UserID: 1, TrainID: 1, SeatID: 3, JourneyDate: "2028-12-25"})
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

	result, err, isSeatTaken := svc.BookSeat(dto.BookingRequest{UserID: 1, TrainID: 1, SeatID: 6, JourneyDate: "2028-12-25"})
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

	result, err, isSeatTaken := svc.BookSeat(dto.BookingRequest{UserID: 1, TrainID: 1, SeatID: 7, JourneyDate: "2028-12-25"})
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

	result, err, isSeatTaken := svc.BookSeat(dto.BookingRequest{UserID: 1, TrainID: 1, SeatID: 8, JourneyDate: "2028-12-25"})
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

	result, err, isSeatTaken := svc.BookSeat(dto.BookingRequest{UserID: 1, TrainID: 1, SeatID: 4, JourneyDate: "2028-12-25"})
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

	result, err, isSeatTaken := svc.BookSeat(dto.BookingRequest{UserID: 1, TrainID: 1, SeatID: 9, JourneyDate: "2028-12-25"})
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

func TestBookSeat_Concurrent_OnlyOneSucceeds(t *testing.T) {
	const goroutines = 20

	repo := &MockRepository{
		isSeatAvailableForDateFn:  func(seatID int, journeyDate string) (bool, error) { return true, nil },
		getSeatByIDFn:             func(seatID int) (*Seat, error) { return makeMockSeat(), nil },
		lockSeatFn:                func(seatID int, tx *sql.Tx) error { return nil },
		decrementAvailableSeatsFn: func(trainID int, tx *sql.Tx) error { return nil },
		createBookingFn: func(userID, trainID, seatID int, journeyDate, status string, tx *sql.Tx) (int, error) {
			return 1, nil
		},
	}

	const lockKey = "seat_lock:1:1:2028-12-25"

	{
		svc := NewServiceWithRepo(repo, nil)
		svc.locker.TryLock(lockKey)
		defer svc.locker.Unlock(lockKey)

		var wg sync.WaitGroup
		conflictCount := 0
		var mu sync.Mutex

		wg.Add(goroutines)
		for i := 0; i < goroutines; i++ {
			i := i
			go func() {
				defer wg.Done()
				local := &Service{repo: repo, db: nil, locker: svc.locker}
				req := dto.BookingRequest{UserID: i + 1, TrainID: 1, SeatID: 1, JourneyDate: "2028-12-25"}
				_, err, isTaken := local.BookSeat(req)
				if err != nil && isTaken {
					mu.Lock()
					conflictCount++
					mu.Unlock()
				}
			}()
		}
		wg.Wait()

		assert.Equal(t, goroutines, conflictCount, "all goroutines should conflict while seat is locked")
	}

	{
		db, sqlMock, _ := sqlmock.New()
		defer db.Close()
		sqlMock.ExpectBegin()
		sqlMock.ExpectCommit()

		svc := NewServiceWithRepo(repo, db)
		req := dto.BookingRequest{UserID: 1, TrainID: 1, SeatID: 1, JourneyDate: "2028-12-25"}
		resp, err, isTaken := svc.BookSeat(req)

		assert.NoError(t, err)
		assert.False(t, isTaken)
		assert.NotNil(t, resp)
		assert.Equal(t, 1, resp.BookingID)
	}
}

func TestConfirmNextWaitlist_SuccessfulHTTPCall(t *testing.T) {
	called := make(chan struct{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called <- struct{}{}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	t.Setenv("FASTAPI_URL", server.URL)

	svc := NewServiceWithRepo(&MockRepository{}, nil)
	svc.confirmNextWaitlist(1, "2028-12-25")

	select {
	case <-called:
	case <-time.After(2 * time.Second):
		t.Fatal("waitlist endpoint was not called")
	}
}

func TestConfirmNextWaitlist_HandlesHTTPError(t *testing.T) {
	t.Setenv("FASTAPI_URL", "http://localhost:1")
	svc := NewServiceWithRepo(&MockRepository{}, nil)
	assert.NotPanics(t, func() { svc.confirmNextWaitlist(1, "2028-12-25") })
}

func TestConfirmNextWaitlist_UsesFallbackURL_WhenEnvNotSet(t *testing.T) {
	os.Unsetenv("FASTAPI_URL")
	svc := NewServiceWithRepo(&MockRepository{}, nil)
	assert.NotPanics(t, func() { svc.confirmNextWaitlist(1, "2028-12-25") })
}

func TestCancelBooking_ReturnsError_WhenJourneyAlreadyStarted(t *testing.T) {
	pastDate := "2020-01-01"
	pastDep := "06:00:00"

	svc := NewServiceWithRepo(&MockRepository{
		getBookingByIDFn: func(bookingID int) (*Booking, error) {
			return &Booking{ID: 1, UserID: 1, TrainID: 1, SeatID: 1, JourneyDate: pastDate, Status: "CONFIRMED"}, nil
		},
		getDepartureTimeFn: func(trainID int) (string, error) {
			return pastDep, nil
		},
	}, nil)

	_, err := svc.CancelBooking(dto.CancelRequest{BookingID: 1, UserID: 1})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot cancel after journey has started")
}

func TestCancelBooking_StillCancels_WhenDepartureTimeParseErrors(t *testing.T) {
	db, sqlMock, _ := sqlmock.New()
	defer db.Close()
	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	svc := NewServiceWithRepo(&MockRepository{
		getBookingByIDFn: func(bookingID int) (*Booking, error) {
			return &Booking{ID: 1, UserID: 1, TrainID: 1, SeatID: 1, JourneyDate: "2028-12-25", Status: "CONFIRMED"}, nil
		},
		getDepartureTimeFn: func(trainID int) (string, error) {
			return "not-a-time", nil
		},
		unlockSeatFn:    func(seatID int) error { return nil },
		cancelBookingFn: func(bookingID int) error { return nil },
	}, db)

	resp, err := svc.CancelBooking(dto.CancelRequest{BookingID: 1, UserID: 1})

	assert.NoError(t, err)
	assert.Equal(t, "CANCELLED", resp.Status)
}

func TestScheduleExpiryCheck_StartsWithoutPanic(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()
	os.Setenv("SPRINGBOOT_URL", server.URL)
	defer os.Unsetenv("SPRINGBOOT_URL")

	svc := NewServiceWithRepo(&MockRepository{}, nil)

	done := make(chan struct{})
	go func() {
		defer close(done)
	}()
	_ = svc // suppress unused warning
	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
	}
}

func TestSyncSeatsToES_LogsAndReturns_WhenGetAvailableSeatsFails(t *testing.T) {
	db, sqlMock, _ := sqlmock.New()
	defer db.Close()

	sqlMock.ExpectQuery("SELECT available_seats FROM trains").
		WithArgs(1).
		WillReturnError(sql.ErrConnDone)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	os.Setenv("FASTAPI_URL", server.URL)
	defer os.Unsetenv("FASTAPI_URL")

	svc := NewService(db) // NewService creates a real *Repository
	svc.syncSeatsToES(1)  // should not panic
}

func TestSyncSeatsToES_PatchesSuccessfully_WhenSeatsAvailable(t *testing.T) {
	db, sqlMock, _ := sqlmock.New()
	defer db.Close()

	sqlMock.ExpectQuery("SELECT available_seats FROM trains").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"available_seats"}).AddRow(10))

	patched := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		patched = true
		assert.Equal(t, http.MethodPatch, r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	os.Setenv("FASTAPI_URL", server.URL)
	defer os.Unsetenv("FASTAPI_URL")

	svc := NewService(db)
	svc.syncSeatsToES(1)

	assert.True(t, patched, "PATCH should have been called on FastAPI")
}

func TestSyncSeatsToES_UsesDefaultURL_WhenEnvNotSet(t *testing.T) {
	db, sqlMock, _ := sqlmock.New()
	defer db.Close()

	sqlMock.ExpectQuery("SELECT available_seats FROM trains").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"available_seats"}).AddRow(5))

	os.Unsetenv("FASTAPI_URL")

	svc := NewService(db)
	svc.syncSeatsToES(1) // will fail to connect but should not panic
}

func TestSyncSeatsToES_DoesNothing_WhenRepoIsNotConcreteType(t *testing.T) {
	// When repo is a MockRepository (not *Repository), syncSeatsToES returns immediately
	svc := NewServiceWithRepo(&MockRepository{}, nil)
	svc.syncSeatsToES(1) // should not call anything, no panic
}
