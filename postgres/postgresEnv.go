package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type SellingInfo struct {
	UserID   uint64
	Amount   float64
	Currency string
	Price    float64
}

type PostgreSettings struct {
	User     string
	Password string
	Host     string
	Port     string
	DbName   string
}

type PostgresHandler interface {
	GetCurrencies() (map[string]float64, error)
	GetUsersNum() (int, error)
	UpdateCurrency(currency string, value float64) error
	GetCurrencyAmount(currency string) (float64, error)
	GetCurrencyValue(currency string) (float64, error)
	UpdateCurrencyAmount(userID uint64, currency string, value float64) error
	AddUser(email, password string) error
	GetUserData(email string) (uint64, string, error)
	GetUserMoney(userID uint64, currency string) (float64, error)
	FindSellers(tx TransactionExecutor, currency string, value float64, floorPrice, ceilPrice float64) ([]*SellingInfo, error)
	AddMoneyToSellingPool(tx TransactionExecutor, currency string, userID uint64, amount, price float64) error
	GetMoneyFromSellingPool(tx TransactionExecutor, currency string, userID uint64, amount, floorPrice, ceilPrice float64) error
	SendMoney(tx TransactionExecutor, senderID, receiverID uint64, currency string, value float64) error
}

type postgresClient struct {
	connection *pgxpool.Conn
}

func (ps *PostgreSettings) Connect() (PostgresHandler, TransactionExecutor) {
	connStr := fmt.Sprintf("postgresql://%s:%s@%s/%s?pool_min_conns=2&prefer_simple_protocol=true", ps.User, ps.Password, ps.Host, ps.DbName)

	pool, err := pgxpool.Connect(context.Background(), connStr)
	if err != nil {
		panic(fmt.Errorf("cannot connect to the postgres database; err: %v", err))
	}

	pgConn, err := pool.Acquire(context.Background())
	if err != nil {
		panic("cannot get connection from the pool")
	}

	err = pgConn.Ping(context.Background())
	if err != nil {
		panic(fmt.Errorf("cannot ping the postgres database; error: %v", err))
	}

	tranConn, err := pool.Acquire(context.Background())
	if err != nil {
		panic("cannot get connection from the pool")
	}

	err = tranConn.Ping(context.Background())
	if err != nil {
		panic(fmt.Errorf("cannot ping the postgres database; error: %v", err))
	}

	return &postgresClient{pgConn}, NewTransactionExecutor(tranConn)
}

func (pc *postgresClient) GetCurrencies() (map[string]float64, error) {
	res := make(map[string]float64)

	rows, err := pc.connection.Query(context.Background(), "SELECT * FROM currencies")
	if err != nil {
		return nil, fmt.Errorf("cannot get currencies from the postgres database; err: %v", err)
	}

	for rows.Next() {
		var currency string
		var value float64
		err = rows.Scan(&currency, &value)

		if err != nil {
			return nil, fmt.Errorf("cannot scan value from the postgres database; err: %v", err)
		}

		res[currency] = value
	}

	return res, nil
}

func (pc *postgresClient) UpdateCurrency(currency string, value float64) error {
	_, err := pc.connection.Exec(context.Background(),
		`UPDATE currencies
		 SET value = $1
		 WHERE currency = $2`,
		value,
		currency)

	if err != nil {
		return fmt.Errorf("postgres can not update currency %v to the new value %v; err: %v", currency, value, err)
	}

	return nil
}

func (pc *postgresClient) GetUsersNum() (int, error) {
	res := 0
	err := pc.connection.QueryRow(context.Background(), "SELECT COUNT(id) FROM users").Scan(&res)

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return 0, fmt.Errorf("cann get number of users from the postgres database; error: %v", err)
	}

	return res, nil
}

func (pc *postgresClient) GetCurrencyAmount(currency string) (float64, error) {
	amount := float64(0)
	err := pc.connection.QueryRow(
		context.Background(),
		`SELECT SUM(amount)
		 FROM users_money
		 WHERE currency = $1`,
		currency,
	).Scan(&amount)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, err
		}

		return 0, fmt.Errorf("postgres cannot return amount of the currency %v; err: %v", currency, err)
	}

	return amount, nil
}

func (pc *postgresClient) GetCurrencyValue(currency string) (float64, error) {
	row := pc.connection.QueryRow(
		context.Background(),
		`SELECT value 
		 FROM currencies 
		 WHERE currency = $1`,
		currency,
	)

	value := float64(0)
	err := row.Scan(&value)
	if err != nil {
		return 0, fmt.Errorf("cannot get currencies'(%v) value; err: %v", currency, err)
	}

	return value, nil
}

func (pc *postgresClient) UpdateCurrencyAmount(userID uint64, currency string, value float64) error {
	_, err := pc.connection.Exec(
		context.Background(),
		`
		 INSERT INTO users_money (amount, user_id, currency)
		 VALUES($1, $2, $3)
		 ON CONFLICT (user_id, currency)
		 DO UPDATE 
		 SET amount = EXCLUDED.amount`,
		value,
		userID,
		currency,
	)

	if err != nil {
		return fmt.Errorf("cannot update user's (id = %v) currency (%v); err: %v", userID, currency, err)
	}

	return nil
}

func (pc *postgresClient) AddUser(email, password string) error {
	_, err := pc.connection.Exec(
		context.Background(),
		`INSERT INTO users (email, pass)
		 VALUES($1, $2)`,
		email,
		password,
	)

	if err != nil {
		return fmt.Errorf("cannot update user's (email: %v, password: %v) data; err: %v", email, password, err)
	}

	return nil
}

func (pc *postgresClient) GetUserData(email string) (uint64, string, error) {
	id := uint64(0)
	password := ""

	row := pc.connection.QueryRow(
		context.Background(),
		`SELECT id, pass 
		 FROM users 
		 WHERE email = $1`,
		email,
	)

	err := row.Scan(&id, &password)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, "", err
		}

		return 0, "", fmt.Errorf("postgres cannot return user's data (email = %v); err: %v", email, err)
	}

	return id, password, nil
}

func (pc *postgresClient) GetUserMoney(userID uint64, currency string) (float64, error) {
	rows := pc.connection.QueryRow(
		context.Background(),
		`SELECT amount 
		 FROM users_money
		 WHERE user_id = $1
		 AND currency = $2`,
		userID,
		currency,
	)

	amount := float64(0)

	err := rows.Scan(&amount)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, err
		}

		return 0, fmt.Errorf("postgres cannot scan user's (id = %v) amount of the currency (%v); err: %v", userID, currency, err)
	}

	return amount, nil
}

func (pc *postgresClient) FindSellers(tx TransactionExecutor, currency string, amountToBuy float64, floorPrice, ceilPrice float64) ([]*SellingInfo, error) {
	rows, err := tx.Query(
		`SELECT users_money.user_id, selling.amount, selling.price
		 FROM users_money 
		 	JOIN selling 
			ON selling.user_id = users_money.user_id
		 WHERE users_money.currency = $1
		 AND selling.price BETWEEN $2 AND $3
		 ORDER BY selling.price`,
		currency,
		floorPrice,
		ceilPrice,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		tx.Rollback()
		return nil, fmt.Errorf("postgres cannot find sellers for the currency %v (amount %v); err: %w", currency, amountToBuy, err)
	}

	sellers := make([]*SellingInfo, 0)
	sum := float64(0)
	skip := false

	for rows.Next() {
		if skip {
			continue
		}

		sellerID := uint64(0)
		sellerMoneyAmount := float64(0)
		price := float64(0)

		err = rows.Scan(&sellerID, &sellerMoneyAmount, &price)
		if err != nil {
			return nil, fmt.Errorf("pgx cannot scan userID or users_money.amount; err: %w", err)
		}

		sum += sellerMoneyAmount

		if sum >= amountToBuy {
			sellers = append(sellers, &SellingInfo{
				UserID:   sellerID,
				Amount:   sellerMoneyAmount - (sum - amountToBuy),
				Price:    price,
				Currency: currency,
			})

			skip = true
			continue
		}

		sellers = append(sellers, &SellingInfo{
			UserID:   sellerID,
			Amount:   sellerMoneyAmount,
			Price:    price,
			Currency: currency,
		})
	}

	if sum < amountToBuy {
		return nil, nil
	}

	return sellers, nil
}

func (pc *postgresClient) SendMoney(tx TransactionExecutor, senderID, receiverID uint64, currency string, value float64) error {
	userMoney := float64(0)

	rows, err := tx.Query(
		`SELECT amount 
		 FROM users_money 
		 WHERE currency = $1
		 AND user_id = $2`,
		currency,
		senderID,
	)

	if err != nil {
		tx.Rollback()

		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%w; user with id %v does not have %v %v", pgx.ErrNoRows, senderID, value, currency)
		}

		return fmt.Errorf("cannot get %v %v from the users_money table; err: %v", value, currency, err)
	}

	for rows.Next() {
		err = rows.Scan(&userMoney)
		if err != nil {
			return err
		}
	}

	err = tx.Exec(
		`UPDATE users_money
		 SET amount = $1
		 WHERE user_id = $2
		 AND currency = $3`,
		userMoney-value,
		senderID,
		currency,
	)

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("cannot sell user's (id = %v) currency(%s); err: %v", senderID, currency, err)
	}

	err = tx.Exec(
		`
		 INSERT into users_money (user_id, currency, amount)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (user_id, currency)
		 DO UPDATE 
		 SET amount = users_money.amount + EXCLUDED.amount`,
		receiverID,
		currency,
		value,
	)

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("cannot update currency amount; err: %v", err)
	}

	return nil
}

func (pc *postgresClient) AddMoneyToSellingPool(tx TransactionExecutor, currency string, userID uint64, amount, price float64) error {
	err := tx.Exec(
		`INSERT INTO selling (currency, user_id, amount, price)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, currency, price) 
		DO UPDATE 
		SET amount = selling.amount + EXCLUDED.amount`,
		currency,
		userID,
		amount,
		price)

	if err != nil {
		tx.Rollback()
		return err
	}

	rows, err := tx.Query(
		`SELECT amount 
		 FROM users_money 
		 WHERE currency = $1
		 AND user_id = $2`,
		currency,
		userID,
	)

	if err != nil {
		tx.Rollback()

		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%w; user with id %v does not have %v %v", pgx.ErrNoRows, userID, amount, currency)
		}

		return fmt.Errorf("cannot get %v %v from the users_money table; err: %v", amount, currency, err)
	}

	userHas := float64(0)
	for rows.Next() {
		err = rows.Scan(&userHas)
		if err != nil {
			return err
		}
	}

	err = tx.Exec(
		`UPDATE users_money
	 	 SET amount = $1
	 	 WHERE user_id = $2
	 	 AND currency = $3`,
		userHas-amount,
		userID,
		currency,
	)

	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (pc *postgresClient) GetMoneyFromSellingPool(tx TransactionExecutor, currency string, userID uint64, amount, floorPrice, ceilPrice float64) error {
	err := tx.Exec(
		`CREATE VIEW buf AS 
		 SELECT id, amount 
		 FROM selling 
		 WHERE user_id = $1
		 AND currency = $2 
		 AND price BETWEEN $4 AND $5 
		 ORDER BY price
		 LIMIT 1;

		 UPDATE selling
		 SET amount = (SELECT amount FROM buf) - $3
		 WHERE id = (SELECT id FROM buf);
		 
		 DROP VIEW buf;`,
		userID,
		currency,
		amount,
		floorPrice,
		ceilPrice)

	if err != nil {
		return err
	}

	err = tx.Exec(
		`INSERT INTO users_money (user_id, currency, amount)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (user_id, currency)
		 DO UPDATE 
		 SET amount = users_money.amount + EXCLUDED.amount`,
		userID,
		currency,
		amount,
	)

	if err != nil {
		return err
	}

	return nil
}
