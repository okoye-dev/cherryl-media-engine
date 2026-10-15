package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jackc/pgx/v5/pgxpool"
	appconfig "github.com/okoye-dev/cherryl-media-engine/internal/config"
	"github.com/okoye-dev/cherryl-media-engine/internal/store"
)

func main() {
	cfg := appconfig.Load()
	if !cfg.S3.IsConfigured() {
		log.Fatal("S3 / Cloudflare R2 is not configured in backend/.env")
	}

	manifestData, err := os.ReadFile("/tmp/clean_photos/manifest.json")
	if err != nil {
		log.Fatalf("failed to read manifest: %v", err)
	}

	var manifest map[string]string
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		log.Fatalf("failed to parse manifest: %v", err)
	}

	ctx := context.Background()

	// Connect to S3 / Cloudflare R2
	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(cfg.S3.Region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.S3.AccessKeyID, cfg.S3.SecretAccessKey, ""),
		),
	)
	if err != nil {
		log.Fatalf("failed to load AWS config: %v", err)
	}

	endpoint := cfg.S3.Endpoint
	if !bytes.HasPrefix([]byte(endpoint), []byte("http://")) && !bytes.HasPrefix([]byte(endpoint), []byte("https://")) {
		endpoint = "https://" + endpoint
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = cfg.S3.ForcePathStyle
		o.BaseEndpoint = aws.String(endpoint)
	})

	// Connect to database to record uploads
	dbURL := cfg.Database.URL
	if dbURL == "" {
		dbURL = os.Getenv("DATABASE_URL")
	}
	var dbPool *pgxpool.Pool
	if dbURL != "" {
		dbPool, err = store.NewPool(ctx, dbURL)
		if err != nil {
			log.Printf("Warning: failed to connect to DB for uploads recording: %v", err)
		} else {
			defer dbPool.Close()
		}
	}

	log.Printf("Uploading %d clean images to Cloudflare R2 under exact keys...\n", len(manifest))

	httpClient := &http.Client{Timeout: 10 * time.Second}
	uploadedCount := 0

	for storageKey, filePath := range manifest {
		data, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("failed to read %s: %v", filePath, err)
			continue
		}

		_, err = client.PutObject(ctx, &s3.PutObjectInput{
			Bucket:        aws.String(cfg.S3.BucketName),
			Key:           aws.String(storageKey),
			Body:          bytes.NewReader(data),
			ContentLength: aws.Int64(int64(len(data))),
			ContentType:   aws.String("image/webp"),
		})
		if err != nil {
			log.Fatalf("failed to upload %s: %v", storageKey, err)
		}

		publicURL := fmt.Sprintf("%s/%s", cfg.S3.PublicURL, storageKey)

		// Record in database uploads table if connected
		if dbPool != nil {
			fileName := filepath.Base(storageKey)
			_, _ = dbPool.Exec(ctx, `
				INSERT INTO uploads (storage_key, file_name, size_bytes, content_type, public_url, created_at)
				VALUES ($1, $2, $3, 'image/webp', $4, NOW())
				ON CONFLICT (storage_key) DO UPDATE
				SET size_bytes = EXCLUDED.size_bytes, public_url = EXCLUDED.public_url
			`, storageKey, fileName, int64(len(data)), publicURL)
		}

		// Verify CDN response
		resp, err := httpClient.Get(publicURL)
		status := "OK"
		if err != nil || resp.StatusCode != 200 {
			if resp != nil {
				status = fmt.Sprintf("HTTP %d", resp.StatusCode)
			} else {
				status = "timeout"
			}
		} else {
			resp.Body.Close()
		}

		log.Printf("  [UPLOADED] %-60s (%6d B) -> %s (%s)", storageKey, len(data), publicURL, status)
		uploadedCount++
	}

	fmt.Printf("\nDone! Successfully uploaded %d images to Cloudflare R2 under original keys.\n", uploadedCount)
}
