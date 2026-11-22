package main

import (
	"fmt"
	"log"
	"os"
	"time"

	ledger "ledger/core"
	"ledger/internal/db"
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

func CheckValid(v Validatable) error {
	return v.Validate()
}

func main() {
	conn, err := db.Open()
	if err != nil {
		log.Fatalf("Ledger startup failed: %v", err)
	}
	defer conn.Close()
	log.Print("Ledger connected to PostgreSQL")

	fmt.Println("Ledger service started")

	s := ledger.NewStorage()

	_, err1 := s.AddTransaction(ledger.Transaction{
		ID:          1,
		Amount:      100.0,
		Category:    "subscription",
		Description: "test1",
		Date:        time.Now(),
	})
	if err1 != nil {
		fmt.Println(err1)
	}

	_, err2 := s.AddTransaction(ledger.Transaction{
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

	s.SetBudget(ledger.Budget{Category: "Хозтовары", Limit: 1000.0})
	s.SetBudget(ledger.Budget{Category: "Развлечения", Limit: 5000.0})

	// превышение лимита
	_, err3 := s.AddTransaction(ledger.Transaction{
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
	_, err4 := s.AddTransaction(ledger.Transaction{
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

	defer f.Close()

	if err := s.LoadBudgets(f); err != nil {
		fmt.Println(err)
	}

	fmt.Println(s.ListBudgets())

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
