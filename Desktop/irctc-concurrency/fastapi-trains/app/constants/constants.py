
APP_NAME = "IRCTC Train Search Service"
APP_VERSION = "1.0.0"
API_PREFIX = "/api/v1"

DB_POOL_SIZE = 5
DB_MAX_OVERFLOW = 10

REDIS_TTL_SECONDS = 300  

MAX_WAITLIST_SIZE = 50
WAITLIST_CONFIRM_PROBABILITY_HIGH = 0.8
WAITLIST_CONFIRM_PROBABILITY_MED = 0.5
WAITLIST_CONFIRM_PROBABILITY_LOW = 0.2

STATUS_PENDING = "PENDING"
STATUS_CONFIRMED = "CONFIRMED"
STATUS_CANCELLED = "CANCELLED"
STATUS_WAITLISTED = "WAITLISTED"

MSG_TRAINS_FOUND = "Trains fetched successfully"
MSG_NO_TRAINS = "No trains found for this route"
MSG_SEATS_FOUND = "Seats fetched successfully"
MSG_WAITLIST_FETCHED = "Waitlist info fetched"
MSG_SERVER_ERROR = "Internal server error"

JWT_SECRET_KEY = "irctc_super_secret_key_2024"
JWT_ALGORITHM = "HS256"
JWT_EXPIRE_MINUTES = 60 * 24 

MSG_REGISTER_SUCCESS = "User registered successfully"
MSG_LOGIN_SUCCESS = "Login successful"
MSG_INVALID_CREDENTIALS = "Invalid email or password"
MSG_EMAIL_EXISTS = "Email already registered"
MSG_UNAUTHORIZED = "Please login to continue"
MSG_TOKEN_EXPIRED = "Session expired, please login again"
