CREATE OR REPLACE FUNCTION give_start_money()
    RETURNS trigger AS 
    $$
    BEGIN 
        INSERT INTO users_money (user_id, currency, amount) 
        VALUES(NEW.id, 'USD', 1000);
        RETURN NEW;
END;
$$
LANGUAGE 'plpgsql';

CREATE OR REPLACE TRIGGER give_money_to_users
AFTER INSERT
ON users
FOR EACH ROW
EXECUTE PROCEDURE give_start_money();

