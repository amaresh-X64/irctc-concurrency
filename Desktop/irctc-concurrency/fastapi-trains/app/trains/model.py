from sqlalchemy import Column, Integer, String, Numeric, Time, TIMESTAMP
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.sql import func

Base = declarative_base()


class Train(Base):
    __tablename__ = "trains"

    id = Column(Integer, primary_key=True, index=True)
    train_number = Column(String(10), unique=True, nullable=False)
    train_name = Column(String(100), nullable=False)
    source = Column(String(50), nullable=False)
    destination = Column(String(50), nullable=False)
    departure_time = Column(Time, nullable=False)
    arrival_time = Column(Time, nullable=False)
    total_seats = Column(Integer, default=100)
    available_seats = Column(Integer, default=100)
    price = Column(Numeric(10, 2), nullable=False)


class Seat(Base):
    __tablename__ = "seats"

    id = Column(Integer, primary_key=True, index=True)
    train_id = Column(Integer, nullable=False)
    seat_number = Column(String(10), nullable=False)
    seat_type = Column(String(20), default="GENERAL")
    is_available = Column(Integer, default=True)