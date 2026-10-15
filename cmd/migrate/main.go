package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/okoye-dev/cherryl-media-engine/internal/config"
	"github.com/okoye-dev/cherryl-media-engine/internal/migrate"
)

func main() {
	dbFlag := flag.String("db", "", "PostgreSQL / Supabase connection string (overrides DATABASE_URL)")
	flag.Parse()

	cfg := config.Load()

	dbURL := *dbFlag
	if dbURL == "" {
		dbURL = cfg.Database.URL
	}

	if dbURL == "" {
		log.Fatal("Migration failed: DATABASE_URL is not set in environment or .env file, and no --db flag was provided.\n" +
			"Example: go run cmd/migrate/main.go --db=\"postgres://postgres.[ref]:[password]@aws-0-[region].pooler.supabase.com:6543/postgres\"")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	log.Println("Connecting to Supabase PostgreSQL database...")
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to initialize database pool: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Database ping failed: %v", err)
	}
	log.Println("Database connection verified. Running auto-migrations...")

	startTime := time.Now()
	if err := migrate.Run(ctx, pool); err != nil {
		log.Fatalf("Migration execution failed: %v", err)
	}

	fmt.Printf("Auto-migrations completed successfully in %s.\n", time.Since(startTime).Round(time.Millisecond))
	os.Exit(0)
}
