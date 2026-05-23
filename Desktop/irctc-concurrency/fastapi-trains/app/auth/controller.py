from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session

from app.auth.service import AuthService
from app.auth.dto.request import RegisterRequest, LoginRequest
from app.middleware.auth import get_db

router = APIRouter(prefix="/auth", tags=["Auth"])


@router.post("/register")
def register(req: RegisterRequest, db: Session = Depends(get_db)):
    service = AuthService(db)
    return service.register(req)

@router.post("/login")
def login(req: LoginRequest, db: Session = Depends(get_db)):
    service = AuthService(db)
    return service.login(req)