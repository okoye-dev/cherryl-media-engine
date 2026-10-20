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

	var artCount, secCount, setCount, migCount, upCount, subCount, dispCount, unconfCount int
	_ = pool.QueryRow(context.Background(), "SELECT count(*) FROM articles").Scan(&artCount)
	_ = pool.QueryRow(context.Background(), "SELECT count(*) FROM articles WHERE media_confirmed = true").Scan(&unconfCount)
	_ = pool.QueryRow(context.Background(), "SELECT count(*) FROM curated_sections").Scan(&secCount)
	_ = pool.QueryRow(context.Background(), "SELECT count(*) FROM site_settings").Scan(&setCount)
	_ = pool.QueryRow(context.Background(), "SELECT count(*) FROM schema_migrations").Scan(&migCount)
	_ = pool.QueryRow(context.Background(), "SELECT count(*) FROM uploads WHERE status = 'confirmed'").Scan(&upCount)
	_ = pool.QueryRow(context.Background(), "SELECT count(*) FROM post_subscribers").Scan(&subCount)
	_ = pool.QueryRow(context.Background(), "SELECT count(*) FROM post_email_dispatches").Scan(&dispCount)

	fmt.Printf("Supabase Verification Results:\n")
	fmt.Printf(" - schema_migrations: %d\n", migCount)
	fmt.Printf(" - articles: %d (media_confirmed=true: %d)\n", artCount, unconfCount)
	fmt.Printf(" - curated_sections: %d\n", secCount)
	fmt.Printf(" - site_settings: %d\n", setCount)
	fmt.Printf(" - uploads (confirmed): %d\n", upCount)
	fmt.Printf(" - post_subscribers: %d\n", subCount)
	fmt.Printf(" - post_email_dispatches: %d\n", dispCount)
}
