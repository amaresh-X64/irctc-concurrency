package booking

import "time"

type Booking struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	TrainID     int       `json:"train_id"`
	SeatID      int       `json:"seat_id"`
	SeatNumber  string    `json:"seat_number"`
	JourneyDate string    `json:"journey_date"`
	Status      string    `json:"status"`
	BookedAt    time.Time `json:"booked_at"`
}

type Seat struct {
	ID          int    `json:"id"`
	TrainID     int    `json:"train_id"`
	SeatNumber  string `json:"seat_number"`
	SeatType    string `json:"seat_type"`
	IsAvailable bool   `json:"is_available"`
}
