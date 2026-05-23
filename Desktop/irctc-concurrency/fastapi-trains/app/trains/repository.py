from sqlalchemy.orm import Session
from sqlalchemy import text
from app.trains.model import Train, Seat
from typing import List, Optional


class TrainRepository:

    def __init__(self, db: Session):
        self.db = db

    def find_by_id(self, train_id: int) -> Optional[Train]:
        return self.db.query(Train).filter(Train.id == train_id).first()

    def find_all_by_date(self, journey_date: str) -> List[dict]:
        q = """
            SELECT t.id, t.train_number, t.train_name,t.source, t.destination,t.departure_time, t.arrival_time,t.total_seats, t.price,
            (t.total_seats - COUNT(b.id)) AS available_seats
            FROM trains t
            LEFT JOIN bookings b ON b.train_id = t.id
                AND b.journey_date = :journey_date
                AND b.status = 'CONFIRMED'
            GROUP BY t.id
            ORDER BY t.id
        """
        rows = self.db.execute(text(q), {"journey_date": journey_date}).fetchall()
        return [self._row_to_dict(r) for r in rows]

    def find_trains_by_date(self, source: str, destination: str, journey_date: str) -> List[dict]:
        q = """
            SELECT t.id, t.train_number, t.train_name,t.source, t.destination,t.departure_time, t.arrival_time,t.total_seats, t.price,
            (t.total_seats - COUNT(b.id)) AS available_seats
            FROM trains t
            LEFT JOIN bookings b ON b.train_id = t.id
                AND b.journey_date = :journey_date
                AND b.status = 'CONFIRMED'
            WHERE LOWER(t.source) LIKE LOWER(:source)
            AND LOWER(t.destination) LIKE LOWER(:destination)
            GROUP BY t.id ORDER BY t.id
        """
        params = {"journey_date": journey_date, "source": f"%{source}%", "destination": f"%{destination}%"}
        rows = self.db.execute(text(q), params).fetchall()
        return [self._row_to_dict(r) for r in rows]

    def find_seats_by_date(self, train_id: int, journey_date: str):
        q = """
            SELECT s.id, s.train_id, s.seat_number, s.seat_type,
                CASE WHEN EXISTS (
                    SELECT 1 FROM bookings b
                    WHERE b.seat_id = s.id AND b.journey_date = :journey_date AND b.status = 'CONFIRMED'
                ) THEN false ELSE true END as is_available
            FROM seats s
            WHERE s.train_id = :train_id
            ORDER BY s.id
        """
        rows = self.db.execute(text(q), {"train_id": train_id, "journey_date": journey_date}).fetchall()
        return [self._seat_row_to_dict(r) for r in rows] 
    def _seat_row_to_dict(self, s) -> dict:
        return {
            "id": s.id,"train_id": s.train_id,"seat_number": s.seat_number,"seat_type": s.seat_type,"is_available": s.is_available
        }

    def _row_to_dict(self, r) -> dict:
        return {
            "id": r.id,"train_number": r.train_number,"train_name": r.train_name,"source": r.source,"destination": r.destination,"departure_time": str(r.departure_time),"arrival_time": str(r.arrival_time),"total_seats": r.total_seats,
            "available_seats": max(0, r.available_seats),"price": float(r.price)
        }