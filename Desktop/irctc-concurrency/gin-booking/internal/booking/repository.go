package booking

import (
	"database/sql"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetSeatByID(seatID int) (*Seat, error) {
	seat := &Seat{}
	err := r.db.QueryRow(`
		SELECT id, train_id, seat_number, seat_type, is_available
		FROM seats WHERE id = $1`, seatID).Scan(&seat.ID, &seat.TrainID, &seat.SeatNumber, &seat.SeatType, &seat.IsAvailable)
	if err != nil {
		return nil, err
	}
	return seat, nil
}

func (r *Repository) LockSeat(seatID int, tx *sql.Tx) error {
	_, err := tx.Exec(`UPDATE seats SET is_available = false WHERE id = $1`, seatID)
	return err
}

func (r *Repository) DecrementAvailableSeats(trainID int, tx *sql.Tx) error {
	_, err := tx.Exec(`UPDATE trains
		SET available_seats = available_seats - 1
		WHERE id = $1 AND available_seats > 0`, trainID)
	return err
}

func (r *Repository) CreateBooking(userID, trainID, seatID int, journeyDate, status string, tx *sql.Tx) (int, error) {
	var bookingID int
	err := tx.QueryRow(`INSERT INTO bookings (user_id, train_id, seat_id, journey_date, status, booked_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`, userID, trainID, seatID, journeyDate, status, time.Now()).Scan(&bookingID)
	return bookingID, err
}

func (r *Repository) GetBookingByID(bookingID int) (*Booking, error) {
	booking := &Booking{}
	err := r.db.QueryRow(`
		SELECT id, user_id, train_id, seat_id, journey_date, status, booked_at
		FROM bookings WHERE id = $1`, bookingID).Scan(&booking.ID, &booking.UserID, &booking.TrainID, &booking.SeatID, &booking.JourneyDate,
		&booking.Status, &booking.BookedAt,
	)
	if err != nil {
		return nil, err
	}
	return booking, nil
}

func (r *Repository) CancelBooking(bookingID int) error {
	_, err := r.db.Exec(`
		UPDATE bookings SET status = 'CANCELLED' WHERE id = $1`, bookingID)
	return err
}

func (r *Repository) UnlockSeat(seatID int) error {
	_, err := r.db.Exec(`
		UPDATE seats SET is_available = true WHERE id = $1`, seatID)
	return err
}

func (r *Repository) UnlockSeatTx(seatID int, tx *sql.Tx) error {
	_, err := tx.Exec(`UPDATE seats SET is_available = true WHERE id = $1`, seatID)
	return err
}

func (r *Repository) DeleteBooking(bookingID int, tx *sql.Tx) error {
	_, err := tx.Exec(`DELETE FROM bookings WHERE id = $1`, bookingID)
	return err
}

func (r *Repository) IncrementAvailableSeats(trainID int, tx *sql.Tx) error {
	_, err := tx.Exec(`
		UPDATE trains
		SET available_seats = available_seats + 1
		WHERE id = $1`, trainID)
	return err
}

func (r *Repository) GetBookingsByUser(userID int) ([]Booking, error) {
	rows, err := r.db.Query(`
		SELECT b.id, b.user_id, b.train_id, b.seat_id, b.journey_date, b.status, b.booked_at,s.seat_number
		FROM bookings b
		JOIN seats s ON s.id = b.seat_id
		WHERE b.user_id = $1
		ORDER BY b.booked_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var bookings []Booking
	for rows.Next() {
		var b Booking
		err := rows.Scan(&b.ID, &b.UserID, &b.TrainID, &b.SeatID, &b.JourneyDate, &b.Status, &b.BookedAt, &b.SeatNumber)
		if err != nil {
			continue
		}
		bookings = append(bookings, b)
	}
	return bookings, nil
}

func (r *Repository) IsSeatAvailableForDate(seatID int, journeyDate string) (bool, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM bookings
		WHERE seat_id = $1
		AND journey_date = $2
		AND status = 'CONFIRMED'
	`, seatID, journeyDate).Scan(&count)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}
func (r *Repository) GetDepartureTime(trainID int) (string, error) {
	var departureTime string
	err := r.db.QueryRow(`
        SELECT departure_time FROM trains WHERE id = $1`, trainID).Scan(&departureTime)
	return departureTime, err
}
func (r *Repository) GetAvailableSeats(trainID int) (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT available_seats FROM trains WHERE id = $1`, trainID).Scan(&count)
	return count, err
}
