package core

import (
	"errors"
	"time"
)

type Transaction struct {
	ID          int
	Amount      float64
	Category    string
	Description string
	Date        time.Time
}

type Budget struct {
	Category string  `json:"Category"`
	Limit    float64 `json:"Limit"`
}

func (t *Transaction) Validate() error {
	var errs []error
	if t.Amount <= 0 {
		errs = append(errs, errors.New("amount must be a positive number"))
	}
	if t.Category == "" {
		errs = append(errs, errors.New("category must't be blank"))
	}
	if t.Description == "" {
		errs = append(errs, errors.New("description must not be blank"))
	}
	return errors.Join(errs...)
}

func (b *Budget) Validate() error {
	var errs []error
	if b.Limit <= 0 {
		errs = append(errs, errors.New("limit must be a positive number"))
	}
	if b.Category == "" {
		errs = append(errs, errors.New("category must't be blank"))
	}
	return errors.Join(errs...)
}
