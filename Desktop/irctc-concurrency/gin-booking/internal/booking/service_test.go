package booking

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"gin-booking/internal/booking/dto"
	"gin-booking/internal/constants"
	redisClient "gin-booking/pkg/redis"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	redisClient.Client = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	os.Exit(m.Run())
}

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

func cleanLockKey(trainID, seatID int, journeyDate string) {
	lockKey := fmt.Sprintf("%s%d:%d:%s", constants.SeatLockPrefix, trainID, seatID, journeyDate)
	redisClient.Client.Del(redisClient.Ctx, lockKey)
}

func setLockKey(trainID, seatID int, journeyDate string) {
	lockKey := fmt.Sprintf("%s%d:%d:%s", constants.SeatLockPrefix, trainID, seatID, journeyDate)
	redisClient.Client.Set(redisClient.Ctx, lockKey, 99, time.Duration(constants.SeatLockTTL)*time.Second)
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
		getBookingByIDFn: func(bookingID int) (*Booking, error) {
			return makeMockBooking(), nil
		},
		unlockSeatFn:    func(seatID int) error { return nil },
		cancelBookingFn: func(bookingID int) error { return nil },
	}
	svc := NewServiceWithRepo(mock, nil)
	result, err := svc.CancelBooking(dto.CancelRequest{BookingID: 1})
	assert.NoError(t, err)
	assert.Equal(t, constants.StatusCancelled, result.Status)
	assert.Equal(t, 1, result.BookingID)
}

func TestIsSeatAvailableForDate_ReturnsTrue(t *testing.T) {
	mock := &MockRepository{
		isSeatAvailableForDateFn: func(seatID int, journeyDate string) (bool, error) {
			return true, nil
		},
	}
	svc := NewServiceWithRepo(mock, nil)
	assert.True(t, svc.IsSeatAvailableForDate(1, "2024-12-25"))
}

func TestIsSeatAvailableForDate_ReturnsFalse_WhenBooked(t *testing.T) {
	mock := &MockRepository{
		isSeatAvailableForDateFn: func(seatID int, journeyDate string) (bool, error) {
			return false, nil
		},
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

func TestIsSeatLocked_ReturnsFalse_WhenNotLocked(t *testing.T) {
	svc := NewServiceWithRepo(&MockRepository{}, nil)
	cleanLockKey(1, 1, "2024-12-25")
	assert.False(t, svc.IsSeatLocked(1, 1, "2024-12-25"))
}

func TestIsSeatLocked_ReturnsTrue_WhenLocked(t *testing.T) {
	svc := NewServiceWithRepo(&MockRepository{}, nil)
	setLockKey(1, 10, "2024-12-25")
	defer cleanLockKey(1, 10, "2024-12-25")
	assert.True(t, svc.IsSeatLocked(1, 10, "2024-12-25"))
}

func TestBookSeat_ReturnsError_WhenRedisUnavailable(t *testing.T) {
	svc := NewServiceWithRepo(&MockRepository{}, nil)
	original := redisClient.Client
	redisClient.Client = redis.NewClient(&redis.Options{Addr: "localhost:1"})
	defer func() { redisClient.Client = original }()

	result, err, isSeatTaken := svc.BookSeat(makeValidBookingRequest())
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.False(t, isSeatTaken)
}

func TestBookSeat_ReturnsConflict_WhenSeatAlreadyLocked(t *testing.T) {
	svc := NewServiceWithRepo(&MockRepository{}, nil)
	req := makeValidBookingRequest()
	cleanLockKey(req.TrainID, req.SeatID, req.JourneyDate)
	setLockKey(req.TrainID, req.SeatID, req.JourneyDate)
	defer cleanLockKey(req.TrainID, req.SeatID, req.JourneyDate)

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
		isSeatAvailableForDateFn: func(seatID int, journeyDate string) (bool, error) {
			return true, nil
		},
	}
	svc := NewServiceWithRepo(mock, db)

	cleanLockKey(1, 5, "2024-12-25")
	defer cleanLockKey(1, 5, "2024-12-25")

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
		isSeatAvailableForDateFn: func(seatID int, journeyDate string) (bool, error) {
			return false, nil
		},
	}
	svc := NewServiceWithRepo(mock, db)

	cleanLockKey(1, 2, "2024-12-25")
	defer cleanLockKey(1, 2, "2024-12-25")

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
		isSeatAvailableForDateFn: func(seatID int, journeyDate string) (bool, error) {
			return true, nil
		},
		getSeatByIDFn: func(seatID int) (*Seat, error) {
			return nil, errors.New("seat not found")
		},
	}
	svc := NewServiceWithRepo(mock, db)

	cleanLockKey(1, 3, "2024-12-25")
	defer cleanLockKey(1, 3, "2024-12-25")

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
		isSeatAvailableForDateFn: func(seatID int, journeyDate string) (bool, error) {
			return true, nil
		},
		getSeatByIDFn: func(seatID int) (*Seat, error) {
			return makeMockSeat(), nil
		},
		lockSeatFn: func(seatID int, tx *sql.Tx) error {
			return errors.New("lock failed")
		},
	}
	svc := NewServiceWithRepo(mock, db)

	cleanLockKey(1, 6, "2024-12-25")
	defer cleanLockKey(1, 6, "2024-12-25")

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
		isSeatAvailableForDateFn: func(seatID int, journeyDate string) (bool, error) {
			return true, nil
		},
		getSeatByIDFn: func(seatID int) (*Seat, error) {
			return makeMockSeat(), nil
		},
		lockSeatFn: func(seatID int, tx *sql.Tx) error { return nil },
		decrementAvailableSeatsFn: func(trainID int, tx *sql.Tx) error {
			return errors.New("decrement failed")
		},
	}
	svc := NewServiceWithRepo(mock, db)

	cleanLockKey(1, 7, "2024-12-25")
	defer cleanLockKey(1, 7, "2024-12-25")

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
		isSeatAvailableForDateFn: func(seatID int, journeyDate string) (bool, error) {
			return true, nil
		},
		getSeatByIDFn: func(seatID int) (*Seat, error) { return makeMockSeat(), nil },
		lockSeatFn:    func(seatID int, tx *sql.Tx) error { return nil },
		decrementAvailableSeatsFn: func(trainID int, tx *sql.Tx) error {
			return nil
		},
		createBookingFn: func(userID, trainID, seatID int, journeyDate, status string, tx *sql.Tx) (int, error) {
			return 0, errors.New("insert failed")
		},
	}
	svc := NewServiceWithRepo(mock, db)

	cleanLockKey(1, 8, "2024-12-25")
	defer cleanLockKey(1, 8, "2024-12-25")

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

	mock := &MockRepository{
		isSeatAvailableForDateFn: func(seatID int, journeyDate string) (bool, error) {
			return true, nil
		},
		getSeatByIDFn: func(seatID int) (*Seat, error) { return makeMockSeat(), nil },
		lockSeatFn:    func(seatID int, tx *sql.Tx) error { return nil },
		decrementAvailableSeatsFn: func(trainID int, tx *sql.Tx) error {
			return nil
		},
		createBookingFn: func(userID, trainID, seatID int, journeyDate, status string, tx *sql.Tx) (int, error) {
			return 42, nil
		},
	}
	svc := NewServiceWithRepo(mock, db)

	cleanLockKey(1, 4, "2024-12-25")
	defer cleanLockKey(1, 4, "2024-12-25")

	result, err, isSeatTaken := svc.BookSeat(dto.BookingRequest{UserID: 1, TrainID: 1, SeatID: 4, JourneyDate: "2024-12-25"})
	assert.NoError(t, err)
	assert.False(t, isSeatTaken)
	assert.NotNil(t, result)
	assert.Equal(t, 42, result.BookingID)
	assert.Equal(t, constants.StatusConfirmed, result.Status)
	assert.Equal(t, "A1", result.SeatNumber)
}

func TestNewService_ReturnsServiceWithRepo(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()
	svc := NewService(db)
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.repo)
	assert.Equal(t, db, svc.db)
}

func TestBookSeat_ReturnsError_WhenCommitFails(t *testing.T) {
	db, sqlMock, _ := sqlmock.New()
	defer db.Close()
	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit().WillReturnError(errors.New("commit failed"))

	mock := &MockRepository{
		isSeatAvailableForDateFn: func(seatID int, journeyDate string) (bool, error) {
			return true, nil
		},
		getSeatByIDFn:             func(seatID int) (*Seat, error) { return makeMockSeat(), nil },
		lockSeatFn:                func(seatID int, tx *sql.Tx) error { return nil },
		decrementAvailableSeatsFn: func(trainID int, tx *sql.Tx) error { return nil },
		createBookingFn: func(userID, trainID, seatID int, journeyDate, status string, tx *sql.Tx) (int, error) {
			return 1, nil
		},
	}
	svc := NewServiceWithRepo(mock, db)

	cleanLockKey(1, 9, "2024-12-25")
	defer cleanLockKey(1, 9, "2024-12-25")

	result, err, isSeatTaken := svc.BookSeat(dto.BookingRequest{UserID: 1, TrainID: 1, SeatID: 9, JourneyDate: "2024-12-25"})
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.False(t, isSeatTaken)
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
	assert.NotPanics(t, func() {
		svc.confirmNextWaitlist(1, "2024-12-25")
	})
}

func TestConfirmNextWaitlist_UsesFallbackURL_WhenEnvNotSet(t *testing.T) {
	os.Unsetenv("FASTAPI_URL")

	svc := NewServiceWithRepo(&MockRepository{}, nil)
	assert.NotPanics(t, func() {
		svc.confirmNextWaitlist(1, "2024-12-25")
	})
}
