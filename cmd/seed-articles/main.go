package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/okoye-dev/cherryl-media-engine/internal/config"
	"github.com/okoye-dev/cherryl-media-engine/internal/migrate"
	"github.com/okoye-dev/cherryl-media-engine/internal/models"
	"github.com/okoye-dev/cherryl-media-engine/internal/seed"
	"github.com/okoye-dev/cherryl-media-engine/internal/store"
	"github.com/okoye-dev/cherryl-media-engine/internal/urlutil"
)

func main() {
	dbFlag := flag.String("db", "", "PostgreSQL / Supabase connection string (overrides DATABASE_URL)")
	apiFlag := flag.String("api", "", "Backend API URL to seed articles via HTTP (e.g. http://localhost:8080)")
	flag.Parse()

	cfg := config.Load()

	dbURL := *dbFlag
	if dbURL == "" {
		dbURL = cfg.Database.URL
	}

	apiURL := *apiFlag

	if dbURL == "" && apiURL == "" {
		fmt.Println("Usage of cherryl-media article seeder:")
		fmt.Println("  Option A (Direct Database):")
		fmt.Println("    go run cmd/seed-articles/main.go --db=\"postgres://postgres.[ref]:[pass]@aws-0-[region].pooler.supabase.com:6543/postgres\"")
		fmt.Println("    (or set DATABASE_URL in backend/.env)")
		fmt.Println("")
		fmt.Println("  Option B (Backend HTTP API):")
		fmt.Println("    go run cmd/seed-articles/main.go --api=\"http://localhost:8080\"")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// 1. Gather all canonical website copy in admin post format
	log.Println("Compiling all website articles...")
	articles := seed.GetAllWebsiteArticles()

	log.Printf("Prepared %d articles for database ingestion.\n", len(articles))

	// Normalizer for media keys
	normalizeMedia := func(art *models.Article) {
		if art.FeaturedImage.StorageKey == "" && art.FeaturedImage.URL != "" {
			key := urlutil.ExtractStorageKey(art.FeaturedImage.URL)
			if strings.HasPrefix(key, "media/") {
				art.FeaturedImage.StorageKey = key
			}
		}
		for i := range art.Body {
			block := &art.Body[i]
			if (block.Type == "image" || block.Type == "video") && block.StorageKey == "" && block.Src != "" {
				key := urlutil.ExtractStorageKey(block.Src)
				if strings.HasPrefix(key, "media/") {
					block.StorageKey = key
				}
			}
		}
	}

	// Option A: Direct PostgreSQL / Supabase write
	if dbURL != "" && apiURL == "" {
		log.Println("Connecting to Supabase PostgreSQL database...")
		pool, err := pgxpool.New(ctx, dbURL)
		if err != nil {
			log.Fatalf("Failed to initialize database pool: %v", err)
		}
		defer pool.Close()

		if err := pool.Ping(ctx); err != nil {
			log.Fatalf("Database ping failed: %v", err)
		}
		log.Println("Database connection verified.")

		// Run auto-migrations first to ensure schema parity
		log.Println("Running schema migrations...")
		if err := migrate.Run(ctx, pool); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}

		pgStore := store.NewPGStore(pool)
		successCount := 0

		for _, art := range articles {
			a := art
			normalizeMedia(&a)
			if err := pgStore.Save(ctx, &a); err != nil {
				log.Printf("  [FAIL] %-40s: %v", a.Slug, err)
			} else {
				successCount++
				socialBadge := "no social"
				if a.SocialVariants.X != "" || a.SocialVariants.Facebook != "" || a.SocialVariants.LinkedIn != "" || a.SocialVariants.Newsletter != "" {
					socialBadge = "social variants attached"
				}
				log.Printf("  [SAVED] %-55s | %-24s | %s", a.Slug, a.Category, socialBadge)
			}
		}

		fmt.Printf("\nDone! Successfully wrote %d of %d articles to database.\n", successCount, len(articles))
		return
	}

	// Option B: HTTP API write
	if apiURL != "" {
		apiURL = strings.TrimSuffix(apiURL, "/")
		log.Printf("Connecting to backend API at %s...\n", apiURL)
		client := &http.Client{Timeout: 10 * time.Second}
		successCount := 0

		for _, art := range articles {
			a := art
			normalizeMedia(&a)
			bodyBytes, err := json.Marshal(a)
			if err != nil {
				log.Printf("  [FAIL] JSON encode %s: %v", a.Slug, err)
				continue
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL+"/api/articles", bytes.NewReader(bodyBytes))
			if err != nil {
				log.Printf("  [FAIL] Request create %s: %v", a.Slug, err)
				continue
			}
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			if err != nil {
				log.Printf("  [FAIL] HTTP post %s: %v", a.Slug, err)
				continue
			}
			resp.Body.Close()

			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
				log.Printf("  [FAIL] HTTP status %d for %s", resp.StatusCode, a.Slug)
			} else {
				successCount++
				log.Printf("  [PUBLISHED] %-55s | %-24s", a.Slug, a.Category)
			}
		}

		fmt.Printf("\nDone! Successfully published %d of %d articles via backend API.\n", successCount, len(articles))
	}
}
