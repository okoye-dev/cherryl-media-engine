package main

import (
	"bytes"
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	appconfig "github.com/okoye-dev/cherryl-media-engine/internal/config"
)

func main() {
	cfg := appconfig.Load()
	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(cfg.S3.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.S3.AccessKeyID, cfg.S3.SecretAccessKey, "")),
	)
	if err != nil {
		log.Fatal(err)
	}
	endpoint := cfg.S3.Endpoint
	if !bytes.HasPrefix([]byte(endpoint), []byte("http://")) && !bytes.HasPrefix([]byte(endpoint), []byte("https://")) {
		endpoint = "https://" + endpoint
	}
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = cfg.S3.ForcePathStyle
		o.BaseEndpoint = aws.String(endpoint)
	})
	out, err := client.ListObjectsV2(context.Background(), &s3.ListObjectsV2Input{
		Bucket: aws.String(cfg.S3.BucketName),
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Total objects in bucket: %d\n", len(out.Contents))
	for _, obj := range out.Contents {
		fmt.Printf(" - %s (%d bytes)\n", *obj.Key, *obj.Size)
	}
}
