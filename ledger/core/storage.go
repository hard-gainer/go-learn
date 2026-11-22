package core

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"ledger/internal/db"
)

var ErrBudgetExceeded = errors.New("budget exceeded")

type Storage struct {
	conn *sql.DB
}

func NewStorage(conn *sql.DB) *Storage {
	return &Storage{conn: conn}
}

func OpenStorage() (*Storage, error) {
	conn, err := db.Open()
	if err != nil {
		return nil, err
	}
	return NewStorage(conn), nil
}

func (s *Storage) Close() error {
	return s.conn.Close()
}

func (s *Storage) SetBudget(budget Budget) error {
	if err := budget.Validate(); err != nil {
		return fmt.Errorf("invalid budget: %w", err)
	}

	tx, err := s.beginCategory(budget.Category)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	const query = `INSERT INTO budgets (category, limit_amount) VALUES ($1, $2::text::numeric(14,2))
		ON CONFLICT (category) DO UPDATE SET limit_amount = EXCLUDED.limit_amount`
	if _, err := tx.Exec(query, budget.Category, strconv.FormatFloat(budget.Limit, 'f', -1, 64)); err != nil {
		return fmt.Errorf("set budget: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit budget: %w", err)
	}
	return nil
}

func (s *Storage) ListBudgets() ([]Budget, error) {
	rows, err := s.conn.Query(`SELECT category, limit_amount FROM budgets ORDER BY category`)
	if err != nil {
		return nil, fmt.Errorf("list budgets: %w", err)
	}
	defer rows.Close()

	budgets := make([]Budget, 0)
	for rows.Next() {
		var budget Budget
		if err := rows.Scan(&budget.Category, &budget.Limit); err != nil {
			return nil, fmt.Errorf("scan budget: %w", err)
		}
		budgets = append(budgets, budget)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read budgets: %w", err)
	}
	return budgets, nil
}

func (s *Storage) AddTransaction(transaction Transaction) (Transaction, error) {
	if err := transaction.Validate(); err != nil {
		return Transaction{}, fmt.Errorf("invalid transaction: %w", err)
	}

	tx, err := s.beginCategory(transaction.Category)
	if err != nil {
		return Transaction{}, err
	}
	defer tx.Rollback()

	amount := strconv.FormatFloat(transaction.Amount, 'f', -1, 64)
	const checkBudget = `
		SELECT COALESCE((SELECT SUM(amount) FROM expenses WHERE category = $1), 0)
		       + $2::text::numeric(14,2) > limit_amount
		FROM budgets
		WHERE category = $1`
	var exceeded bool
	err = tx.QueryRow(checkBudget, transaction.Category, amount).Scan(&exceeded)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return Transaction{}, fmt.Errorf("check budget: %w", err)
	}
	if exceeded {
		return Transaction{}, ErrBudgetExceeded
	}

	const insert = `
		INSERT INTO expenses (amount, category, description, date)
		VALUES ($1::text::numeric(14,2), $2, $3, $4::text::date)
		RETURNING id, amount, date`
	err = tx.QueryRow(insert, amount, transaction.Category, transaction.Description, transaction.Date.Format(time.DateOnly)).
		Scan(&transaction.ID, &transaction.Amount, &transaction.Date)
	if err != nil {
		return Transaction{}, fmt.Errorf("insert transaction: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Transaction{}, fmt.Errorf("commit transaction: %w", err)
	}
	return transaction, nil
}

func (s *Storage) ListTransactions() ([]Transaction, error) {
	const query = `
		SELECT id, amount, category, COALESCE(description, ''), date
		FROM expenses ORDER BY date DESC, id DESC`
	rows, err := s.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("list transactions: %w", err)
	}
	defer rows.Close()

	transactions := make([]Transaction, 0)
	for rows.Next() {
		var transaction Transaction
		if err := rows.Scan(&transaction.ID, &transaction.Amount, &transaction.Category, &transaction.Description, &transaction.Date); err != nil {
			return nil, fmt.Errorf("scan transaction: %w", err)
		}
		transactions = append(transactions, transaction)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read transactions: %w", err)
	}
	return transactions, nil
}

func (s *Storage) beginCategory(category string) (*sql.Tx, error) {
	tx, err := s.conn.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	// Serialize budget changes and expense creation even when this category has no budget row yet.
	if _, err := tx.Exec(`SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, category); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("lock category: %w", err)
	}
	return tx, nil
}
