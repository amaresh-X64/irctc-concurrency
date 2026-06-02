package booking

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestIsPaid_ReturnsFalse_WhenPaymentServiceUnreachable(t *testing.T) {
	svc := &Service{}
	t.Setenv("SPRINGBOOT_URL", "http://localhost:1") // nothing listening there

	result := svc.isPaid(1)

	assert.False(t, result)
}

func TestIsPaid_ReturnsFalse_WhenPaymentReturns404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	svc := &Service{}
	t.Setenv("SPRINGBOOT_URL", server.URL)

	result := svc.isPaid(42)

	assert.False(t, result)
}

func TestIsPaid_ReturnsFalse_WhenPaymentStatusNotSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":{"status":"PENDING"}}`))
	}))
	defer server.Close()

	svc := &Service{}
	t.Setenv("SPRINGBOOT_URL", server.URL)

	result := svc.isPaid(42)

	assert.False(t, result)
}

func TestIsPaid_ReturnsTrue_WhenPaymentStatusIsSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":{"status":"SUCCESS"}}`))
	}))
	defer server.Close()

	svc := &Service{}
	t.Setenv("SPRINGBOOT_URL", server.URL)

	result := svc.isPaid(42)

	assert.True(t, result)
}

func TestIsPaid_ReturnsFalse_WhenResponseBodyIsInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not-json`))
	}))
	defer server.Close()

	svc := &Service{}
	t.Setenv("SPRINGBOOT_URL", server.URL)

	result := svc.isPaid(42)

	assert.False(t, result)
}

func TestIsPaid_ReturnsFalse_WhenPaymentReturnsNon200NonSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	svc := &Service{}
	t.Setenv("SPRINGBOOT_URL", server.URL)

	result := svc.isPaid(42)

	assert.False(t, result)
}

func TestPurgeUnpaidBooking_CommitsSuccessfully_WhenAllStepsPass(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := &MockRepository{
		deleteBookingFn:           func(bookingID int, tx *sql.Tx) error { return nil },
		unlockSeatTxFn:            func(seatID int, tx *sql.Tx) error { return nil },
		incrementAvailableSeatsFn: func(trainID int, tx *sql.Tx) error { return nil },
	}

	mock.ExpectBegin()
	mock.ExpectCommit()

	svc := &Service{repo: repo, db: db}
	svc.purgeUnpaidBooking(1, 1, 1)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPurgeUnpaidBooking_Rollsback_WhenDeleteBookingFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := &MockRepository{
		deleteBookingFn: func(bookingID int, tx *sql.Tx) error {
			return sql.ErrConnDone
		},
	}

	mock.ExpectBegin()
	mock.ExpectRollback()

	svc := &Service{repo: repo, db: db}
	svc.purgeUnpaidBooking(1, 1, 1) // should not panic

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPurgeUnpaidBooking_Rollsback_WhenUnlockSeatFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := &MockRepository{
		deleteBookingFn: func(bookingID int, tx *sql.Tx) error { return nil },
		unlockSeatTxFn:  func(seatID int, tx *sql.Tx) error { return sql.ErrConnDone },
	}

	mock.ExpectBegin()
	mock.ExpectRollback()

	svc := &Service{repo: repo, db: db}
	svc.purgeUnpaidBooking(1, 1, 1)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPurgeUnpaidBooking_Rollsback_WhenIncrementFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := &MockRepository{
		deleteBookingFn:           func(bookingID int, tx *sql.Tx) error { return nil },
		unlockSeatTxFn:            func(seatID int, tx *sql.Tx) error { return nil },
		incrementAvailableSeatsFn: func(trainID int, tx *sql.Tx) error { return sql.ErrConnDone },
	}

	mock.ExpectBegin()
	mock.ExpectRollback()

	svc := &Service{repo: repo, db: db}
	svc.purgeUnpaidBooking(1, 1, 1)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPurgeUnpaidBooking_LogsError_WhenDBBeginFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin().WillReturnError(sql.ErrConnDone)

	svc := &Service{repo: &MockRepository{}, db: db}
	svc.purgeUnpaidBooking(1, 1, 1) // should not panic

	assert.NoError(t, mock.ExpectationsWereMet())
}
