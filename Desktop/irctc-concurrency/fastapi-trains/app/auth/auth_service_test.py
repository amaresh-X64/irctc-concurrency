import pytest
from unittest.mock import MagicMock, patch
from app.auth.service import AuthService
from app.auth.dto.request import RegisterRequest, LoginRequest


def make_service():
    db = MagicMock()
    return AuthService(db)


def make_register_req():
    return RegisterRequest(
        name="Nivethitha",
        email="nive@gmail.com",
        password="Password@123",
        phone="9999999999"
    )


def make_login_req():
    return LoginRequest(
        email="nive@gmail.com",
        password="Password@123"
    )


def test_register_returns_success_when_new_user_registers():
    service = make_service()
    service.repo.find_by_email = MagicMock(return_value=None)
    mock_user = MagicMock()
    mock_user.id = 1
    mock_user.name = "Nivethitha"
    mock_user.email = "nive@gmail.com"
    service.repo.create_user = MagicMock(return_value=mock_user)

    result = service.register(make_register_req())

    assert result["success"] is True
    assert "access_token" in result["data"]
    assert result["data"]["user_id"] == 1
    assert result["data"]["email"] == "nive@gmail.com"


def test_register_returns_error_when_email_already_exists():
    service = make_service()
    service.repo.find_by_email = MagicMock(return_value=MagicMock())
    service.repo.create_user = MagicMock() 
    result = service.register(make_register_req())

    assert result["success"] is False
    service.repo.create_user.assert_not_called()


def test_login_returns_token_when_valid_credentials_provided():
    service = make_service()
    mock_user = MagicMock()
    mock_user.id = 1
    mock_user.name = "Nivethitha"
    mock_user.email = "nive@gmail.com"
    mock_user.password_hash = "$2b$12$validhash"
    service.repo.find_by_email = MagicMock(return_value=mock_user)

    with patch("bcrypt.checkpw", return_value=True):
        result = service.login(make_login_req())

    assert result["success"] is True
    assert "access_token" in result["data"]
    assert result["data"]["token_type"] == "bearer"


def test_login_returns_error_when_email_not_registered():
    service = make_service()
    service.repo.find_by_email = MagicMock(return_value=None)
    result = service.login(make_login_req())
    assert result["success"] is False


def test_login_returns_error_when_password_is_wrong():
    service = make_service()
    mock_user = MagicMock()
    mock_user.password_hash = "$2b$12$validhash"
    service.repo.find_by_email = MagicMock(return_value=mock_user)

    with patch("bcrypt.checkpw", return_value=False):
        result = service.login(make_login_req())

    assert result["success"] is False


def test_generate_token_returns_valid_jwt_for_user():
    service = make_service()
    token = service._generate_token(1, "nive@gmail.com")

    assert token is not None
    assert isinstance(token, str)
    assert len(token) > 0