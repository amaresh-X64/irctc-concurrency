import pytest
from unittest.mock import MagicMock
from app.trains.service import TrainService


def make_service():
    db = MagicMock()
    return TrainService(db)


def make_train():
    return {
        "id": 1,"train_number": "12345","train_name": "Chennai Express","source": "Chennai","destination": "Mumbai", "departure_time": "06:00","arrival_time": "22:00","total_seats": 10,"available_seats": 8,"price": 1200.0
    }


def make_seat():
    return {
        "id": 1,"train_id": 1,"seat_number": "S1","seat_type": "SLEEPER","is_available": True
    }


def test_search_trains_returns_success_when_trains_found():
    service = make_service()
    service.repo.find_trains_by_date = MagicMock(return_value=[make_train()])

    result = service.search_trains("Chennai", "Mumbai", "2027-06-01")

    assert result["success"] is True
    assert len(result["data"]) == 1
    assert result["data"][0]["train_number"] == "12345"


def test_search_trains_returns_no_trains_message_when_empty():
    service = make_service()
    service.repo.find_trains_by_date = MagicMock(return_value=[])

    result = service.search_trains("Chennai", "Mumbai", "2027-06-01")

    assert result["success"] is True
    assert result["data"] == []


def test_get_all_trains_returns_all_trains_for_date():
    service = make_service()
    service.repo.find_all_by_date = MagicMock(return_value=[make_train()])

    result = service.get_all_trains("2027-06-01")

    assert result["success"] is True
    assert len(result["data"]) == 1


def test_get_all_trains_returns_empty_when_no_trains():
    service = make_service()
    service.repo.find_all_by_date = MagicMock(return_value=[])

    result = service.get_all_trains("2027-06-01")

    assert result["success"] is True
    assert result["data"] == []


def test_get_seats_returns_seats_when_train_exists():
    service = make_service()
    service.repo.find_by_id = MagicMock(return_value=make_train())
    service.repo.find_seats_by_date = MagicMock(return_value=[make_seat()])

    result = service.get_seats(1, "2027-06-01")

    assert result["success"] is True
    assert len(result["data"]) == 1
    assert result["data"][0]["seat_number"] == "S1"


def test_get_seats_returns_error_when_train_not_found():
    service = make_service()
    service.repo.find_by_id = MagicMock(return_value=None)
    service.repo.find_seats_by_date = MagicMock()

    result = service.get_seats(99, "2027-06-01")

    assert result["success"] is False
    service.repo.find_seats_by_date.assert_not_called()


def test_get_seats_returns_empty_when_all_seats_booked():
    service = make_service()
    service.repo.find_by_id = MagicMock(return_value=make_train())
    service.repo.find_seats_by_date = MagicMock(return_value=[])

    result = service.get_seats(1, "2027-06-01")

    assert result["success"] is True
    assert result["data"] == []