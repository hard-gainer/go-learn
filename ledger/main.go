package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
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

type Budget struct {
	Category string  `json:"Category"`
	Limit    float64 `json:"Limit"`
}

var budgetStorage = make(map[string]Budget, 4)

func SetBudget(b Budget) {
	budgetStorage[b.Category] = b
}

func LoadBudgets(r io.Reader) error {
	var err error
	budgets := make([]Budget, 0, 2)

	if err = json.NewDecoder(r).Decode(&budgets); err != nil {
		return fmt.Errorf("error decoding json file: %w", err)
	}

	for _, budget := range budgets {
		SetBudget(budget)
	}
	return err
}

func (s *Storage) AddTransaction(tx Transaction) error {
	if tx.Amount < 0 || tx.Category == "" || tx.Description == "" {
		return fmt.Errorf("invalid transaction")
	}

	if budget, ok := budgetStorage[tx.Category]; ok {
		total := s.computeTotalByCategory(tx.Category) + tx.Amount
		if budget.Limit < total {
			return errors.New("budget exceeded")
		}
	}

	*s = append(*s, tx)
	return nil
}

func (s *Storage) ListTransactions() []Transaction {
	return *s
}

func (s *Storage) computeTotalByCategory(category string) float64 {
	var sum float64
	for _, tx := range *s {
		if tx.Category == category {
			sum += tx.Amount
		}
	}
	return sum
}

func main() {
	fmt.Println("Ledger service started")

	s := make(Storage, 0, 10)

	err1 := s.AddTransaction(Transaction{
		ID:          1,
		Amount:      100.0,
		Category:    "subscription",
		Description: "test1",
		Date:        time.Now(),
	})
	if err1 != nil {
		fmt.Println(err1)
	}

	err2 := s.AddTransaction(Transaction{
		ID:          2,
		Amount:      300.0,
		Category:    "subscription",
		Description: "test2",
		Date:        time.Now(),
	})
	if err2 != nil {
		fmt.Println(err2)
	}

	for _, tx := range s.ListTransactions() {
		fmt.Println(tx)
	}

	SetBudget(Budget{Category: "Хозтовары", Limit: 1000.0})
	SetBudget(Budget{Category: "Развлечения", Limit: 5000.0})

	// превышение лимита
	err3 := s.AddTransaction(Transaction{
		ID:          3,
		Amount:      50000.0,
		Category:    "Развлечения",
		Description: "test3",
		Date:        time.Now(),
	})
	if err3 != nil {
		fmt.Println(err3)
	}

	// должна пройти
	err4 := s.AddTransaction(Transaction{
		ID:          4,
		Amount:      1000.0,
		Category:    "Развлечения",
		Description: "test4",
		Date:        time.Now(),
	})
	if err4 != nil {
		fmt.Println(err4)
	}

	f, fileErr := os.Open("ledger/budgets.json")
	if fileErr != nil {
		panic(fileErr)
	}

	err := LoadBudgets(f)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(budgetStorage)
}
