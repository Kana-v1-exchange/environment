package enviroment

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v4"
)

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
}

type postgresClient struct {
	connection *pgx.Conn
}

func (ps *PostgreSettings) Connect() PostgresHandler {
	connStr := fmt.Sprintf("postgresql://%s:%s@%s/%s", ps.User, ps.Password, ps.Host, ps.DbName)

	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		panic(fmt.Errorf("cannot connect to the postgres database; err: %v", err))
	}

	err = conn.Ping(context.Background())
	if err != nil {
		panic(fmt.Errorf("cannot ping the postgres database; error: %v", err))
	}

	return &postgresClient{conn}
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
		 SET value = $1`,
		value)

	if err != nil {
		return fmt.Errorf("postgres can not update currency %v to the new value %v; err: %v", currency, value, err)
	}

	return nil
}

func (pc *postgresClient) GetUsersNum() (int, error) {
	res := 0
	err := pc.connection.QueryRow(context.Background(), "SELECT COUNT(id) FROM users").Scan(&res)

	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("cann get number of users from the postgres database; error: %v", err)
	}

	return res, nil
}
