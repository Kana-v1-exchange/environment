CREATE TABLE selling (
    id SERIAL PRIMARY KEY,
    currency VARCHAR(10) NOT NULL,
    user_id INT REFERENCES users(id) NOT NULL,
    amount FLOAT NOT NULL,
    price FLOAT NOT NULL
);

ALTER TABLE selling
ADD CONSTRAINT user_currency_price_constraint UNIQUE(user_id, currency, price);
