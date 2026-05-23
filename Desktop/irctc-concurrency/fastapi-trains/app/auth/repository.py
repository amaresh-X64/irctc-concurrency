from sqlalchemy.orm import Session
from app.auth.model import User
from typing import Optional


class AuthRepository:

    def __init__(self, db: Session):
        self.db = db

    def find_by_email(self, email: str) -> Optional[User]:
        return self.db.query(User).filter(User.email == email).first()

    def create_user(self, name: str, email: str,password_hash: str, phone: str) -> User:
        user = User(name=name,email=email,password_hash=password_hash,phone=phone)
        self.db.add(user)
        self.db.commit()
        self.db.refresh(user)
        return user