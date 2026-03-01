CREATE TABLE IF NOT EXISTS users (
    id CHAR(36) PRIMARY KEY,
    email VARCHAR(255) NOT NULL, 
    validated_email BOOLEAN NOT NULL, 
    username VARCHAR(255) NOT NULL, 
    hashed_password TEXT NOT NULL
);
