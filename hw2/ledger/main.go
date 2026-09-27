package main

import (
	"fmt"
	"time"
)

type Transaction struct {
	ID          int
	Amount      float64
	Category    string
	Description string
	Date        time.Time
}

type Storage []Transaction

func (s *Storage) AddTransaction(tx Transaction) error {
	if tx.Amount < 0 || tx.Category == "" || tx.Description == "" {
		return fmt.Errorf("invalid transaction")
	}

	*s = append(*s, tx)
	return nil
}

func (s *Storage) ListTransactions() []Transaction {
	return *s
}

func main() {
	fmt.Println("Ledger service started")

	s := make(Storage, 0, 2)
	s.AddTransaction(Transaction{
		ID:          1,
		Amount:      100.0,
		Category:    "subscription",
		Description: "test1",
		Date:        time.Now(),
	})
	s.AddTransaction(Transaction{
		ID:          2,
		Amount:      300.0,
		Category:    "subscription",
		Description: "test2",
		Date:        time.Now(),
	})

	for _, tx := range s.ListTransactions() {
		fmt.Println(tx)
	}
}
