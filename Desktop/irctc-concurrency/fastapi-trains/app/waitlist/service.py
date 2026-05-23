from sqlalchemy.orm import Session
from app.waitlist.repository import WaitlistRepository
from app.constants.constants import (WAITLIST_CONFIRM_PROBABILITY_HIGH,WAITLIST_CONFIRM_PROBABILITY_MED, WAITLIST_CONFIRM_PROBABILITY_LOW)
from app.helpers.response import success_response, error_response

class WaitlistService:

    def __init__(self, db: Session):
        self.repo = WaitlistRepository(db)

    def get_waitlist_info(self, train_id: int, journey_date: str):
        result = self.repo.get_waitlist_count(train_id, journey_date)
        waiting_count = result.waiting_count if result else 0

        if waiting_count <= 5:
            probability = WAITLIST_CONFIRM_PROBABILITY_HIGH
            message = "High chance of confirmation!"
        elif waiting_count <= 15:
            probability = WAITLIST_CONFIRM_PROBABILITY_MED
            message = "Moderate chance of confirmation"
        else:
            probability = WAITLIST_CONFIRM_PROBABILITY_LOW
            message = "Low chance of confirmation"

        return success_response("Waitlist info fetched", {"total_waiting": waiting_count,"confirmation_probability": probability,"message": message})

    def join_waitlist(self, user_id: int, train_id: int, journey_date: str):
        self.repo.cleanup_user_waitlist(user_id, train_id, journey_date)

        existing = self.repo.find_existing(user_id, train_id, journey_date)
        if existing:
            return success_response("Already on waitlist", {"position": existing.position,"already_exists": True,"confirmation_probability": self._get_probability(existing.position),"message": f"You are already #{existing.position} on the waitlist"
            })

        self.repo.insert_waitlist(user_id, train_id, journey_date)
        new_entry = self.repo.find_existing(user_id, train_id, journey_date)
        position = new_entry.position if new_entry else 1

        return success_response("Added to waitlist", {"position": position,"already_exists": False,"confirmation_probability": self._get_probability(position),"message": f"You are #{position} on the waitlist"})

    def get_user_waitlist(self, user_id: int):
        results = self.repo.get_user_waitlist(user_id)
        data = []
        for r in results:
            data.append({
                "id": r.id,"train_id": r.train_id,"train_name": r.train_name,"user_name": r.user_name,"source": r.source,"destination": r.destination,"journey_date": str(r.journey_date),"position": r.position,"status": r.status,"probability": self._get_probability(r.position)})
        return success_response("Waitlist fetched", data)

    def confirm_next_waitlist(self, train_id: int, journey_date: str):
        result = self.repo.get_first_waiting(train_id, journey_date)
        if not result:
            return success_response("No one in waitlist", {"confirmed": False})

        seat = self.repo.get_available_seat(train_id, journey_date)
        if not seat:
            return error_response("No seats available", {"confirmed": False})
        booking = self.repo.create_booking(result.user_id, train_id, seat.id, journey_date)
        self.repo.confirm_waitlist_entry(result.id)

        return success_response("Waitlist confirmed and booking created", {
            "confirmed": True,"user_id": result.user_id,"booking_id": booking.id,"seat_number": seat.seat_number,"waitlist_id": result.id})

    def _get_probability(self, position: int) -> float:
        if position <= 5:
            return WAITLIST_CONFIRM_PROBABILITY_HIGH
        elif position <= 15:
            return WAITLIST_CONFIRM_PROBABILITY_MED
        else:
            return WAITLIST_CONFIRM_PROBABILITY_LOW