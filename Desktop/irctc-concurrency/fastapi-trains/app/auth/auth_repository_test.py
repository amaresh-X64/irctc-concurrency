import pytest
from unittest.mock import MagicMock
from app.auth.repository import AuthRepository


def make_repo():
    db = MagicMock()
    return AuthRepository(db)

def test_find_by_email_returns_user_when_email_exists():
    repo = make_repo()
    mock_user = MagicMock()
    mock_user.email = "test@example.com"
    repo.db.query.return_value.filter.return_value.first.return_value = mock_user

    result = repo.find_by_email("test@example.com")

    assert result is not None
    assert result.email == "test@example.com"


def test_find_by_email_returns_none_when_email_not_found():
    repo = make_repo()
    repo.db.query.return_value.filter.return_value.first.return_value = None

    result = repo.find_by_email("nive@gmail.com")
    assert result is None


def test_create_user_returns_user_when_valid_data_provided():
    repo = make_repo()
    mock_user = MagicMock()
    mock_user.id = 1
    mock_user.name = "Nivethitha"
    mock_user.email = "nive@gmail.com"
    mock_user.phone = "9999999999"

    result = repo.create_user(name="Nivethitha",email="nive@gmail.com", password_hash="hashedpassword",phone="9999999999")

    repo.db.add.assert_called_once()
    repo.db.commit.assert_called_once()
    repo.db.refresh.assert_called_once()


def test_create_user_calls_commit_when_user_is_saved():
    repo = make_repo()
    repo.create_user( name="Test User",email="test@example.com",password_hash="hashedpassword",phone="9999999999")
    repo.db.commit.assert_called_once()


def test_create_user_calls_add_with_user_object():
    repo = make_repo()
    repo.create_user(name="Test User",email="test@example.com", password_hash="hashedpassword",phone="9999999999")
    repo.db.add.assert_called_once()