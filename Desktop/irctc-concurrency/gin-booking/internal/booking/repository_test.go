package booking_test

import (
	"database/sql"
	"testing"
	"time"

	"gin-booking/internal/booking"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestNewRepository_ShouldReturnRepository_WhenDBProvided(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := booking.NewRepository(db)

	assert.NotNil(t, repo)
}

func TestGetSeatByID_ShouldReturnSeat_WhenSeatExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := booking.NewRepository(db)

	rows := sqlmock.NewRows([]string{
		"id",
		"train_id",
		"seat_number",
		"seat_type",
		"is_available",
	}).AddRow(1, 101, "A1", "GENERAL", true)

	mock.ExpectQuery("SELECT id, train_id, seat_number, seat_type, is_available").
		WithArgs(1).
		WillReturnRows(rows)

	seat, err := repo.GetSeatByID(1)

	assert.NoError(t, err)
	assert.NotNil(t, seat)
	assert.Equal(t, 1, seat.ID)
	assert.Equal(t, 101, seat.TrainID)
	assert.Equal(t, "A1", seat.SeatNumber)
	assert.True(t, seat.IsAvailable)
}

func TestGetSeatByID_ShouldReturnError_WhenSeatNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := booking.NewRepository(db)

	mock.ExpectQuery("SELECT id, train_id, seat_number, seat_type, is_available").
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	seat, err := repo.GetSeatByID(999)

	assert.Error(t, err)
	assert.Nil(t, seat)
}

func TestLockSeat_ShouldReturnNil_WhenSeatLockedSuccessfully(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := booking.NewRepository(db)

	mock.ExpectBegin()

	tx, err := db.Begin()
	assert.NoError(t, err)

	mock.ExpectExec("UPDATE seats SET is_available = false").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.LockSeat(1, tx)

	assert.NoError(t, err)
}

func TestLockSeat_ShouldReturnError_WhenDatabaseFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := booking.NewRepository(db)

	mock.ExpectBegin()

	tx, err := db.Begin()
	assert.NoError(t, err)

	mock.ExpectExec("UPDATE seats SET is_available = false").
		WithArgs(1).
		WillReturnError(sql.ErrConnDone)

	err = repo.LockSeat(1, tx)

	assert.Error(t, err)
}

func TestDecrementAvailableSeats_ShouldReturnNil_WhenSeatsUpdated(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := booking.NewRepository(db)

	mock.ExpectBegin()

	tx, err := db.Begin()
	assert.NoError(t, err)

	mock.ExpectExec("UPDATE trains").
		WithArgs(101).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.DecrementAvailableSeats(101, tx)

	assert.NoError(t, err)
}

func TestDecrementAvailableSeats_ShouldReturnError_WhenDatabaseFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := booking.NewRepository(db)

	mock.ExpectBegin()

	tx, err := db.Begin()
	assert.NoError(t, err)

	mock.ExpectExec("UPDATE trains").
		WithArgs(101).
		WillReturnError(sql.ErrConnDone)

	err = repo.DecrementAvailableSeats(101, tx)

	assert.Error(t, err)
}

func TestCreateBooking_ShouldReturnBookingID_WhenBookingCreated(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := booking.NewRepository(db)

	mock.ExpectBegin()

	tx, err := db.Begin()
	assert.NoError(t, err)

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1001)

	mock.ExpectQuery("INSERT INTO bookings").
		WithArgs(
			1,
			101,
			1,
			"2025-06-01",
			"CONFIRMED",
			sqlmock.AnyArg(),
		).
		WillReturnRows(rows)

	bookingID, err := repo.CreateBooking(
		1,
		101,
		1,
		"2025-06-01",
		"CONFIRMED",
		tx,
	)

	assert.NoError(t, err)
	assert.Equal(t, 1001, bookingID)
}

func TestCreateBooking_ShouldReturnError_WhenInsertFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := booking.NewRepository(db)

	mock.ExpectBegin()

	tx, err := db.Begin()
	assert.NoError(t, err)

	mock.ExpectQuery("INSERT INTO bookings").
		WillReturnError(sql.ErrConnDone)

	bookingID, err := repo.CreateBooking(
		1,
		101,
		1,
		"2025-06-01",
		"CONFIRMED",
		tx,
	)

	assert.Error(t, err)
	assert.Equal(t, 0, bookingID)
}

func TestGetBookingByID_ShouldReturnBooking_WhenBookingExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := booking.NewRepository(db)

	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id",
		"user_id",
		"train_id",
		"seat_id",
		"journey_date",
		"status",
		"booked_at",
	}).AddRow(
		1,
		1,
		101,
		1,
		"2025-06-01",
		"CONFIRMED",
		now,
	)

	mock.ExpectQuery("SELECT id, user_id, train_id, seat_id, journey_date, status, booked_at").
		WithArgs(1).
		WillReturnRows(rows)

	bookingObj, err := repo.GetBookingByID(1)

	assert.NoError(t, err)
	assert.NotNil(t, bookingObj)
	assert.Equal(t, 1, bookingObj.ID)
	assert.Equal(t, 1, bookingObj.UserID)
}

func TestGetBookingByID_ShouldReturnError_WhenBookingMissing(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := booking.NewRepository(db)

	mock.ExpectQuery("SELECT id, user_id, train_id, seat_id, journey_date, status, booked_at").
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	bookingObj, err := repo.GetBookingByID(999)

	assert.Error(t, err)
	assert.Nil(t, bookingObj)
}

func TestCancelBooking_ShouldReturnNil_WhenBookingCancelled(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := booking.NewRepository(db)

	mock.ExpectExec("UPDATE bookings SET status = 'CANCELLED'").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.CancelBooking(1)

	assert.NoError(t, err)
}

func TestCancelBooking_ShouldReturnError_WhenUpdateFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := booking.NewRepository(db)

	mock.ExpectExec("UPDATE bookings SET status = 'CANCELLED'").
		WithArgs(1).
		WillReturnError(sql.ErrConnDone)

	err = repo.CancelBooking(1)

	assert.Error(t, err)
}

func TestUnlockSeat_ShouldReturnNil_WhenSeatUnlocked(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := booking.NewRepository(db)

	mock.ExpectExec("UPDATE seats SET is_available = true").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.UnlockSeat(1)

	assert.NoError(t, err)
}

func TestUnlockSeat_ShouldReturnError_WhenUpdateFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := booking.NewRepository(db)

	mock.ExpectExec("UPDATE seats SET is_available = true").
		WithArgs(1).
		WillReturnError(sql.ErrConnDone)

	err = repo.UnlockSeat(1)

	assert.Error(t, err)
}

func TestGetBookingsByUser_ShouldReturnBookings_WhenBookingsExist(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := booking.NewRepository(db)

	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id",
		"user_id",
		"train_id",
		"seat_id",
		"journey_date",
		"status",
		"booked_at",
		"seat_number",
	}).
		AddRow(1, 1, 101, 1, "2025-06-01", "CONFIRMED", now, "A1").
		AddRow(2, 1, 101, 2, "2025-06-02", "CONFIRMED", now, "A2")

	mock.ExpectQuery("SELECT b.id, b.user_id").
		WithArgs(1).
		WillReturnRows(rows)

	bookings, err := repo.GetBookingsByUser(1)

	assert.NoError(t, err)
	assert.Len(t, bookings, 2)
	assert.Equal(t, "A1", bookings[0].SeatNumber)
	assert.Equal(t, "A2", bookings[1].SeatNumber)
}

func TestGetBookingsByUser_ShouldReturnEmptySlice_WhenNoBookingsExist(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := booking.NewRepository(db)

	rows := sqlmock.NewRows([]string{
		"id",
		"user_id",
		"train_id",
		"seat_id",
		"journey_date",
		"status",
		"booked_at",
		"seat_number",
	})

	mock.ExpectQuery("SELECT b.id, b.user_id").
		WithArgs(1).
		WillReturnRows(rows)

	bookings, err := repo.GetBookingsByUser(1)

	assert.NoError(t, err)
	assert.Len(t, bookings, 0)
}

func TestGetBookingsByUser_ShouldReturnError_WhenQueryFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := booking.NewRepository(db)

	mock.ExpectQuery("SELECT b.id, b.user_id").
		WithArgs(1).
		WillReturnError(sql.ErrConnDone)

	bookings, err := repo.GetBookingsByUser(1)

	assert.Error(t, err)
	assert.Nil(t, bookings)
}

func TestIsSeatAvailableForDate_ShouldReturnTrue_WhenSeatNotBooked(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := booking.NewRepository(db)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(0)

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM bookings").
		WithArgs(1, "2025-06-01").
		WillReturnRows(rows)

	available, err := repo.IsSeatAvailableForDate(1, "2025-06-01")

	assert.NoError(t, err)
	assert.True(t, available)
}

func TestIsSeatAvailableForDate_ShouldReturnFalse_WhenSeatAlreadyBooked(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := booking.NewRepository(db)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(1)

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM bookings").
		WithArgs(1, "2025-06-01").
		WillReturnRows(rows)

	available, err := repo.IsSeatAvailableForDate(1, "2025-06-01")

	assert.NoError(t, err)
	assert.False(t, available)
}

func TestIsSeatAvailableForDate_ShouldReturnError_WhenDatabaseFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := booking.NewRepository(db)

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM bookings").
		WithArgs(1, "2025-06-01").
		WillReturnError(sql.ErrConnDone)

	available, err := repo.IsSeatAvailableForDate(1, "2025-06-01")

	assert.Error(t, err)
	assert.False(t, available)
}
