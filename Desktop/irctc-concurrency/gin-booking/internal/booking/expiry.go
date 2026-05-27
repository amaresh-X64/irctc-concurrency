package booking

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"gin-booking/internal/constants"
)

// paymentCheckResponse mirrors the shape springboot-payment returns for
// GET /api/v1/payments/booking/{bookingId}
type paymentCheckResponse struct {
	Data struct {
		Status string `json:"status"`
	} `json:"data"`
}

// ScheduleExpiryCheck fires once after BookingExpirySeconds. It calls the
// payment service; if no SUCCESS payment exists the booking is hard-deleted
// and the seat + train counter are restored.
//
// Called as a goroutine immediately after a successful BookSeat commit.
func (s *Service) ScheduleExpiryCheck(bookingID, trainID, seatID int) {
	ttl := time.Duration(constants.BookingExpirySeconds) * time.Second
	log.Printf("[expiry] booking %d — will check in %v", bookingID, ttl)

	time.Sleep(ttl)

	if s.isPaid(bookingID) {
		log.Printf("[expiry] booking %d — payment found, no action", bookingID)
		return
	}

	log.Printf("[expiry] booking %d — no SUCCESS payment after %v, purging", bookingID, ttl)
	s.purgeUnpaidBooking(bookingID, trainID, seatID)
}

// isPaid returns true only when the payment service reports status SUCCESS.
// Any network error, 404, or non-SUCCESS status is treated as "not paid".
func (s *Service) isPaid(bookingID int) bool {
	paymentURL := os.Getenv("SPRINGBOOT_URL")

	url := fmt.Sprintf("%s/api/v1/payments/booking/%d", paymentURL, bookingID)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		log.Printf("[expiry] booking %d — payment service unreachable: %v", bookingID, err)
		return false
	}
	defer resp.Body.Close()

	// 404 → no payment record at all
	if resp.StatusCode == http.StatusNotFound {
		return false
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("[expiry] booking %d — payment service returned %d", bookingID, resp.StatusCode)
		return false
	}

	var result paymentCheckResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("[expiry] booking %d — failed to decode payment response: %v", bookingID, err)
		return false
	}

	return result.Data.Status == constants.PaymentStatusSuccess
}

// purgeUnpaidBooking deletes the booking row, resets the seat to available,
// and increments the train's available_seats counter — all in one transaction.
func (s *Service) purgeUnpaidBooking(bookingID, trainID, seatID int) {
	tx, err := s.db.Begin()
	if err != nil {
		log.Printf("[expiry] booking %d — failed to start purge tx: %v", bookingID, err)
		return
	}

	if err := s.repo.DeleteBooking(bookingID, tx); err != nil {
		tx.Rollback()
		log.Printf("[expiry] booking %d — DeleteBooking failed: %v", bookingID, err)
		return
	}

	if err := s.repo.UnlockSeatTx(seatID, tx); err != nil {
		tx.Rollback()
		log.Printf("[expiry] booking %d — UnlockSeat failed: %v", bookingID, err)
		return
	}

	if err := s.repo.IncrementAvailableSeats(trainID, tx); err != nil {
		tx.Rollback()
		log.Printf("[expiry] booking %d — IncrementAvailableSeats failed: %v", bookingID, err)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("[expiry] booking %d — purge commit failed: %v", bookingID, err)
		return
	}

	log.Printf("[expiry] booking %d purged — seat %d unlocked, train %d seat count restored",
		bookingID, seatID, trainID)
}