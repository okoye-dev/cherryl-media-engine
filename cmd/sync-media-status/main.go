package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/okoye-dev/cherryl-media-engine/internal/config"
	"github.com/okoye-dev/cherryl-media-engine/internal/storage"
	"github.com/okoye-dev/cherryl-media-engine/internal/store"
	"github.com/okoye-dev/cherryl-media-engine/internal/urlutil"
)

func main() {
	cfg := config.Load()
	if cfg.Database.URL == "" {
		log.Fatal("DATABASE_URL must be configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	s3Storage, err := storage.NewS3Storage(&cfg.S3)
	if err != nil {
		log.Fatalf("failed to initialize S3: %v", err)
	}

	pool, err := store.NewPool(ctx, cfg.Database.URL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	pg := store.NewPGStore(pool)

	articles, err := pg.List(ctx, "all")
	if err != nil {
		log.Fatalf("failed to list articles: %v", err)
	}

	fmt.Printf("Analyzing media assets for %d database articles...\n", len(articles))

	confirmedArticlesCount := 0

	for _, art := range articles {
		var keys []string

		// Cover image key
		if art.FeaturedImage.StorageKey != "" {
			keys = append(keys, art.FeaturedImage.StorageKey)
		} else if key := urlutil.ExtractStorageKey(art.FeaturedImage.URL); strings.HasPrefix(key, "media/") {
			keys = append(keys, key)
			art.FeaturedImage.StorageKey = key
		}

		// Body blocks keys
		for i := range art.Body {
			b := &art.Body[i]
			if b.Type == "image" || b.Type == "video" {
				if b.StorageKey != "" {
					keys = append(keys, b.StorageKey)
				} else if key := urlutil.ExtractStorageKey(b.Src); strings.HasPrefix(key, "media/") {
					keys = append(keys, key)
					b.StorageKey = key
				}
			}
		}

		allConfirmed := true
		for _, key := range keys {
			// Ensure it exists in uploads table
			isConf, _ := pg.IsConfirmed(ctx, key)
			if !isConf {
				// Check S3 size or register fallback size
				publicURL := s3Storage.GetPublicURL(key)
				size, err := s3Storage.GetFileSize(ctx, key)
				if err != nil || size <= 0 {
					size = 250000 // default fallback size if offline
				}
				if confErr := pg.ConfirmUpload(ctx, key, size, publicURL); confErr != nil {
					log.Printf("confirm upload %s: %v", key, confErr)
					allConfirmed = false
				} else {
					fmt.Printf("  ✓ Confirmed upload asset: %s (size: %d bytes)\n", key, size)
				}
			}
		}

		if art.Status == "published" && allConfirmed {
			art.MediaConfirmed = true
			if err := pg.Save(ctx, &art); err != nil {
				log.Printf("save article %s: %v", art.Slug, err)
			} else {
				confirmedArticlesCount++
			}
		}
	}

	fmt.Printf("\nSynchronization complete:\n")
	fmt.Printf(" - Total articles checked: %d\n", len(articles))
	fmt.Printf(" - Published articles with media_confirmed=true: %d\n", confirmedArticlesCount)
}
