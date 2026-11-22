package main

import (
	"log"

	ledger "ledger/core"
)

func main() {
	store, err := ledger.OpenStorage()
	if err != nil {
		log.Fatalf("Ledger startup failed: %v", err)
	}
	defer store.Close()
	log.Print("Ledger connected to PostgreSQL")
}
