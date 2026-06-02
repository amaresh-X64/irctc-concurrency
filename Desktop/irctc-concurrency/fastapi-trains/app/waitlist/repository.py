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

    def confirm_next_atomically(self, train_id: int, journey_date: str):
        result = self.db.execute(
            text("""
                WITH locked_waitlist AS (
                    SELECT id, user_id, position
                    FROM waitlist
                    WHERE train_id  = :train_id
                      AND journey_date = :journey_date
                      AND status = 'WAITING'
                    ORDER BY position ASC
                    LIMIT 1
                    FOR UPDATE SKIP LOCKED
                ),
                locked_seat AS (
                    SELECT s.id AS seat_id, s.seat_number
                    FROM seats s
                    WHERE s.train_id = :train_id
                      AND NOT EXISTS (
                          SELECT 1 FROM bookings b
                          WHERE b.seat_id    = s.id
                            AND b.journey_date = :journey_date
                            AND b.status = 'CONFIRMED'
                      )
                    LIMIT 1
                    FOR UPDATE SKIP LOCKED
                ),
                new_booking AS (
                    INSERT INTO bookings (user_id, train_id, seat_id, journey_date, status, booked_at)
                    SELECT lw.user_id, :train_id, ls.seat_id, :journey_date, 'CONFIRMED', NOW()
                    FROM locked_waitlist lw, locked_seat ls
                    RETURNING id AS booking_id, seat_id, user_id
                ),
                updated_waitlist AS (
                    UPDATE waitlist
                    SET status = 'CONFIRMED'
                    FROM locked_waitlist lw
                    WHERE waitlist.id = lw.id
                    RETURNING waitlist.id AS waitlist_id
                )
                SELECT
                    nb.booking_id,
                    nb.user_id,
                    nb.seat_id,
                    ls.seat_number,
                    uw.waitlist_id
                FROM new_booking nb
                JOIN locked_seat   ls ON ls.seat_id = nb.seat_id
                JOIN updated_waitlist uw ON TRUE
            """),
            {"train_id": train_id, "journey_date": journey_date}
        ).fetchone()

        self.db.commit()
        return result

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