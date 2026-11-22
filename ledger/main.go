package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	ledger "ledger/core"
)

// 1. Создайте интерфейс Validatable с единственным методом Validate() error. Назначение интерфейса — задать контракт на проверку валидности данных структуры.
// 2. Реализуйте интерфейс Validatable для структуры Transaction. Метод Validate() для Transaction должен проверять корректность полей транзакции, например:
// · Amount не должен быть 0 или отрицательным,
// · Category не должна быть пустой строкой,
// · Date должна быть в корректном формате (при использовании строки можно пропустить проверку формата).
// Если все поля заполнены корректно, должно возвращаться nil, иначе должно возвращаться соответствующая ошибка с описанием проблемы.
// Методы Validate() не должны логировать, они только возвращают error. Логирование происходит в вызывающем коде.
// 3. Реализуйте интерфейс Validatable для структуры Budget. Метод Validate() для Budget проверяет:
// · Limit (лимит) — должен быть положительным числом;
// · Category — не должна быть пустая.
// При неверных данных необходимо вернуть ошибку, при корректных — nil.
// 4. Модифицируйте функции добавления данных в Ledger: перед добавлением новой транзакции или бюджета, вызывайте метод Validate() для соответствующей структуры. Если метод вернул ошибку, не добавляйте объект и передайте ошибку дальше или логируйте её. Таким образом, вы отделите логику проверки данных от основной бизнес-логики.
// 5. Для демонстрации работы интерфейсов напишите функцию, принимающую параметр типа Validatable (например, func CheckValid(v Validatable) error), которая вызывает v.Validate() и возвращает результат. В main протестируйте эту функцию с объектом Transaction и Budget, убедившись, что она корректно работает с разными типами (полиморфизм через интерфейс).

type Validatable interface {
	Validate() error
}

type Storage []ledger.Transaction

var budgetStorage = make(map[string]ledger.Budget, 4)

func SetBudget(b ledger.Budget) {
	budgetStorage[b.Category] = b
}

func LoadBudgets(r io.Reader) error {
	var err error
	budgets := make([]ledger.Budget, 0, 2)

	if err = json.NewDecoder(r).Decode(&budgets); err != nil {
		return fmt.Errorf("error decoding json file: %w", err)
	}

	for _, budget := range budgets {
		SetBudget(budget)
	}
	return err
}

func (s *Storage) AddTransaction(tx ledger.Transaction) error {
	if err := tx.Validate(); err != nil {
		return fmt.Errorf("invalid transaction: %w", err)
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

func (s *Storage) ListTransactions() []ledger.Transaction {
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

func CheckValid(v Validatable) error {
	return v.Validate()
}

func main() {
	fmt.Println("Ledger service started")

	s := make(Storage, 0, 10)

	err1 := s.AddTransaction(ledger.Transaction{
		ID:          1,
		Amount:      100.0,
		Category:    "subscription",
		Description: "test1",
		Date:        time.Now(),
	})
	if err1 != nil {
		fmt.Println(err1)
	}

	err2 := s.AddTransaction(ledger.Transaction{
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

	SetBudget(ledger.Budget{Category: "Хозтовары", Limit: 1000.0})
	SetBudget(ledger.Budget{Category: "Развлечения", Limit: 5000.0})

	// превышение лимита
	err3 := s.AddTransaction(ledger.Transaction{
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
	err4 := s.AddTransaction(ledger.Transaction{
		ID:          4,
		Amount:      1000.0,
		Category:    "Развлечения",
		Description: "test4",
		Date:        time.Now(),
	})
	if err4 != nil {
		fmt.Println(err4)
	}

	f, fileErr := os.Open("budgets.json")
	if fileErr != nil {
		panic(fileErr)
	}

	err := LoadBudgets(f)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(budgetStorage)

	var b Validatable = &ledger.Budget{}
	validBErr := CheckValid(b)
	if validBErr != nil {
		fmt.Println(validBErr)
	}

	var t Validatable = &ledger.Transaction{}
	validTErr := CheckValid(t)
	if validTErr != nil {
		fmt.Println(validTErr)
	}
}
