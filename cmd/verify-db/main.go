package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/okoye-dev/cherryl-media-engine/internal/config"
	"github.com/okoye-dev/cherryl-media-engine/internal/store"
)

func main() {
	cfg := config.Load()

	dbURL := cfg.Database.URL
	if dbURL == "" {
		dbURL = os.Getenv("DATABASE_URL")
	}

	pool, err := store.NewPool(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("pool: %v", err)
	}
	defer pool.Close()

	var artCount, secCount, setCount, migCount int
	_ = pool.QueryRow(context.Background(), "SELECT count(*) FROM articles").Scan(&artCount)
	_ = pool.QueryRow(context.Background(), "SELECT count(*) FROM curated_sections").Scan(&secCount)
	_ = pool.QueryRow(context.Background(), "SELECT count(*) FROM site_settings").Scan(&setCount)
	_ = pool.QueryRow(context.Background(), "SELECT count(*) FROM schema_migrations").Scan(&migCount)

	fmt.Printf("Supabase Verification Results:\n")
	fmt.Printf(" - schema_migrations: %d\n", migCount)
	fmt.Printf(" - articles: %d\n", artCount)
	fmt.Printf(" - curated_sections: %d\n", secCount)
	fmt.Printf(" - site_settings: %d\n", setCount)
}
