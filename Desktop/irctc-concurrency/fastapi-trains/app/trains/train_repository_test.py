import pytest
from unittest.mock import MagicMock
from app.trains.repository import TrainRepository


def make_repo():
    db = MagicMock()
    return TrainRepository(db)


def make_row(overrides={}):
    row = MagicMock()
    row.id = 1
    row.train_number = "12345"
    row.train_name = "Chennai Express"
    row.source = "Chennai"
    row.destination = "Mumbai"
    row.departure_time = "06:00"
    row.arrival_time = "22:00"
    row.total_seats = 10
    row.available_seats = 8
    row.price = 1200.0
    for k, v in overrides.items():
        setattr(row, k, v)
    return row


def make_seat_row(overrides={}):
    row = MagicMock()
    row.id = 1
    row.train_id = 1
    row.seat_number = "S1"
    row.seat_type = "SLEEPER"
    row.is_available = True
    for k, v in overrides.items():
        setattr(row, k, v)
    return row


def test_find_by_id_returns_train_when_exists():
    repo = make_repo()
    mock_train = MagicMock()
    repo.db.query().filter().first = MagicMock(return_value=mock_train)

    result = repo.find_by_id(1)

    assert result == mock_train


def test_find_by_id_returns_none_when_not_found():
    repo = make_repo()
    repo.db.query().filter().first = MagicMock(return_value=None)

    result = repo.find_by_id(99)

    assert result is None


def test_find_all_by_date_returns_list_of_dicts():
    repo = make_repo()
    repo.db.execute().fetchall = MagicMock(return_value=[make_row()])

    result = repo.find_all_by_date("2027-06-01")

    assert len(result) == 1
    assert result[0]["train_number"] == "12345"
    assert result[0]["source"] == "Chennai"


def test_find_all_by_date_returns_empty_when_no_trains():
    repo = make_repo()
    repo.db.execute().fetchall = MagicMock(return_value=[])
    result = repo.find_all_by_date("2027-06-01")
    assert result == []


def test_find_trains_by_date_returns_matching_trains():
    repo = make_repo()
    repo.db.execute().fetchall = MagicMock(return_value=[make_row()])

    result = repo.find_trains_by_date("Chennai", "Mumbai", "2027-06-01")

    assert len(result) == 1
    assert result[0]["destination"] == "Mumbai"


def test_find_trains_by_date_returns_empty_when_no_match():
    repo = make_repo()
    repo.db.execute().fetchall = MagicMock(return_value=[])

    result = repo.find_trains_by_date("Delhi", "Goa", "2027-06-01")

    assert result == []


def test_find_seats_by_date_returns_available_seats():
    repo = make_repo()
    repo.db.execute().fetchall = MagicMock(return_value=[make_seat_row()])

    result = repo.find_seats_by_date(1, "2027-06-01")

    assert len(result) == 1
    assert result[0]["seat_number"] == "S1"
    assert result[0]["is_available"] is True


def test_find_seats_by_date_returns_booked_seat_as_unavailable():
    repo = make_repo()
    repo.db.execute().fetchall = MagicMock(return_value=[make_seat_row({"is_available": False})])

    result = repo.find_seats_by_date(1, "2027-06-01")

    assert result[0]["is_available"] is False


def test_find_seats_by_date_returns_empty_when_no_seats():
    repo = make_repo()
    repo.db.execute().fetchall = MagicMock(return_value=[])

    result = repo.find_seats_by_date(1, "2027-06-01")

    assert result == []


def test_row_to_dict_clamps_negative_available_seats_to_zero():
    repo = make_repo()
    row = make_row({"available_seats": -2})

    result = repo._row_to_dict(row)

    assert result["available_seats"] == 0


def test_row_to_dict_converts_price_to_float():
    repo = make_repo()
    row = make_row({"price": 1200})

    result = repo._row_to_dict(row)

    assert isinstance(result["price"], float)