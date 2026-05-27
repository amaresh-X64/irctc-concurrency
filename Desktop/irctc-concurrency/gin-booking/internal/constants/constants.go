package constants

import (
	"log"
	"os"
)

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
	BookingExpirySeconds = 300
)

const (
	// PaymentStatusSuccess must match the status string springboot-payment writes.
	PaymentStatusSuccess = "SUCCESS"
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

const JWTAlgorithm = "HS256"

// JWTSecretKey is loaded once at startup from the environment.
// The process exits immediately if the variable is missing — a blank
// secret is worse than a crash because it silently accepts any token.
var JWTSecretKey string

func init() {
	JWTSecretKey = os.Getenv("JWT_SECRET_KEY")
	if JWTSecretKey == "" {
		log.Fatal("JWT_SECRET_KEY environment variable is not set")
	}
}