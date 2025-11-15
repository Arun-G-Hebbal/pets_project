-- USERS TABLE
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL
);

-- OWNERS TABLE
CREATE TABLE owners (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    contact TEXT,
    email TEXT
);

-- PETS TABLE
CREATE TABLE pets (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    species TEXT,
    breed TEXT,
    owner_id INT REFERENCES owners(id) ON DELETE CASCADE,
    medical_history TEXT
);

-- APPOINTMENTS TABLE
CREATE TABLE appointments (
    id SERIAL PRIMARY KEY,
    pet_id INT REFERENCES pets(id) ON DELETE CASCADE,
    appointment_date TEXT,
    appointment_time TEXT,
    reason TEXT
);

-- FILE UPLOAD TABLE
CREATE TABLE file_records (
    id SERIAL PRIMARY KEY,
    pet_id INT REFERENCES pets(id) ON DELETE CASCADE,
    file_name TEXT NOT NULL,
    file_path TEXT NOT NULL,
    uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
