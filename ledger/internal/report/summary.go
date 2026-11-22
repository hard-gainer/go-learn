package report

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"ledger/core"
	"ledger/internal/cache"
)

const summaryTTL = 30 * time.Second

var errInvalidRange = errors.New("report start date must not be after end date")

type Service struct {
	store *core.Storage
	redis *cache.Redis
}

func New(store *core.Storage, redis *cache.Redis) *Service {
	return &Service{store: store, redis: redis}
}

func (s *Service) GetReportSummary(ctx context.Context, from, to time.Time) ([]core.CategoryTotal, error) {
	summary, _, err := s.getSummary(ctx, from, to)
	return summary, err
}

func (s *Service) getSummary(ctx context.Context, from, to time.Time) ([]core.CategoryTotal, string, error) {
	start, end := from.Format(time.DateOnly), to.Format(time.DateOnly)
	if start > end {
		return nil, "", errInvalidRange
	}

	if s.redis == nil {
		summary, err := s.store.SummarizeExpenses(ctx, from, to)
		return summary, "BYPASS", err
	}
	key := "report:summary:" + start + ":" + end
	if data, err := s.redis.Get(ctx, key); err != nil {
		log.Printf("read report cache: %v", err)
	} else if data != nil {
		var cached []core.CategoryTotal
		if err := json.Unmarshal(data, &cached); err != nil {
			log.Printf("decode report cache: %v", err)
		} else if cached != nil {
			return cached, "HIT", nil
		}
	}

	summary, err := s.store.SummarizeExpenses(ctx, from, to)
	if err != nil {
		return nil, "", err
	}
	data, err := json.Marshal(summary)
	if err != nil {
		log.Printf("encode report cache: %v", err)
	} else if err := s.redis.Set(ctx, key, data, summaryTTL); err != nil {
		log.Printf("write report cache: %v", err)
	} else {
		return summary, "MISS", nil
	}
	return summary, "BYPASS", nil
}
