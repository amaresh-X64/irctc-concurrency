from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session
from app.waitlist.service import WaitlistService
from app.middleware.auth import get_db
from app.waitlist.dto.request import WaitlistRequest

router = APIRouter(prefix="/waitlist", tags=["Waitlist"])


@router.post("/join")
def join_waitlist(user_id: int, train_id: int, journey_date: str,db: Session = Depends(get_db)):
    service = WaitlistService(db)
    return service.join_waitlist(user_id, train_id, journey_date)

@router.get("/user/{user_id}")
def get_user_waitlist(user_id: int,db: Session = Depends(get_db)):
    service = WaitlistService(db)
    return service.get_user_waitlist(user_id)

@router.get("/{train_id}")
def get_waitlist_info(train_id: int,journey_date: str,db: Session = Depends(get_db)):
    service = WaitlistService(db)
    return service.get_waitlist_info(train_id, journey_date)

@router.post("/confirm-next")
def confirm_next_waitlist(train_id: int, journey_date: str,db: Session = Depends(get_db)):
    service = WaitlistService(db)
    return service.confirm_next_waitlist(train_id, journey_date)