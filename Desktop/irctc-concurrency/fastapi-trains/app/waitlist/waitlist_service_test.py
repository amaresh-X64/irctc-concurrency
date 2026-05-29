import pytest
from unittest.mock import MagicMock
from app.waitlist.service import WaitlistService
from app.constants.constants import (
    WAITLIST_CONFIRM_PROBABILITY_HIGH,
    WAITLIST_CONFIRM_PROBABILITY_MED,
    WAITLIST_CONFIRM_PROBABILITY_LOW
)


def make_service():
    db = MagicMock()
    return WaitlistService(db)


def make_waitlist_row(overrides={}):
    row = MagicMock()
    row.id = 1
    row.user_id = 1
    row.train_id = 1
    row.train_name = "Chennai Express"
    row.user_name = "Nivethitha"
    row.source = "Chennai"
    row.destination = "Mumbai"
    row.journey_date = "2027-06-01"
    row.position = 1
    row.status = "WAITING"
    for k, v in overrides.items():
        setattr(row, k, v)
    return row


def test_get_waitlist_info_returns_high_probability_when_few_waiting():
    service = make_service()
    service.repo.get_waitlist_count = MagicMock(return_value=MagicMock(waiting_count=5))

    result = service.get_waitlist_info(1, "2027-06-01")

    assert result["success"] is True
    assert result["data"]["confirmation_probability"] == WAITLIST_CONFIRM_PROBABILITY_HIGH
    assert result["data"]["total_waiting"] == 5


def test_get_waitlist_info_returns_medium_probability_when_moderate_waiting():
    service = make_service()
    service.repo.get_waitlist_count = MagicMock(return_value=MagicMock(waiting_count=15))
    result = service.get_waitlist_info(1, "2027-06-01")
    assert result["data"]["confirmation_probability"] == WAITLIST_CONFIRM_PROBABILITY_MED


def test_get_waitlist_info_returns_low_probability_when_many_waiting():
    service = make_service()
    service.repo.get_waitlist_count = MagicMock(return_value=MagicMock(waiting_count=20))
    result = service.get_waitlist_info(1, "2027-06-01")
    assert result["data"]["confirmation_probability"] == WAITLIST_CONFIRM_PROBABILITY_LOW


def test_get_waitlist_info_returns_zero_when_no_result():
    service = make_service()
    service.repo.get_waitlist_count = MagicMock(return_value=None)
    result = service.get_waitlist_info(1, "2027-06-01")
    assert result["data"]["total_waiting"] == 0

def test_join_waitlist_returns_already_exists_when_user_already_in_waitlist():
    service = make_service()
    service.repo.cleanup_user_waitlist = MagicMock()
    service.repo.find_existing = MagicMock(return_value=make_waitlist_row({"position": 2}))
    service.repo.insert_waitlist = MagicMock()  

    result = service.join_waitlist(1, 1, "2027-06-01")

    assert result["success"] is True
    assert result["data"]["already_exists"] is True
    assert result["data"]["position"] == 2
    service.repo.insert_waitlist.assert_not_called()

def test_join_waitlist_adds_user_when_not_in_waitlist():
    service = make_service()
    service.repo.cleanup_user_waitlist = MagicMock()
    service.repo.find_existing = MagicMock(side_effect=[None, make_waitlist_row({"position": 3})])
    service.repo.insert_waitlist = MagicMock()

    result = service.join_waitlist(1, 1, "2027-06-01")

    assert result["success"] is True
    assert result["data"]["already_exists"] is False
    assert result["data"]["position"] == 3
    service.repo.insert_waitlist.assert_called_once()

def test_get_user_waitlist_returns_list_for_user():
    service = make_service()
    service.repo.get_user_waitlist = MagicMock(return_value=[make_waitlist_row()])

    result = service.get_user_waitlist(1)

    assert result["success"] is True
    assert len(result["data"]) == 1
    assert result["data"][0]["train_name"] == "Chennai Express"


def test_get_user_waitlist_returns_empty_when_no_entries():
    service = make_service()
    service.repo.get_user_waitlist = MagicMock(return_value=[])

    result = service.get_user_waitlist(1)

    assert result["success"] is True
    assert result["data"] == []


def test_get_probability_returns_high_for_position_1():
    service = make_service()
    assert service._get_probability(1) == WAITLIST_CONFIRM_PROBABILITY_HIGH


def test_get_probability_returns_medium_for_position_10():
    service = make_service()
    assert service._get_probability(10) == WAITLIST_CONFIRM_PROBABILITY_MED


def test_get_probability_returns_low_for_position_20():
    service = make_service()
    assert service._get_probability(20) == WAITLIST_CONFIRM_PROBABILITY_LOW


def test_confirm_next_waitlist_returns_false_when_no_one_waiting():
    service = make_service()
    service.repo.confirm_next_atomically = MagicMock(return_value=None)

    result = service.confirm_next_waitlist(1, "2027-06-01")

    assert result["success"] is True
    assert result["data"]["confirmed"] is False


def test_confirm_next_waitlist_confirms_booking_when_seat_available():
    service = make_service()
    mock_row = MagicMock()
    mock_row.user_id = 5
    mock_row.booking_id = 99
    mock_row.seat_number = "S1"
    mock_row.seat_id = 3
    mock_row.waitlist_id = 1
    service.repo.confirm_next_atomically = MagicMock(return_value=mock_row)

    result = service.confirm_next_waitlist(1, "2027-06-01")

    assert result["success"] is True
    assert result["data"]["confirmed"] is True
    assert result["data"]["user_id"] == 5
    assert result["data"]["booking_id"] == 99
    assert result["data"]["seat_number"] == "S1"
    service.repo.confirm_next_atomically.assert_called_once_with(1, "2027-06-01")