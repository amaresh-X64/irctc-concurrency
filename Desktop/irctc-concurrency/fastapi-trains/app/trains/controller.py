from fastapi import APIRouter, Depends, Query
from sqlalchemy.orm import Session

from app.trains.service import TrainService
from app.middleware.auth import get_db, get_current_user

router = APIRouter(prefix="/trains", tags=["Trains"], dependencies=[Depends(get_current_user)])


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