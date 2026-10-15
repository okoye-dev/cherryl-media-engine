# cherryl-media-engine

Lightweight, high-performance Go backend service for Cherryl Media publishing and discovery operations.

## Features
- **S3 Presigned Uploads**: Generates secure presigned PUT URLs for direct client-to-bucket uploads with 15-minute expiry.
- **Supabase PostgreSQL**: Persistent storage for articles and uploads with embedded SQL migrations (`pgx/v5`).
- **In-Memory Fallback**: Runs smoothly without setup if `DATABASE_URL` is not supplied.
- **Link Resolution**: Differentiates internal article links from external web resources.
- **Fast HTTP APIs**: REST endpoints with built-in CORS for Next.js frontend consumption.

## Configuration
Set environment variables:
- `PORT`: HTTP port (default `8080`)
- `DATABASE_URL`: Supabase / PostgreSQL connection string
- `S3_ENDPOINT`: S3 / Cloudflare R2 endpoint URL
- `S3_REGION`: S3 region (default `auto` or `us-east-1`)
- `S3_ACCESS_KEY_ID`: Access key ID
- `S3_SECRET_ACCESS_KEY`: Secret access key
- `S3_BUCKET_NAME`: S3 / R2 bucket name
- `S3_PUBLIC_URL`: Public CDN / delivery domain

## Running
```bash
go run ./cmd/server
```

Or build the binary:
```bash
go build -o bin/server ./cmd/server
./bin/server
```
