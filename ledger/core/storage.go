package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"slices"
	"sync"
)

var ErrBudgetExceeded = errors.New("budget exceeded")

type Storage struct {
	mu           sync.Mutex
	transactions []Transaction
	budgets      map[string]Budget
	lastID       int
}

func NewStorage() *Storage {
	return &Storage{budgets: make(map[string]Budget)}
}

// Reset clears transactions, budgets, and the transaction ID counter.
func (s *Storage) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.transactions = nil
	clear(s.budgets)
	s.lastID = 0
}

func (s *Storage) SetBudget(budget Budget) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.budgets[budget.Category] = budget
}

func (s *Storage) LoadBudgets(r io.Reader) error {
	var budgets []Budget
	if err := json.NewDecoder(r).Decode(&budgets); err != nil {
		return fmt.Errorf("error decoding json file: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, budget := range budgets {
		s.budgets[budget.Category] = budget
	}
	return nil
}

func (s *Storage) AddTransaction(tx Transaction) (Transaction, error) {
	if err := tx.Validate(); err != nil {
		return Transaction{}, fmt.Errorf("invalid transaction: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if budget, ok := s.budgets[tx.Category]; ok {
		total := s.computeTotalByCategory(tx.Category) + tx.Amount
		if total > budget.Limit {
			return Transaction{}, ErrBudgetExceeded
		}
	}

	if tx.ID == 0 {
		tx.ID = s.lastID + 1
	}
	s.lastID = max(s.lastID, tx.ID)
	s.transactions = append(s.transactions, tx)
	return tx, nil
}

func (s *Storage) ListTransactions() []Transaction {
	s.mu.Lock()
	defer s.mu.Unlock()

	return slices.Clone(s.transactions)
}

func (s *Storage) computeTotalByCategory(category string) float64 {
	var total float64
	for _, tx := range s.transactions {
		if tx.Category == category {
			total += tx.Amount
		}
	}
	return total
}

func (s *Storage) ListBudgets() map[string]Budget {
	s.mu.Lock()
	defer s.mu.Unlock()

	return maps.Clone(s.budgets)
}
