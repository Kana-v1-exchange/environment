
ALTER TABLE users_money
ADD CONSTRAINT unique_user_currency UNIQUE(user_id, currency);
