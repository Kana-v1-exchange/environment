package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type TransactionExecutor interface {
	Begin() error
	Commit() error
	Rollback() error
	LockMoney() error
	Exec(query string, args ...interface{}) error
	Query(query string, args ...interface{}) (pgx.Rows, error)
}

type transExec struct {
	connection *pgxpool.Conn
	tx         pgx.Tx
	isTxBegun  bool
}

func NewTransactionExecutor(connection *pgxpool.Conn) TransactionExecutor {
	return &transExec{connection, nil, false}
}

func (te *transExec) Begin() error {
	if te.isTxBegun {
		return nil
	}

	var err error
	te.tx, err = te.connection.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("cannot start transaction; err: %w", err)
	}

	te.isTxBegun = true
	return nil
}

func (te *transExec) Commit() error {
	if !te.isTxBegun {
		return nil
	}

	te.isTxBegun = false
	return te.tx.Commit(context.Background())
}

func (te *transExec) Rollback() error {
	if !te.isTxBegun {
		return nil
	}

	te.isTxBegun = false
	return te.tx.Rollback(context.Background())
}

func (te *transExec) Exec(query string, args ...interface{}) error {
	_, err := te.tx.Exec(context.Background(), query, args...)
	return err
}

func (te *transExec) Query(query string, args ...interface{}) (pgx.Rows, error) {
	return te.tx.Query(context.Background(), query, args...)
}

func (te *transExec) LockMoney() error {
	err := te.Begin()
	if err != nil {
		return err
	}

	return te.Exec("LOCK TABLE selling IN ACCESS EXCLUSIVE MODE")
}
