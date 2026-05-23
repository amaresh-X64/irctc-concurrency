CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    phone VARCHAR(15),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS trains (
    id SERIAL PRIMARY KEY,
    train_number VARCHAR(10) UNIQUE NOT NULL,
    train_name VARCHAR(100) NOT NULL,
    source VARCHAR(50) NOT NULL,
    destination VARCHAR(50) NOT NULL,
    departure_time TIME NOT NULL,
    arrival_time TIME NOT NULL,
    total_seats INTEGER NOT NULL DEFAULT 100,
    available_seats INTEGER NOT NULL DEFAULT 100,
    price NUMERIC(10,2) NOT NULL
);

CREATE TABLE IF NOT EXISTS seats (
    id SERIAL PRIMARY KEY,
    train_id INTEGER REFERENCES trains(id) ON DELETE CASCADE,
    seat_number VARCHAR(10) NOT NULL,
    seat_type VARCHAR(20) DEFAULT 'GENERAL',
    is_available BOOLEAN DEFAULT TRUE,
    UNIQUE(train_id, seat_number)
);

CREATE TABLE IF NOT EXISTS bookings (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    train_id INTEGER REFERENCES trains(id),
    seat_id INTEGER REFERENCES seats(id),
    journey_date DATE NOT NULL,
    status VARCHAR(20) DEFAULT 'PENDING',
    booked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT booking_status CHECK (status IN ('PENDING','CONFIRMED','CANCELLED','WAITLISTED'))
);

CREATE TABLE IF NOT EXISTS payments (
    id SERIAL PRIMARY KEY,
    booking_id INTEGER REFERENCES bookings(id),
    user_id INTEGER REFERENCES users(id),
    amount NUMERIC(10,2) NOT NULL,
    status VARCHAR(20) DEFAULT 'PENDING',
    payment_method VARCHAR(30) DEFAULT 'MOCK_UPI',
    transaction_id VARCHAR(50) UNIQUE,
    paid_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT payment_status CHECK (status IN ('PENDING','SUCCESS','FAILED','REFUNDED'))
);

CREATE TABLE IF NOT EXISTS pnr_records (
    id SERIAL PRIMARY KEY,
    pnr_number VARCHAR(20) UNIQUE NOT NULL,
    booking_id INTEGER REFERENCES bookings(id),
    payment_id INTEGER REFERENCES payments(id),
    user_id INTEGER REFERENCES users(id),
    train_id INTEGER REFERENCES trains(id),
    seat_number VARCHAR(10),
    journey_date DATE,
    status VARCHAR(20) DEFAULT 'CONFIRMED',
    generated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS waitlist (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    train_id INTEGER REFERENCES trains(id),
    journey_date DATE NOT NULL,
    position INTEGER NOT NULL,
    status VARCHAR(20) DEFAULT 'WAITING',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT waitlist_status CHECK (status IN ('WAITING','CONFIRMED','CANCELLED'))
);

INSERT INTO trains (train_number, train_name, source, destination, departure_time, arrival_time, total_seats, available_seats, price) VALUES
('12345','Chennai Express','Chennai','Mumbai','06:00','22:00', 10, 10, 1200),
('12346','Rajdhani Express','Delhi','Mumbai','08:00', '20:00', 10, 10, 1500),
('12347','Shatabdi Express','Bangalore','Chennai','07:00', '11:00', 10, 10, 800),
('12348','Duronto Express','Kolkata','Delhi','10:00', '06:00', 10, 10, 1800),
('12349','Vande Bharat Express','Chennai','Bangalore', '06:00', '10:30', 10, 10, 950),
('12350','Garib Rath Express','Hyderabad','Chennai', '05:30', '16:00', 10, 10, 1100),
('12351','Tejas Express','Mumbai','Goa','06:00', '14:30', 10, 10, 1400),
('12352','Intercity Express','Pune','Mumbai','07:15', '11:30', 10, 10, 650),
('12353','Konkan Express','Mangalore','Mumbai','09:00', '23:45', 10, 10, 1600),
('12354','Deccan Queen','Mumbai','Pune', '17:00','20:30', 10, 10, 700),
('12355','Mysore Express','Mysore','Chennai','21:00', '06:00', 10, 10, 1000),
('12356','Coromandel Express','Chennai','Kolkata','08:30', '10:00', 10, 10, 2200),
('12357','Humsafar Express','Delhi','Bangalore','18:00', '14:00', 10, 10, 2500),
('12358','Jan Shatabdi','Coimbatore','Chennai','06:45', '13:00', 10, 10, 850),
('12359','Udyan Express','Bangalore','Mumbai','20:00', '13:00', 10, 10, 1700);

INSERT INTO seats (train_id, seat_number, seat_type) VALUES
(1,'1A','FIRST_AC'),(1,'1B','FIRST_AC'),(1,'2A','SECOND_AC'),
(1,'2B','SECOND_AC'),(1,'3A','THIRD_AC'),(1,'3B','THIRD_AC'),
(1,'S1','SLEEPER'),(1,'S2','SLEEPER'),(1,'G1','GENERAL'),(1,'G2','GENERAL'),

(2,'1A','FIRST_AC'),(2,'1B','FIRST_AC'),(2,'2A','SECOND_AC'),
(2,'2B','SECOND_AC'),(2,'3A','THIRD_AC'),(2,'3B','THIRD_AC'),
(2,'S1','SLEEPER'),(2,'S2','SLEEPER'),(2,'G1','GENERAL'),(2,'G2','GENERAL'),

(3,'1A','FIRST_AC'),(3,'1B','FIRST_AC'),(3,'2A','SECOND_AC'),
(3,'2B','SECOND_AC'),(3,'3A','THIRD_AC'),(3,'3B','THIRD_AC'),
(3,'S1','SLEEPER'),(3,'S2','SLEEPER'),(3,'G1','GENERAL'),(3,'G2','GENERAL'),

(4,'1A','FIRST_AC'),(4,'1B','FIRST_AC'),(4,'2A','SECOND_AC'),
(4,'2B','SECOND_AC'),(4,'3A','THIRD_AC'),(4,'3B','THIRD_AC'),
(4,'S1','SLEEPER'),(4,'S2','SLEEPER'),(4,'G1','GENERAL'),(4,'G2','GENERAL'),

(5,'1A','FIRST_AC'),(5,'1B','FIRST_AC'),(5,'2A','SECOND_AC'),
(5,'2B','SECOND_AC'),(5,'3A','THIRD_AC'),(5,'3B','THIRD_AC'),
(5,'S1','SLEEPER'),(5,'S2','SLEEPER'),(5,'G1','GENERAL'),(5,'G2','GENERAL'),

(6,'1A','FIRST_AC'),(6,'1B','FIRST_AC'),(6,'2A','SECOND_AC'),
(6,'2B','SECOND_AC'),(6,'3A','THIRD_AC'),(6,'3B','THIRD_AC'),
(6,'S1','SLEEPER'),(6,'S2','SLEEPER'),(6,'G1','GENERAL'),(6,'G2','GENERAL'),

(7,'1A','FIRST_AC'),(7,'1B','FIRST_AC'),(7,'2A','SECOND_AC'),
(7,'2B','SECOND_AC'),(7,'3A','THIRD_AC'),(7,'3B','THIRD_AC'),
(7,'S1','SLEEPER'),(7,'S2','SLEEPER'),(7,'G1','GENERAL'),(7,'G2','GENERAL'),

(8,'1A','FIRST_AC'),(8,'1B','FIRST_AC'),(8,'2A','SECOND_AC'),
(8,'2B','SECOND_AC'),(8,'3A','THIRD_AC'),(8,'3B','THIRD_AC'),
(8,'S1','SLEEPER'),(8,'S2','SLEEPER'),(8,'G1','GENERAL'),(8,'G2','GENERAL'),

(9,'1A','FIRST_AC'),(9,'1B','FIRST_AC'),(9,'2A','SECOND_AC'),
(9,'2B','SECOND_AC'),(9,'3A','THIRD_AC'),(9,'3B','THIRD_AC'),
(9,'S1','SLEEPER'),(9,'S2','SLEEPER'),(9,'G1','GENERAL'),(9,'G2','GENERAL'),

(10,'1A','FIRST_AC'),(10,'1B','FIRST_AC'),(10,'2A','SECOND_AC'),
(10,'2B','SECOND_AC'),(10,'3A','THIRD_AC'),(10,'3B','THIRD_AC'),
(10,'S1','SLEEPER'),(10,'S2','SLEEPER'),(10,'G1','GENERAL'),(10,'G2','GENERAL'),

(11,'1A','FIRST_AC'),(11,'1B','FIRST_AC'),(11,'2A','SECOND_AC'),
(11,'2B','SECOND_AC'),(11,'3A','THIRD_AC'),(11,'3B','THIRD_AC'),
(11,'S1','SLEEPER'),(11,'S2','SLEEPER'),(11,'G1','GENERAL'),(11,'G2','GENERAL'),

(12,'1A','FIRST_AC'),(12,'1B','FIRST_AC'),(12,'2A','SECOND_AC'),
(12,'2B','SECOND_AC'),(12,'3A','THIRD_AC'),(12,'3B','THIRD_AC'),
(12,'S1','SLEEPER'),(12,'S2','SLEEPER'),(12,'G1','GENERAL'),(12,'G2','GENERAL'),
(13,'1A','FIRST_AC'),(13,'1B','FIRST_AC'),(13,'2A','SECOND_AC'),
(13,'2B','SECOND_AC'),(13,'3A','THIRD_AC'),(13,'3B','THIRD_AC'),
(13,'S1','SLEEPER'),(13,'S2','SLEEPER'),(13,'G1','GENERAL'),(13,'G2','GENERAL'),

(14,'1A','FIRST_AC'),(14,'1B','FIRST_AC'),(14,'2A','SECOND_AC'),
(14,'2B','SECOND_AC'),(14,'3A','THIRD_AC'),(14,'3B','THIRD_AC'),
(14,'S1','SLEEPER'),(14,'S2','SLEEPER'),(14,'G1','GENERAL'),(14,'G2','GENERAL'),

(15,'1A','FIRST_AC'),(15,'1B','FIRST_AC'),(15,'2A','SECOND_AC'),
(15,'2B','SECOND_AC'),(15,'3A','THIRD_AC'),(15,'3B','THIRD_AC'),
(15,'S1','SLEEPER'),(15,'S2','SLEEPER'),(15,'G1','GENERAL'),(15,'G2','GENERAL');

