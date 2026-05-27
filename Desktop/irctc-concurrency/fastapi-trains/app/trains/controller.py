from fastapi import APIRouter, Depends, Query, HTTPException
from pydantic import BaseModel
from sqlalchemy.orm import Session

from app.trains.service import TrainService
from app.middleware.auth import get_db, get_current_user
from app.search import elastic as es_client

router = APIRouter(prefix="/trains", tags=["Trains"], dependencies=[Depends(get_current_user)])

# Internal router — no JWT required; called only by peer services inside Docker network
internal_router = APIRouter(prefix="/internal/trains", tags=["Internal"])


# ── Public endpoints ─────────────────────────────────────────────────────────

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


# ── Internal endpoint ─────────────────────────────────────────────────────────

class SeatCountUpdate(BaseModel):
    available_seats: int


@internal_router.patch("/{train_id}/seats")
def update_train_seats_in_es(train_id: int, body: SeatCountUpdate):
    """
    Called by gin-booking after every successful booking or cancellation.
    Refreshes the available_seats field in the Elasticsearch document so
    that subsequent searches reflect the latest seat count.
    """
    client = es_client.get_es_client()
    if client is None:
        # ES is down — not a hard error, just a stale read risk
        return {"updated": False, "reason": "ES unavailable"}
    es_client.update_available_seats(client, train_id, body.available_seats)
    return {"updated": True, "train_id": train_id, "available_seats": body.available_seats}