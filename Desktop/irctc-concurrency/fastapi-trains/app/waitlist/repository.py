from sqlalchemy.orm import Session
from sqlalchemy import text


class WaitlistRepository:

    def __init__(self, db: Session):
        self.db = db

    def get_waitlist_count(self, train_id: int, journey_date: str):
        return self.db.execute(
            text("""
                SELECT COUNT(*) as waiting_count
                FROM waitlist
                WHERE train_id = :train_id
                AND journey_date = :journey_date
                AND status = 'WAITING'
            """),
            {"train_id": train_id, "journey_date": journey_date}
        ).fetchone()

    def find_existing(self, user_id: int, train_id: int, journey_date: str):
        return self.db.execute(
            text("""
                SELECT id, position, status FROM waitlist
                WHERE user_id = :user_id
                AND train_id = :train_id
                AND journey_date = :journey_date
                AND status = 'WAITING'
                ORDER BY created_at DESC
                LIMIT 1
            """),
            {"user_id": user_id, "train_id": train_id, "journey_date": journey_date}
        ).fetchone()

    def insert_waitlist(self, user_id: int, train_id: int, journey_date: str):
        self.db.execute(
            text("""
                INSERT INTO waitlist (user_id, train_id, journey_date, position, status)
                VALUES (:user_id, :train_id, :journey_date, 
                    (SELECT COALESCE(MAX(position), 0) + 1 
                    FROM waitlist 
                    WHERE train_id = :train_id 
                    AND journey_date = :journey_date 
                    AND status = 'WAITING'),'WAITING')
            """),
            {"user_id": user_id, "train_id": train_id, "journey_date": journey_date}
        )
        self.db.commit()

    def get_user_waitlist(self, user_id: int):
        return self.db.execute(
            text("""
                SELECT w.id, w.train_id, w.journey_date,
                    w.position, w.status, w.created_at,
                    t.train_name, t.source, t.destination,
                    u.name as user_name
                FROM waitlist w
                JOIN trains t ON t.id = w.train_id
                JOIN users u ON u.id = w.user_id
                WHERE w.user_id = :user_id
                AND w.status = 'WAITING'
                ORDER BY w.position ASC
            """),
            {"user_id": user_id}
        ).fetchall()

    def get_first_waiting(self, train_id: int, journey_date: str):
        return self.db.execute(
            text("""
                SELECT id, user_id, position
                FROM waitlist
                WHERE train_id = :train_id
                AND journey_date = :journey_date
                AND status = 'WAITING'
                ORDER BY position ASC
                LIMIT 1
            """),
            {"train_id": train_id, "journey_date": journey_date}
        ).fetchone()

    def get_available_seat(self, train_id: int, journey_date: str):
        return self.db.execute(
            text("""
                SELECT s.id, s.seat_number FROM seats s
                WHERE s.train_id = :train_id
                AND NOT EXISTS (
                    SELECT 1 FROM bookings b
                    WHERE b.seat_id = s.id
                    AND b.journey_date = :journey_date
                    AND b.status = 'CONFIRMED'
                )
                LIMIT 1
            """),
            {"train_id": train_id, "journey_date": journey_date}
        ).fetchone()

    def create_booking(self, user_id: int, train_id: int, seat_id: int, journey_date: str):
        return self.db.execute(
            text("""
                INSERT INTO bookings (user_id, train_id, seat_id, journey_date, status, booked_at)
                VALUES (:user_id, :train_id, :seat_id, :journey_date, 'CONFIRMED', NOW())
                RETURNING id
            """),
            {"user_id": user_id, "train_id": train_id,
             "seat_id": seat_id, "journey_date": journey_date}
        ).fetchone()

    def confirm_waitlist_entry(self, waitlist_id: int):
        self.db.execute(
            text("UPDATE waitlist SET status = 'CONFIRMED' WHERE id = :id"),
            {"id": waitlist_id}
        )
        self.db.commit()

    def cleanup_user_waitlist(self, user_id: int, train_id: int, journey_date: str):
        self.db.execute(
            text("""
                DELETE FROM waitlist
                WHERE user_id = :user_id
                AND train_id = :train_id
                AND journey_date = :journey_date
                AND status = 'CONFIRMED'
            """),
            {"user_id": user_id, "train_id": train_id, "journey_date": journey_date}
        )
        self.db.commit()