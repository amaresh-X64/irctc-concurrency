from sqlalchemy.orm import Session
from app.waitlist.repository import WaitlistRepository
from app.constants.constants import (WAITLIST_CONFIRM_PROBABILITY_HIGH,
                                     WAITLIST_CONFIRM_PROBABILITY_MED,
                                     WAITLIST_CONFIRM_PROBABILITY_LOW)
from app.helpers.response import success_response, error_response


class WaitlistService:

    def __init__(self, db: Session):
        self.repo = WaitlistRepository(db)

    def get_waitlist_info(self, train_id: int, journey_date: str):
        result = self.repo.get_waitlist_count(train_id, journey_date)
        waiting_count = result.waiting_count if result else 0
        probability, message = self._probability_message(waiting_count)
        return success_response("Waitlist info fetched", {
            "total_waiting": waiting_count,
            "confirmation_probability": probability,
            "message": message,
        })

    def join_waitlist(self, user_id: int, train_id: int, journey_date: str):
        self.repo.cleanup_user_waitlist(user_id, train_id, journey_date)

        existing = self.repo.find_existing(user_id, train_id, journey_date)
        if existing:
            return success_response("Already on waitlist", {
                "position": existing.position,
                "already_exists": True,
                "confirmation_probability": self._get_probability(existing.position),
                "message": f"You are already #{existing.position} on the waitlist",
            })

        self.repo.insert_waitlist(user_id, train_id, journey_date)
        new_entry = self.repo.find_existing(user_id, train_id, journey_date)
        position = new_entry.position if new_entry else 1

        return success_response("Added to waitlist", {
            "position": position,
            "already_exists": False,
            "confirmation_probability": self._get_probability(position),
            "message": f"You are #{position} on the waitlist",
        })

    def get_user_waitlist(self, user_id: int):
        results = self.repo.get_user_waitlist(user_id)
        data = [{
            "id": r.id,
            "train_id": r.train_id,
            "train_name": r.train_name,
            "user_name": r.user_name,
            "source": r.source,
            "destination": r.destination,
            "journey_date": str(r.journey_date),
            "position": r.position,
            "status": r.status,
            "probability": self._get_probability(r.position),
        } for r in results]
        return success_response("Waitlist fetched", data)

    def confirm_next_waitlist(self, train_id: int, journey_date: str):
        """
        Delegates entirely to the repository's atomic CTE.
        No seat is double-booked even under concurrent cancellations.
        """
        row = self.repo.confirm_next_atomically(train_id, journey_date)

        if row is None:
            return success_response("No one in waitlist or no seats available", {"confirmed": False})

        return success_response("Waitlist confirmed and booking created", {
            "confirmed": True,
            "user_id": row.user_id,
            "booking_id": row.booking_id,
            "seat_number": row.seat_number,
            "waitlist_id": row.waitlist_id,
        })

    # ── helpers ──────────────────────────────────────────────────────────────

    def _get_probability(self, position: int) -> float:
        if position <= 5:
            return WAITLIST_CONFIRM_PROBABILITY_HIGH
        elif position <= 15:
            return WAITLIST_CONFIRM_PROBABILITY_MED
        return WAITLIST_CONFIRM_PROBABILITY_LOW

    def _probability_message(self, count: int):
        if count <= 5:
            return WAITLIST_CONFIRM_PROBABILITY_HIGH, "High chance of confirmation!"
        elif count <= 15:
            return WAITLIST_CONFIRM_PROBABILITY_MED, "Moderate chance of confirmation"
        return WAITLIST_CONFIRM_PROBABILITY_LOW, "Low chance of confirmation"