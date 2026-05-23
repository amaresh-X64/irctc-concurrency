package constants

const (
	AppName    = "IRCTC Booking Service"
	AppVersion = "1.0.0"
	APIPrefix  = "/api/v1"
)

const (
	StatusPending    = "PENDING"
	StatusConfirmed  = "CONFIRMED"
	StatusCancelled  = "CANCELLED"
	StatusWaitlisted = "WAITLISTED"
)

const (
	SeatLockPrefix = "seat_lock:"
	SeatLockTTL    = 300
)

const (
	MsgBookingSuccess  = "Seat booked successfully"
	MsgSeatTaken       = "Seat already taken — try another seat"
	MsgBookingNotFound = "Booking not found"
	MsgCancelSuccess   = "Booking cancelled successfully"
	MsgServerError     = "Internal server error"
	MsgInvalidRequest  = "Invalid request body"
	MsgUnauthorized    = "Please login to continue"
	MsgTokenExpired    = "Session expired, please login again"
)

const (
	JWTSecretKey = "irctc_super_secret_key_2024"
	JWTAlgorithm = "HS256"
)
