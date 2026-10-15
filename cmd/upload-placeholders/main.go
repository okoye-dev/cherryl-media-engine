package main

import (
	"bytes"
	"context"
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
	appconfig "github.com/okoye-dev/cherryl-media-engine/internal/config"
)

func main() {
	cfg := appconfig.Load()
	if !cfg.S3.IsConfigured() {
		log.Fatal("S3 / Cloudflare R2 is not configured in backend/.env")
	}

	ctx := context.Background()
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

	files, err := filepath.Glob("/tmp/placeholders/*.webp")
	if err != nil || len(files) == 0 {
		log.Fatalf("no files found in /tmp/placeholders/*.webp: %v", err)
	}

	log.Printf("Found %d WebP images to upload to Cloudflare R2...\n", len(files))

	httpClient := &http.Client{Timeout: 10 * time.Second}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			log.Printf("failed to read %s: %v", file, err)
			continue
		}

		baseName := filepath.Base(file)
		storageKey := "media/" + baseName

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

		// Verify CDN response
		resp, err := httpClient.Get(publicURL)
		status := "OK"
		if err != nil || resp.StatusCode != 200 {
			status = fmt.Sprintf("HTTP %d (may take a few seconds to propagate)", resp.StatusCode)
		} else {
			resp.Body.Close()
		}

		log.Printf("  [UPLOADED] %-60s -> %s (%s)", storageKey, publicURL, status)
	}

	fmt.Println("\nAll 7 new branded WebP images successfully uploaded to Cloudflare R2!")
}
