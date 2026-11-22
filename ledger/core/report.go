package core

import (
	"context"
	"fmt"
	"time"
)

type CategoryTotal struct {
	Category string  `json:"category"`
	Total    float64 `json:"total"`
}

func (s *Storage) SummarizeExpenses(ctx context.Context, from, to time.Time) ([]CategoryTotal, error) {
	const query = `
		SELECT category, SUM(amount)
		FROM expenses
		WHERE date BETWEEN $1::text::date AND $2::text::date
		GROUP BY category
		ORDER BY category`
	rows, err := s.conn.QueryContext(ctx, query, from.Format(time.DateOnly), to.Format(time.DateOnly))
	if err != nil {
		return nil, fmt.Errorf("query report summary: %w", err)
	}
	defer rows.Close()

	result := make([]CategoryTotal, 0)
	for rows.Next() {
		var total CategoryTotal
		if err := rows.Scan(&total.Category, &total.Total); err != nil {
			return nil, fmt.Errorf("scan report summary: %w", err)
		}
		result = append(result, total)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read report summary: %w", err)
	}
	return result, nil
}
