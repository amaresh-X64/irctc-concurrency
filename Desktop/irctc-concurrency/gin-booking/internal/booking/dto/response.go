package dto

import "time"

type BookingResponse struct {
	BookingID   int       `json:"booking_id"`
	UserID      int       `json:"user_id"`
	TrainID     int       `json:"train_id"`
	SeatID      int       `json:"seat_id"`
	SeatNumber  string    `json:"seat_number"`
	JourneyDate string    `json:"journey_date"`
	Status      string    `json:"status"`
	BookedAt    time.Time `json:"booked_at"`
}

type CancelResponse struct {
	BookingID int    `json:"booking_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}
