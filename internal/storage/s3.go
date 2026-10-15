package storage

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	appconfig "github.com/okoye-dev/cherryl-media-engine/internal/config"
)

const (
	UploadPresignTTL = 15 * time.Minute
	DownloadPresignTTL = 1 * time.Hour
)

var ErrObjectNotFound = errors.New("object not found")

type Storage interface {
	GetPresignedUploadURL(ctx context.Context, key string, size int64, contentType string) (string, error)
	GetPublicURL(key string) string
	GetPresignedURL(ctx context.Context, key string, forceDownload bool) (string, error)
	DeleteFile(ctx context.Context, key string) error
	GetFileSize(ctx context.Context, key string) (int64, error)
	IsConfigured() bool
}

type S3Storage struct {
	client     *s3.Client
	bucketName string
	publicURL  string
	endpoint   string
	useSSL     bool
	configured bool
}

func NewS3Storage(cfg *appconfig.S3Config) (*S3Storage, error) {
	if !cfg.IsConfigured() {
		return &S3Storage{
			configured: false,
			bucketName: cfg.BucketName,
			publicURL:  cfg.PublicURL,
		}, nil
	}

	httpClient := &http.Client{
		Timeout: 10 * time.Minute,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 20,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		),
		config.WithHTTPClient(httpClient),
	)
	if err != nil {
		return nil, fmt.Errorf("loading AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = cfg.ForcePathStyle
		if cfg.Endpoint != "" {
			endpoint := cfg.Endpoint
			if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
				scheme := "https://"
				if !cfg.UseSSL {
					scheme = "http://"
				}
				endpoint = scheme + endpoint
			}
			o.BaseEndpoint = aws.String(endpoint)
		}
	})

	return &S3Storage{
		client:     client,
		bucketName: cfg.BucketName,
		publicURL:  strings.TrimRight(cfg.PublicURL, "/"),
		endpoint:   cfg.Endpoint,
		useSSL:     cfg.UseSSL,
		configured: true,
	}, nil
}

func (s *S3Storage) IsConfigured() bool {
	return s.configured
}

func (s *S3Storage) GetPresignedUploadURL(ctx context.Context, key string, size int64, contentType string) (string, error) {
	if !s.configured {
		return "", errors.New("s3 storage is not configured")
	}

	input := &s3.PutObjectInput{
		Bucket:        aws.String(s.bucketName),
		Key:           aws.String(key),
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(contentType),
	}

	request, err := s3.NewPresignClient(s.client).PresignPutObject(ctx, input, func(opts *s3.PresignOptions) {
		opts.Expires = UploadPresignTTL
	})
	if err != nil {
		return "", fmt.Errorf("presigning upload %q: %w", key, err)
	}

	return request.URL, nil
}

func (s *S3Storage) GetPublicURL(key string) string {
	if s.publicURL != "" {
		return fmt.Sprintf("%s/%s", s.publicURL, strings.TrimLeft(key, "/"))
	}
	if s.endpoint != "" {
		scheme := "https://"
		if !s.useSSL {
			scheme = "http://"
		}
		endpoint := strings.TrimPrefix(strings.TrimPrefix(s.endpoint, "https://"), "http://")
		return fmt.Sprintf("%s%s/%s/%s", scheme, endpoint, s.bucketName, strings.TrimLeft(key, "/"))
	}
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.bucketName, strings.TrimLeft(key, "/"))
}

func (s *S3Storage) GetPresignedURL(ctx context.Context, key string, forceDownload bool) (string, error) {
	if !s.configured {
		return "", errors.New("s3 storage is not configured")
	}

	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}
	if forceDownload {
		input.ResponseContentDisposition = aws.String("attachment")
	}

	request, err := s3.NewPresignClient(s.client).PresignGetObject(ctx, input, func(opts *s3.PresignOptions) {
		opts.Expires = DownloadPresignTTL
	})
	if err != nil {
		return "", fmt.Errorf("presigning get %q: %w", key, err)
	}

	return request.URL, nil
}

func (s *S3Storage) DeleteFile(ctx context.Context, key string) error {
	if !s.configured {
		return errors.New("s3 storage is not configured")
	}

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("deleting %q: %w", key, err)
	}
	return nil
}

func (s *S3Storage) GetFileSize(ctx context.Context, key string) (int64, error) {
	if !s.configured {
		return 0, errors.New("s3 storage is not configured")
	}

	result, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			switch apiErr.ErrorCode() {
			case "NotFound", "NoSuchKey":
				return 0, ErrObjectNotFound
			}
		}
		return 0, fmt.Errorf("head %q: %w", key, err)
	}

	return aws.ToInt64(result.ContentLength), nil
}
