from sqlalchemy.orm import Session
from jose import jwt
import bcrypt
from datetime import datetime, timedelta

from app.auth.repository import AuthRepository
from app.auth.dto.request import RegisterRequest, LoginRequest
from app.helpers.response import success_response, error_response
import app.constants.constants  as const


class AuthService:

    def __init__(self, db: Session):
        self.repo = AuthRepository(db)

    def register(self, req: RegisterRequest):
        existing = self.repo.find_by_email(req.email)
        if existing:
            return error_response(const.MSG_EMAIL_EXISTS, None)

        hashed = bcrypt.hashpw(req.password.encode("utf-8"),bcrypt.gensalt()).decode("utf-8")
        user = self.repo.create_user(name=req.name,email=req.email,password_hash=hashed,phone=req.phone)
        token = self._generate_token(user.id, user.email)

        return success_response(const.MSG_REGISTER_SUCCESS, {"access_token": token,"token_type":"bearer","user_id":user.id,"name":user.name,"email":user.email})

    def login(self, req: LoginRequest):
        user = self.repo.find_by_email(req.email)
        if not user:
            return error_response(const.MSG_INVALID_CREDENTIALS, None)

        is_valid = bcrypt.checkpw(req.password.encode("utf-8"),user.password_hash.encode("utf-8"))
        if not is_valid:
            return error_response(const.MSG_INVALID_CREDENTIALS, None)

        token = self._generate_token(user.id, user.email)
        return success_response(const.MSG_LOGIN_SUCCESS, {"access_token": token,"token_type":"bearer","user_id":user.id, "name":user.name,
        "email":user.email})
        
    def _generate_token(self, user_id: int, email: str) -> str:
        expire = datetime.utcnow() + timedelta(minutes=const.JWT_EXPIRE_MINUTES)
        payload = {
            "sub":   str(user_id),
            "email": email,
            "exp":   expire,
        }
        return jwt.encode(payload, const.JWT_SECRET_KEY, algorithm=const.JWT_ALGORITHM)