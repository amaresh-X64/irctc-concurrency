from pydantic import BaseModel
from typing import Optional, List
from datetime import time


class SeatResponse(BaseModel):
    id: int
    seat_number: str
    seat_type: str
    is_available: bool

    class Config:
        from_attributes = True


class TrainResponse(BaseModel):
    id: int
    train_number: str
    train_name: str
    source: str
    destination: str
    departure_time: str
    arrival_time: str
    total_seats: int
    available_seats: int
    price: float

    class Config:
        from_attributes = True


class TrainWithSeatsResponse(BaseModel):
    train: TrainResponse
    seats: List[SeatResponse]