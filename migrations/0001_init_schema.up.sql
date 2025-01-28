-- +goose Up
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    balance DECIMAL(10, 2) NOT NULL DEFAULT 0.00
);

CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    amount DECIMAL(10, 2) NOT NULL,
    type VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);