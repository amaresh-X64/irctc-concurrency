import pytest
from unittest.mock import MagicMock
from app.waitlist.repository import WaitlistRepository


def make_repo():
    db = MagicMock()
    return WaitlistRepository(db)


def make_waitlist_row(overrides={}):
    row = MagicMock()
    row.id = 1
    row.user_id = 1
    row.train_id = 1
    row.position = 1
    row.status = "WAITING"
    row.journey_date = "2026-05-19"
    row.train_name = "Chennai Express"
    row.source = "Chennai"
    row.destination = "Mumbai"
    row.user_name = "Nivethitha"
    for k, v in overrides.items():
        setattr(row, k, v)
    return row


def test_get_waitlist_count_returns_count_row():
    repo = make_repo()
    mock_row = MagicMock(waiting_count=3)
    repo.db.execute().fetchone = MagicMock(return_value=mock_row)

    result = repo.get_waitlist_count(1, "2026-05-19")

    assert result.waiting_count == 3


def test_get_waitlist_count_returns_none_when_no_result():
    repo = make_repo()
    repo.db.execute().fetchone = MagicMock(return_value=None)

    result = repo.get_waitlist_count(1, "2026-05-19")

    assert result is None


def test_find_existing_returns_entry_when_user_in_waitlist():
    repo = make_repo()
    repo.db.execute().fetchone = MagicMock(return_value=make_waitlist_row())

    result = repo.find_existing(1, 1, "2026-05-19")

    assert result.position == 1
    assert result.status == "WAITING"


def test_find_existing_returns_none_when_user_not_in_waitlist():
    repo = make_repo()
    repo.db.execute().fetchone = MagicMock(return_value=None)

    result = repo.find_existing(99, 1, "2026-05-19")

    assert result is None


def test_insert_waitlist_executes_and_commits():
    repo = make_repo()

    repo.insert_waitlist(1, 1, "2026-05-19")

    repo.db.execute.assert_called_once()
    repo.db.commit.assert_called_once()


def test_get_user_waitlist_returns_list_of_entries():
    repo = make_repo()
    repo.db.execute().fetchall = MagicMock(return_value=[make_waitlist_row()])

    result = repo.get_user_waitlist(1)

    assert len(result) == 1
    assert result[0].train_name == "Chennai Express"


def test_get_user_waitlist_returns_empty_when_no_entries():
    repo = make_repo()
    repo.db.execute().fetchall = MagicMock(return_value=[])

    result = repo.get_user_waitlist(1)

    assert result == []


def test_get_first_waiting_returns_first_in_queue():
    repo = make_repo()
    repo.db.execute().fetchone = MagicMock(return_value=make_waitlist_row({"position": 1}))

    result = repo.get_first_waiting(1, "2026-05-19")

    assert result.position == 1


def test_get_first_waiting_returns_none_when_empty():
    repo = make_repo()
    repo.db.execute().fetchone = MagicMock(return_value=None)

    result = repo.get_first_waiting(1, "2026-05-19")

    assert result is None


def test_get_available_seat_returns_seat_when_available():
    repo = make_repo()
    seat = MagicMock(id=1, seat_number="S1")
    repo.db.execute().fetchone = MagicMock(return_value=seat)

    result = repo.get_available_seat(1, "2026-05-19")

    assert result.seat_number == "S1"


def test_get_available_seat_returns_none_when_all_booked():
    repo = make_repo()
    repo.db.execute().fetchone = MagicMock(return_value=None)

    result = repo.get_available_seat(1, "2026-05-19")

    assert result is None


def test_create_booking_returns_booking_with_id():
    repo = make_repo()
    booking = MagicMock(id=99)
    repo.db.execute().fetchone = MagicMock(return_value=booking)

    result = repo.create_booking(1, 1, 3, "2026-05-19")

    assert result.id == 99


def test_confirm_waitlist_entry_executes_and_commits():
    repo = make_repo()

    repo.confirm_waitlist_entry(1)

    repo.db.execute.assert_called_once()
    repo.db.commit.assert_called_once()


def test_cleanup_user_waitlist_executes_and_commits():
    repo = make_repo()

    repo.cleanup_user_waitlist(1, 1, "2026-05-19")

    repo.db.execute.assert_called_once()
    repo.db.commit.assert_called_once()