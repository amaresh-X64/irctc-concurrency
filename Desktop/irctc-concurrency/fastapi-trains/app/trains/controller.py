from fastapi import APIRouter, Depends, Query
from sqlalchemy.orm import Session
from app.trains.service import TrainService
from app.middleware.auth import get_db

router = APIRouter(prefix="/trains", tags=["Trains"])

@router.get("/")
def get_all_trains(journey_date: str=  Query(..., description="Journey date in YYYY-MM-DD format"),db: Session = Depends(get_db)):
    service = TrainService(db)
    return service.get_all_trains(journey_date)

@router.get("/search")
def search_trains(source: str = Query(..., description="Source station"),destination: str = Query(..., description="Destination station"),journey_date: str =  Query(..., description="Journey date in YYYY-MM-DD format"),db: Session = Depends(get_db)):
    service = TrainService(db)
    return service.search_trains(source, destination, journey_date)

@router.get("/{train_id}/seats")
def get_seats(train_id: int,journey_date: str =  Query(..., description="Journey date in YYYY-MM-DD format"),db: Session = Depends(get_db)
):
    service = TrainService(db)
    return service.get_seats(train_id, journey_date)