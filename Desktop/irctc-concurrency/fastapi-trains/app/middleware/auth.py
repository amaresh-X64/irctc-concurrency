import os
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker, Session
from typing import Generator

from fastapi import Depends, HTTPException, status
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from jose import jwt, JWTError
import app.constants.constants as const

DATABASE_URL = os.getenv(
    "DATABASE_URL",
    "postgresql://irctc_user:irctc_pass@localhost:5432/irctc_db"
)

engine = create_engine(DATABASE_URL)
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)

def get_db() -> Generator[Session, None, None]:
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()


_bearer = HTTPBearer()

def get_current_user(
    credentials: HTTPAuthorizationCredentials = Depends(_bearer),
) -> dict:
    token = credentials.credentials
    try:
        payload = jwt.decode(
            token,
            const.JWT_SECRET_KEY,
            algorithms=[const.JWT_ALGORITHM],  # pins algorithm, blocks alg:none
        )
        return payload
    except JWTError:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail=const.MSG_TOKEN_EXPIRED,
            headers={"WWW-Authenticate": "Bearer"},
        )