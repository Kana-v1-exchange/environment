CREATE TABLE currencies (
    currency VARCHAR(10) PRIMARY KEY,
    value FLOAT NOT NULL
);

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE,
    pass VARCHAR(255)
);

CREATE TABLE users_money(
    id INT PRIMARY KEY,
    user_id INT REFERENCES users(id) NOT NULL,
    currency VARCHAR(255),
    amount FLOAT
);



