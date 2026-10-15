package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/okoye-dev/cherryl-media-engine/internal/models"
)

var ErrNotFound = errors.New("not found")

type ArticleStore interface {
	List(ctx context.Context) ([]models.Article, error)
	GetBySlug(ctx context.Context, slug string) (*models.Article, error)
	Save(ctx context.Context, a *models.Article) error
}

type UploadStore interface {
	RecordUpload(ctx context.Context, key, name string, size int64, contentType, publicURL string) error
}

type PGStore struct {
	pool *pgxpool.Pool
}

func NewPGStore(pool *pgxpool.Pool) *PGStore {
	return &PGStore{pool: pool}
}

func (s *PGStore) List(ctx context.Context) ([]models.Article, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, slug, title, deck, category, cover_image, blocks, read_time, status, published_at, created_at
		FROM articles
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list articles: %w", err)
	}
	defer rows.Close()

	var articles []models.Article
	for rows.Next() {
		var (
			id, slug, title, category, status string
			deck, coverImage, readTime        *string
			blocksJSON                        []byte
			publishedAt, createdAt            *time.Time
		)

		err := rows.Scan(&id, &slug, &title, &deck, &category, &coverImage, &blocksJSON, &readTime, &status, &publishedAt, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("scan article: %w", err)
		}

		art := models.Article{
			ID:           id,
			Slug:         slug,
			Title:        title,
			Category:     models.PillarCategory(category),
			CategoryName: category,
			Status:       models.PublishingStatus(status),
		}
		if deck != nil {
			art.Deck = *deck
		}
		if coverImage != nil {
			art.FeaturedImage.URL = *coverImage
		}
		if readTime != nil {
			art.ReadTime = *readTime
		}
		if publishedAt != nil {
			art.PublishedAt = *publishedAt
		}
		if createdAt != nil {
			art.CreatedAt = *createdAt
		}
		if len(blocksJSON) > 0 {
			_ = json.Unmarshal(blocksJSON, &art.Body)
		}

		articles = append(articles, art)
	}

	return articles, nil
}

func (s *PGStore) GetBySlug(ctx context.Context, slug string) (*models.Article, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, slug, title, deck, category, cover_image, blocks, read_time, status, published_at, created_at
		FROM articles
		WHERE slug = $1
	`, slug)

	var (
		id, sl, title, category, status string
		deck, coverImage, readTime      *string
		blocksJSON                      []byte
		publishedAt, createdAt          *time.Time
	)

	err := row.Scan(&id, &sl, &title, &deck, &category, &coverImage, &blocksJSON, &readTime, &status, &publishedAt, &createdAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get article %q: %w", slug, err)
	}

	art := &models.Article{
		ID:           id,
		Slug:         sl,
		Title:        title,
		Category:     models.PillarCategory(category),
		CategoryName: category,
		Status:       models.PublishingStatus(status),
	}
	if deck != nil {
		art.Deck = *deck
	}
	if coverImage != nil {
		art.FeaturedImage.URL = *coverImage
	}
	if readTime != nil {
		art.ReadTime = *readTime
	}
	if publishedAt != nil {
		art.PublishedAt = *publishedAt
	}
	if createdAt != nil {
		art.CreatedAt = *createdAt
	}
	if len(blocksJSON) > 0 {
		_ = json.Unmarshal(blocksJSON, &art.Body)
	}

	return art, nil
}

func (s *PGStore) Save(ctx context.Context, a *models.Article) error {
	blocksJSON, err := json.Marshal(a.Body)
	if err != nil {
		blocksJSON = []byte("[]")
	}

	if a.ID == "" {
		a.ID = fmt.Sprintf("art-%d", time.Now().UnixNano())
	}
	if a.CreatedAt.IsZero() {
		a.CreatedAt = time.Now()
	}

	_, err = s.pool.Exec(ctx, `
		INSERT INTO articles (slug, title, deck, category, cover_image, blocks, read_time, status, published_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
		ON CONFLICT (slug) DO UPDATE
		SET title = EXCLUDED.title,
		    deck = EXCLUDED.deck,
		    category = EXCLUDED.category,
		    cover_image = EXCLUDED.cover_image,
		    blocks = EXCLUDED.blocks,
		    read_time = EXCLUDED.read_time,
		    status = EXCLUDED.status,
		    published_at = EXCLUDED.published_at,
		    updated_at = NOW()
	`, a.Slug, a.Title, a.Deck, string(a.Category), a.FeaturedImage.URL, blocksJSON, a.ReadTime, string(a.Status), a.PublishedAt, a.CreatedAt)

	if err != nil {
		return fmt.Errorf("upsert article %q: %w", a.Slug, err)
	}
	return nil
}

func (s *PGStore) RecordUpload(ctx context.Context, key, name string, size int64, contentType, publicURL string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO uploads (storage_key, file_name, size_bytes, content_type, public_url, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (storage_key) DO UPDATE
		SET public_url = EXCLUDED.public_url
	`, key, name, size, contentType, publicURL)
	if err != nil {
		return fmt.Errorf("record upload: %w", err)
	}
	return nil
}

type MemoryStore struct {
	mu       sync.RWMutex
	articles map[string]models.Article
	uploads  map[string]string
}

func NewMemoryStore() *MemoryStore {
	m := &MemoryStore{
		articles: make(map[string]models.Article),
		uploads:  make(map[string]string),
	}
	// Initial seed
	m.articles["federal-executive-council-approves-unified-energy-framework"] = models.Article{
		ID:           "sample-01",
		Slug:         "federal-executive-council-approves-unified-energy-framework",
		Title:        "Federal Executive Council Approves Unified Energy and Distribution Guidelines",
		Kicker:       "NATIONAL POLICY ANNOUNCEMENT",
		Deck:         "Presidential council formalizes revised grid codes, state electricity market transitions, and sovereign payment assurance frameworks.",
		Category:     models.CategoryCurrentAffairs,
		CategoryName: "Nigerian Current Affairs",
		Author: models.Author{
			Name: "Bisi Alabi",
			Role: "Public Service & Utility Reporter",
		},
		FeaturedImage: models.FeaturedImage{
			URL:     "/assets/images/sample.webp",
			Caption: "Federal Executive Council in session, Abuja.",
			Credit:  "State House Photo",
		},
		ReadTime:    "4 min read",
		Status:      models.StatusPublished,
		PublishedAt: time.Now(),
		CreatedAt:   time.Now(),
	}
	return m
}

func (m *MemoryStore) List(ctx context.Context) ([]models.Article, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make([]models.Article, 0, len(m.articles))
	for _, a := range m.articles {
		res = append(res, a)
	}
	return res, nil
}

func (m *MemoryStore) GetBySlug(ctx context.Context, slug string) (*models.Article, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	a, ok := m.articles[slug]
	if !ok {
		return nil, ErrNotFound
	}
	return &a, nil
}

func (m *MemoryStore) Save(ctx context.Context, a *models.Article) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if a.ID == "" {
		a.ID = fmt.Sprintf("mem-%d", time.Now().UnixNano())
	}
	m.articles[a.Slug] = *a
	return nil
}

func (m *MemoryStore) RecordUpload(ctx context.Context, key, name string, size int64, contentType, publicURL string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.uploads[key] = publicURL
	return nil
}
