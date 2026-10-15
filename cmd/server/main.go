package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/okoye-dev/cherryl-media-engine/internal/config"
	"github.com/okoye-dev/cherryl-media-engine/internal/handlers"
	"github.com/okoye-dev/cherryl-media-engine/internal/migrate"
	"github.com/okoye-dev/cherryl-media-engine/internal/seed"
	"github.com/okoye-dev/cherryl-media-engine/internal/storage"
	"github.com/okoye-dev/cherryl-media-engine/internal/store"
)

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	cfg := config.Load()

	// S3 Storage setup (S3, Cloudflare R2, MinIO)
	s3Storage, err := storage.NewS3Storage(&cfg.S3)
	if err != nil {
		log.Fatalf("failed to initialize S3 storage: %v", err)
	}

	// Database setup (PostgreSQL / Supabase)
	var articleStore store.ArticleStore
	var uploadStore store.UploadStore
	var userStore store.UserStore

	if cfg.Database.URL != "" {
		ctx := context.Background()
		pool, err := pgxpool.New(ctx, cfg.Database.URL)
		if err != nil {
			log.Fatalf("failed to connect to postgres: %v", err)
		}
		defer pool.Close()

		if err := pool.Ping(ctx); err != nil {
			log.Fatalf("failed to ping postgres: %v", err)
		}
		log.Println("connected to PostgreSQL successfully")

		if err := migrate.Run(ctx, pool); err != nil {
			log.Fatalf("migration failed: %v", err)
		}

		pg := store.NewPGStore(pool)
		articleStore = pg
		uploadStore = pg
		userStore = pg

		// Seed PostgreSQL with canonical articles and default admin if not already present
		if err := seed.Run(ctx, pg); err != nil {
			log.Printf("initial database seeding warning: %v", err)
		}
	} else {
		mem := store.NewMemoryStore()
		articleStore = mem
		uploadStore = mem
		userStore = mem

		// Seed Memory store
		if err := seed.Run(context.Background(), mem); err != nil {
			log.Printf("initial memory seeding warning: %v", err)
		}
	}

	mux := http.NewServeMux()
	articleHandler := handlers.NewArticleHandler(articleStore, s3Storage)
	uploadHandler := handlers.NewUploadHandler(s3Storage, uploadStore)
	authHandler := handlers.NewAuthHandler(userStore, cfg.JWTSecret)

	// Authentication routes
	mux.HandleFunc("POST /api/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/auth/logout", authHandler.Logout)
	mux.HandleFunc("GET /api/auth/me", authHandler.Me)

	// Article routes
	mux.HandleFunc("GET /api/articles", articleHandler.ListArticles)
	mux.HandleFunc("GET /api/articles/", articleHandler.GetArticle)
	mux.HandleFunc("POST /api/articles", articleHandler.SaveArticle)
	mux.HandleFunc("POST /api/articles/{id}/publish", articleHandler.PublishArticle)
	mux.HandleFunc("PUT /api/articles/{id}/publish", articleHandler.PublishArticle)
	mux.HandleFunc("GET /api/links/resolve", articleHandler.ResolveLink)

	// AI discovery candidates and sources routes
	mux.HandleFunc("GET /api/candidates", articleHandler.ListCandidates)
	mux.HandleFunc("GET /api/candidates/", articleHandler.GetCandidate)
	mux.HandleFunc("GET /api/sources", articleHandler.ListSources)

	// Presigned S3 upload & confirm landing routes
	mux.HandleFunc("POST /api/uploads/presign", uploadHandler.PresignUpload)
	mux.HandleFunc("POST /api/uploads/confirm", uploadHandler.ConfirmUpload)
	mux.HandleFunc("POST /api/files/presign-upload", uploadHandler.PresignUpload)
	mux.HandleFunc("PUT /api/uploads/direct", uploadHandler.DirectUpload)
	mux.HandleFunc("POST /api/uploads/direct", uploadHandler.DirectUpload)

	// Local uploaded files static server (fallback)
	_ = os.MkdirAll("./uploads", 0755)
	mux.Handle("GET /uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	// Health check
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		status := map[string]any{
			"status":   "online",
			"service":  "cherryl-media-engine",
			"s3":       s3Storage.IsConfigured(),
			"database": cfg.Database.URL != "",
		}
		_, _ = w.Write([]byte(`{"status":"online","service":"cherryl-media-engine"}`))
		_ = status
	})

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: withCORS(mux),
	}

	log.Printf("Cherryl Media Engine running on http://localhost:%s\n", cfg.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}
}
