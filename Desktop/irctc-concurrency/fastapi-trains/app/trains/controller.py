from fastapi import APIRouter, Depends, Query, HTTPException
from pydantic import BaseModel
from sqlalchemy.orm import Session

from app.trains.service import TrainService
from app.middleware.auth import get_db, get_current_user
from app.search import elastic as es_client

router = APIRouter(prefix="/trains", tags=["Trains"], dependencies=[Depends(get_current_user)])

internal_router = APIRouter(prefix="/internal/trains", tags=["Internal"])


@router.get("/")
def get_all_trains(
    journey_date: str = Query(..., description="Journey date in YYYY-MM-DD format"),
    db: Session = Depends(get_db),
):
    return TrainService(db).get_all_trains(journey_date)


@router.get("/search")
def search_trains(
    source: str = Query(...),
    destination: str = Query(...),
    journey_date: str = Query(...),
    db: Session = Depends(get_db),
):
    return TrainService(db).search_trains(source, destination, journey_date)


@router.get("/{train_id}/seats")
def get_seats(
    train_id: int,
    journey_date: str = Query(...),
    db: Session = Depends(get_db),
):
    return TrainService(db).get_seats(train_id, journey_date)


class SeatCountUpdate(BaseModel):
    available_seats: int


@internal_router.patch("/{train_id}/seats")
def update_train_seats_in_es(train_id: int, body: SeatCountUpdate):
    client = es_client.get_es_client()
    if client is None:
        return {"updated": False, "reason": "ES unavailable"}
    es_client.update_available_seats(client, train_id, body.available_seats)
    return {"updated": True, "train_id": train_id, "available_seats": body.available_seats}