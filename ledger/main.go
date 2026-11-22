package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	ledger "ledger/core"
	"ledger/internal/cache"
	"ledger/internal/report"
)

func main() {
	from := flag.String("from", "", "report start date (YYYY-MM-DD)")
	to := flag.String("to", "", "report end date (YYYY-MM-DD)")
	flag.Parse()

	if err := run(*from, *to); err != nil {
		log.Fatal(err)
	}
}

func run(from, to string) error {
	reportRequested := from != "" || to != ""
	var start, end time.Time
	if reportRequested {
		var err error
		start, err = time.Parse(time.DateOnly, from)
		if err != nil {
			return fmt.Errorf("invalid -from date, expected YYYY-MM-DD: %w", err)
		}
		end, err = time.Parse(time.DateOnly, to)
		if err != nil {
			return fmt.Errorf("invalid -to date, expected YYYY-MM-DD: %w", err)
		}
	}

	store, err := ledger.OpenStorage()
	if err != nil {
		return fmt.Errorf("Ledger startup failed: %w", err)
	}
	defer store.Close()
	log.Print("Ledger connected to PostgreSQL")

	ctx := context.Background()
	redis, err := cache.Open(ctx)
	if err != nil {
		log.Printf("Ledger Redis cache unavailable: %v", err)
	} else {
		defer redis.Close()
		log.Print("Ledger connected to Redis")
	}

	if !reportRequested {
		return nil
	}
	summary, err := report.New(store, redis).GetReportSummary(ctx, start, end)
	if err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(summary)
}
