# ==============================================================================
# Cherryl Media Engine — Multi-Stage Production Dockerfile
# ==============================================================================

# Build Stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Cache Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and database migrations
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY db/ ./db/

# Build statically linked binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/bin/server ./cmd/server

# Runner Stage
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S cherryl && adduser -S cherryl -G cherryl

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/server /app/server
COPY --from=builder /app/db/migrations /app/db/migrations

# Create uploads directory for fallback local file storage
RUN mkdir -p /app/uploads && chown -R cherryl:cherryl /app

USER cherryl

# Expose single port
EXPOSE 8080

ENV PORT=8080

CMD ["/app/server"]
