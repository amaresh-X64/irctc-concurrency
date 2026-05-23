package dto

type BookingRequest struct {
	UserID      int    `json:"user_id" binding:"required"`
	TrainID     int    `json:"train_id" binding:"required"`
	SeatID      int    `json:"seat_id" binding:"required"`
	JourneyDate string `json:"journey_date" binding:"required"`
}

type CancelRequest struct {
	BookingID int `json:"booking_id" binding:"required"`
	UserID    int `json:"user_id" binding:"required"`
}
