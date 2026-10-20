package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/okoye-dev/cherryl-media-engine/internal/auth"
	"github.com/okoye-dev/cherryl-media-engine/internal/models"
)

var ErrNotFound = errors.New("not found")

type ArticleStore interface {
	List(ctx context.Context, status string) ([]models.Article, error)
	GetBySlug(ctx context.Context, slug string) (*models.Article, error)
	GetByID(ctx context.Context, id string) (*models.Article, error)
	GetByIDs(ctx context.Context, ids []string) ([]models.Article, error)
	Search(ctx context.Context, query string, limit int) ([]models.Article, error)
	Save(ctx context.Context, a *models.Article) error
	Publish(ctx context.Context, idOrSlug string, publishedBy string) (*models.Article, error)
	CheckSlugAvailable(ctx context.Context, slug string, excludeID string) (bool, error)
	ListTags(ctx context.Context, query string) ([]string, error)
	ListCandidates(ctx context.Context) ([]models.SuggestedStory, error)
	GetCandidate(ctx context.Context, id string) (*models.SuggestedStory, error)
	SaveCandidate(ctx context.Context, s *models.SuggestedStory) error
	ListSources(ctx context.Context) ([]models.Source, error)
}

type UserStore interface {
	GetUserByEmail(ctx context.Context, email string) (*auth.User, string, error)
	GetUserByID(ctx context.Context, id string) (*auth.User, error)
	CreateUser(ctx context.Context, u *auth.User, passwordHash string) error
}

type UploadStore interface {
	RecordUpload(ctx context.Context, key, name string, size int64, contentType, publicURL string) error
	ConfirmUpload(ctx context.Context, key string, size int64, publicURL string) error
}

type SettingsStore interface {
	GetSetting(ctx context.Context, key string) ([]byte, error)
	SetSetting(ctx context.Context, key string, val []byte) error
	GetAllSettings(ctx context.Context) (map[string]any, error)
}

type SectionStore interface {
	GetSection(ctx context.Context, sectionKey string) (*models.CuratedSection, error)
	GetAllSections(ctx context.Context) ([]models.CuratedSection, error)
	SaveSection(ctx context.Context, sectionKey string, articleIDs []string, metadata map[string]any) error
}

type PGStore struct {
	pool *pgxpool.Pool
}

func NewPGStore(pool *pgxpool.Pool) *PGStore {
	return &PGStore{pool: pool}
}

const articleSelectFields = `
	id, slug, title, kicker, deck, category, cover_image, author_name, author_role, 
	blocks, sources, social_variants, read_time, status, candidate_id, ai_review_notes, 
	excerpt, original_url, is_lead, urgent, published_at, created_at, tags,
	COALESCE(subtitle, ''), COALESCE(seo_title, ''), COALESCE(seo_slug, ''), 
	COALESCE(seo_description, ''), COALESCE(send_webpushr, false), COALESCE(related_stories, '[]'::jsonb)
`

func scanArticleRows(rows pgx.Rows) ([]models.Article, error) {
	var articles []models.Article
	for rows.Next() {
		var (
			id, slug, title, category, status string
			kicker, deck, coverImage, authorName, authorRole, readTime, candidateID, aiNotes, excerpt, origURL *string
			isLead, urgent bool
			blocksJSON, sourcesJSON, socialJSON []byte
			publishedAt, createdAt *time.Time
			tags []string
			subtitle, seoTitle, seoSlug, seoDescription string
			sendWebpushr bool
			relatedJSON []byte
		)

		err := rows.Scan(
			&id, &slug, &title, &kicker, &deck, &category, &coverImage, &authorName, &authorRole, 
			&blocksJSON, &sourcesJSON, &socialJSON, &readTime, &status, &candidateID, &aiNotes, 
			&excerpt, &origURL, &isLead, &urgent, &publishedAt, &createdAt, &tags,
			&subtitle, &seoTitle, &seoSlug, &seoDescription, &sendWebpushr, &relatedJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("scan article row: %w", err)
		}

		art := models.Article{
			ID:             id,
			Slug:           slug,
			Title:          title,
			Category:       models.PillarCategory(category),
			CategoryName:   category,
			Status:         models.PublishingStatus(status),
			Tags:           tags,
			IsLead:         isLead,
			Urgent:         urgent,
			Subtitle:       subtitle,
			SEOTitle:       seoTitle,
			SEOSlug:        seoSlug,
			SEODescription: seoDescription,
			SendWebpushr:   sendWebpushr,
		}
		if kicker != nil {
			art.Kicker = *kicker
		}
		if deck != nil {
			art.Deck = *deck
		}
		if art.Subtitle == "" && deck != nil {
			art.Subtitle = *deck
		}
		if art.Deck == "" && subtitle != "" {
			art.Deck = subtitle
		}
		if coverImage != nil {
			art.FeaturedImage.URL = *coverImage
		}
		if authorName != nil {
			art.Author.Name = *authorName
		}
		if authorRole != nil {
			art.Author.Role = *authorRole
		}
		if readTime != nil {
			art.ReadTime = *readTime
		}
		if candidateID != nil {
			art.CandidateID = *candidateID
		}
		if aiNotes != nil {
			art.AIReviewNotes = *aiNotes
		}
		if excerpt != nil {
			art.Excerpt = *excerpt
		}
		if origURL != nil {
			art.OriginalURL = *origURL
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
		if len(sourcesJSON) > 0 {
			_ = json.Unmarshal(sourcesJSON, &art.Sources)
		}
		if len(socialJSON) > 0 {
			_ = json.Unmarshal(socialJSON, &art.SocialVariants)
		}
		if len(relatedJSON) > 0 {
			_ = json.Unmarshal(relatedJSON, &art.RelatedStories)
		}

		articles = append(articles, art)
	}

	return articles, rows.Err()
}

func scanArticleRow(row pgx.Row, identifier string) (*models.Article, error) {
	var (
		id, slug, title, category, status string
		kicker, deck, coverImage, authorName, authorRole, readTime, candidateID, aiNotes, excerpt, origURL *string
		isLead, urgent bool
		blocksJSON, sourcesJSON, socialJSON []byte
		publishedAt, createdAt *time.Time
		tags []string
		subtitle, seoTitle, seoSlug, seoDescription string
		sendWebpushr bool
		relatedJSON []byte
	)

	err := row.Scan(
		&id, &slug, &title, &kicker, &deck, &category, &coverImage, &authorName, &authorRole, 
		&blocksJSON, &sourcesJSON, &socialJSON, &readTime, &status, &candidateID, &aiNotes, 
		&excerpt, &origURL, &isLead, &urgent, &publishedAt, &createdAt, &tags,
		&subtitle, &seoTitle, &seoSlug, &seoDescription, &sendWebpushr, &relatedJSON,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan article %q: %w", identifier, err)
	}

	art := &models.Article{
		ID:             id,
		Slug:           slug,
		Title:          title,
		Category:       models.PillarCategory(category),
		CategoryName:   category,
		Status:         models.PublishingStatus(status),
		Tags:           tags,
		IsLead:         isLead,
		Urgent:         urgent,
		Subtitle:       subtitle,
		SEOTitle:       seoTitle,
		SEOSlug:        seoSlug,
		SEODescription: seoDescription,
		SendWebpushr:   sendWebpushr,
	}
	if kicker != nil {
		art.Kicker = *kicker
	}
	if deck != nil {
		art.Deck = *deck
	}
	if art.Subtitle == "" && deck != nil {
		art.Subtitle = *deck
	}
	if art.Deck == "" && subtitle != "" {
		art.Deck = subtitle
	}
	if coverImage != nil {
		art.FeaturedImage.URL = *coverImage
	}
	if authorName != nil {
		art.Author.Name = *authorName
	}
	if authorRole != nil {
		art.Author.Role = *authorRole
	}
	if readTime != nil {
		art.ReadTime = *readTime
	}
	if candidateID != nil {
		art.CandidateID = *candidateID
	}
	if aiNotes != nil {
		art.AIReviewNotes = *aiNotes
	}
	if excerpt != nil {
		art.Excerpt = *excerpt
	}
	if origURL != nil {
		art.OriginalURL = *origURL
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
	if len(sourcesJSON) > 0 {
		_ = json.Unmarshal(sourcesJSON, &art.Sources)
	}
	if len(socialJSON) > 0 {
		_ = json.Unmarshal(socialJSON, &art.SocialVariants)
	}
	if len(relatedJSON) > 0 {
		_ = json.Unmarshal(relatedJSON, &art.RelatedStories)
	}

	return art, nil
}

func (s *PGStore) List(ctx context.Context, status string) ([]models.Article, error) {
	baseQuery := "SELECT " + articleSelectFields + " FROM articles"
	var rows pgx.Rows
	var err error

	if status == "" || status == "published" {
		rows, err = s.pool.Query(ctx, baseQuery+" WHERE status = 'published' AND published_at <= NOW() ORDER BY published_at DESC")
	} else if status == "all" {
		rows, err = s.pool.Query(ctx, baseQuery+" ORDER BY created_at DESC")
	} else {
		rows, err = s.pool.Query(ctx, baseQuery+" WHERE status = $1 ORDER BY created_at DESC", status)
	}
	if err != nil {
		return nil, fmt.Errorf("list articles: %w", err)
	}
	defer rows.Close()

	return scanArticleRows(rows)
}

func (s *PGStore) GetBySlug(ctx context.Context, slug string) (*models.Article, error) {
	row := s.pool.QueryRow(ctx, "SELECT "+articleSelectFields+" FROM articles WHERE slug = $1", slug)
	return scanArticleRow(row, slug)
}

func (s *PGStore) GetByID(ctx context.Context, id string) (*models.Article, error) {
	row := s.pool.QueryRow(ctx, "SELECT "+articleSelectFields+" FROM articles WHERE id = $1", id)
	return scanArticleRow(row, id)
}

func (s *PGStore) GetByIDs(ctx context.Context, ids []string) ([]models.Article, error) {
	if len(ids) == 0 {
		return []models.Article{}, nil
	}
	rows, err := s.pool.Query(ctx, "SELECT "+articleSelectFields+" FROM articles WHERE id = ANY($1)", ids)
	if err != nil {
		return nil, fmt.Errorf("get articles by ids: %w", err)
	}
	defer rows.Close()

	articles, err := scanArticleRows(rows)
	if err != nil {
		return nil, err
	}

	byID := make(map[string]models.Article, len(articles))
	for _, a := range articles {
		byID[a.ID] = a
	}

	ordered := make([]models.Article, 0, len(ids))
	for _, id := range ids {
		if a, ok := byID[id]; ok {
			ordered = append(ordered, a)
		}
	}
	return ordered, nil
}

func (s *PGStore) Search(ctx context.Context, query string, limit int) ([]models.Article, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return s.List(ctx, "published")
	}

	pattern := "%" + trimmed + "%"
	rows, err := s.pool.Query(ctx, "SELECT "+articleSelectFields+`
		FROM articles
		WHERE title ILIKE $1 OR kicker ILIKE $1 OR deck ILIKE $1 OR category ILIKE $1 OR excerpt ILIKE $1 OR subtitle ILIKE $1 OR $2 = ANY(tags)
		ORDER BY published_at DESC LIMIT $3
	`, pattern, trimmed, limit)
	if err != nil {
		return nil, fmt.Errorf("search articles: %w", err)
	}
	defer rows.Close()

	return scanArticleRows(rows)
}

func (s *PGStore) CheckSlugAvailable(ctx context.Context, slug string, excludeID string) (bool, error) {
	var count int
	var err error
	if excludeID != "" {
		err = s.pool.QueryRow(ctx, "SELECT count(*) FROM articles WHERE LOWER(slug) = LOWER($1) AND id != $2::uuid", slug, excludeID).Scan(&count)
	} else {
		err = s.pool.QueryRow(ctx, "SELECT count(*) FROM articles WHERE LOWER(slug) = LOWER($1)", slug).Scan(&count)
	}
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

func (s *PGStore) ListTags(ctx context.Context, query string) ([]string, error) {
	var rows pgx.Rows
	var err error
	trimmed := strings.TrimSpace(query)
	if trimmed != "" {
		pattern := "%" + trimmed + "%"
		rows, err = s.pool.Query(ctx, `
			SELECT DISTINCT tag 
			FROM (SELECT unnest(tags) AS tag FROM articles) t 
			WHERE tag ILIKE $1 
			ORDER BY tag ASC LIMIT 50
		`, pattern)
	} else {
		rows, err = s.pool.Query(ctx, `
			SELECT DISTINCT tag 
			FROM (SELECT unnest(tags) AS tag FROM articles) t 
			WHERE tag != '' 
			ORDER BY tag ASC LIMIT 100
		`)
	}
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err == nil && t != "" {
			tags = append(tags, t)
		}
	}
	return tags, nil
}

func (s *PGStore) Save(ctx context.Context, a *models.Article) error {
	blocksBytes, _ := json.Marshal(a.Body)
	if a.Body == nil {
		blocksBytes = []byte("[]")
	}
	sourcesBytes, _ := json.Marshal(a.Sources)
	if a.Sources == nil {
		sourcesBytes = []byte("[]")
	}
	socialBytes, _ := json.Marshal(a.SocialVariants)
	relatedBytes, _ := json.Marshal(a.RelatedStories)
	if a.RelatedStories == nil {
		relatedBytes = []byte("[]")
	}

	var artID string
	if _, err := uuid.Parse(a.ID); err == nil {
		artID = a.ID
	}
	if a.CreatedAt.IsZero() {
		a.CreatedAt = time.Now()
	}

	tags := a.Tags
	if tags == nil {
		tags = []string{}
	}

	if a.Subtitle != "" && a.Deck == "" {
		a.Deck = a.Subtitle
	} else if a.Deck != "" && a.Subtitle == "" {
		a.Subtitle = a.Deck
	}

	_, err := s.pool.Exec(ctx, `
		INSERT INTO articles (
			id, slug, title, kicker, deck, category, cover_image, author_name, author_role, 
			blocks, sources, social_variants, read_time, status, candidate_id, ai_review_notes, 
			excerpt, original_url, is_lead, urgent, published_at, created_at, updated_at, tags,
			subtitle, seo_title, seo_slug, seo_description, send_webpushr, related_stories
		)
		VALUES (
			COALESCE(NULLIF($1, '')::uuid, gen_random_uuid()), $2, $3, $4, $5, $6, $7, $8, $9, 
			$10::jsonb, $11::jsonb, $12::jsonb, $13, $14, NULLIF($15, '')::uuid, $16, 
			$17, $18, $19, $20, $21, $22, NOW(), $23,
			$24, $25, $26, $27, $28, $29::jsonb
		)
		ON CONFLICT (slug) DO UPDATE
		SET title = EXCLUDED.title,
		    subtitle = EXCLUDED.subtitle,
		    kicker = EXCLUDED.kicker,
		    deck = EXCLUDED.deck,
		    category = EXCLUDED.category,
		    cover_image = EXCLUDED.cover_image,
		    author_name = EXCLUDED.author_name,
		    author_role = EXCLUDED.author_role,
		    blocks = EXCLUDED.blocks,
		    sources = EXCLUDED.sources,
		    social_variants = EXCLUDED.social_variants,
		    read_time = EXCLUDED.read_time,
		    status = EXCLUDED.status,
		    tags = EXCLUDED.tags,
		    excerpt = EXCLUDED.excerpt,
		    original_url = EXCLUDED.original_url,
		    is_lead = EXCLUDED.is_lead,
		    urgent = EXCLUDED.urgent,
		    ai_review_notes = EXCLUDED.ai_review_notes,
		    seo_title = EXCLUDED.seo_title,
		    seo_slug = EXCLUDED.seo_slug,
		    seo_description = EXCLUDED.seo_description,
		    send_webpushr = EXCLUDED.send_webpushr,
		    related_stories = EXCLUDED.related_stories,
		    published_at = EXCLUDED.published_at,
		    updated_at = NOW()
	`, artID, a.Slug, a.Title, a.Kicker, a.Deck, string(a.Category), a.FeaturedImage.URL, a.Author.Name, a.Author.Role, string(blocksBytes), string(sourcesBytes), string(socialBytes), a.ReadTime, string(a.Status), a.CandidateID, a.AIReviewNotes, a.Excerpt, a.OriginalURL, a.IsLead, a.Urgent, a.PublishedAt, a.CreatedAt, tags, a.Subtitle, a.SEOTitle, a.SEOSlug, a.SEODescription, a.SendWebpushr, string(relatedBytes))

	if err != nil {
		return fmt.Errorf("upsert article %q: %w", a.Slug, err)
	}
	return nil
}

func (s *PGStore) ListCandidates(ctx context.Context) ([]models.SuggestedStory, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, headline, pitch, category, relevance_score, sources, extracted_facts, suggested_draft, status, article_id, created_at
		FROM suggested_stories
		ORDER BY relevance_score DESC, created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list suggested stories: %w", err)
	}
	defer rows.Close()

	var list []models.SuggestedStory
	for rows.Next() {
		var (
			id, headline, category, status string
			pitch, articleID               *string
			relScore                       float64
			sourcesJSON, factsJSON, draftJSON []byte
			createdAt                      time.Time
		)

		err := rows.Scan(&id, &headline, &pitch, &category, &relScore, &sourcesJSON, &factsJSON, &draftJSON, &status, &articleID, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("scan suggested story: %w", err)
		}

		item := models.SuggestedStory{
			ID:             id,
			Headline:       headline,
			Category:       category,
			RelevanceScore: relScore,
			Status:         status,
			CreatedAt:      createdAt,
		}
		if pitch != nil {
			item.Pitch = *pitch
		}
		if articleID != nil {
			item.ArticleID = *articleID
		}
		if len(sourcesJSON) > 0 {
			_ = json.Unmarshal(sourcesJSON, &item.Sources)
		}
		if len(factsJSON) > 0 {
			_ = json.Unmarshal(factsJSON, &item.ExtractedFacts)
		}
		if len(draftJSON) > 0 {
			_ = json.Unmarshal(draftJSON, &item.SuggestedDraft)
		}

		list = append(list, item)
	}
	return list, nil
}

func (s *PGStore) GetCandidate(ctx context.Context, id string) (*models.SuggestedStory, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, headline, pitch, category, relevance_score, sources, extracted_facts, suggested_draft, status, article_id, created_at
		FROM suggested_stories
		WHERE id = $1
	`, id)

	var (
		cId, headline, category, status string
		pitch, articleID               *string
		relScore                       float64
		sourcesJSON, factsJSON, draftJSON []byte
		createdAt                      time.Time
	)

	err := row.Scan(&cId, &headline, &pitch, &category, &relScore, &sourcesJSON, &factsJSON, &draftJSON, &status, &articleID, &createdAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get candidate: %w", err)
	}

	item := &models.SuggestedStory{
		ID:             cId,
		Headline:       headline,
		Category:       category,
		RelevanceScore: relScore,
		Status:         status,
		CreatedAt:      createdAt,
	}
	if pitch != nil {
		item.Pitch = *pitch
	}
	if articleID != nil {
		item.ArticleID = *articleID
	}
	if len(sourcesJSON) > 0 {
		_ = json.Unmarshal(sourcesJSON, &item.Sources)
	}
	if len(factsJSON) > 0 {
		_ = json.Unmarshal(factsJSON, &item.ExtractedFacts)
	}
	if len(draftJSON) > 0 {
		_ = json.Unmarshal(draftJSON, &item.SuggestedDraft)
	}

	return item, nil
}

func (s *PGStore) SaveCandidate(ctx context.Context, sStory *models.SuggestedStory) error {
	sourcesJSON, _ := json.Marshal(sStory.Sources)
	factsJSON, _ := json.Marshal(sStory.ExtractedFacts)
	draftJSON, _ := json.Marshal(sStory.SuggestedDraft)

	_, err := s.pool.Exec(ctx, `
		INSERT INTO suggested_stories (headline, pitch, category, relevance_score, sources, extracted_facts, suggested_draft, status, updated_at)
		VALUES ($1, $2, $3, $4, $5::jsonb, $6::jsonb, $7::jsonb, $8, NOW())
	`, sStory.Headline, sStory.Pitch, sStory.Category, sStory.RelevanceScore, string(sourcesJSON), string(factsJSON), string(draftJSON), sStory.Status)
	return err
}

func (s *PGStore) ListSources(ctx context.Context) ([]models.Source, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, type, feed_url, site_url, category_hint, reliability_score, active, last_fetched_at, created_at
		FROM sources
		WHERE active = true
		ORDER BY reliability_score DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sources []models.Source
	for rows.Next() {
		var src models.Source
		var siteURL, catHint *string
		var lastFetched *time.Time
		err := rows.Scan(&src.ID, &src.Name, &src.Type, &src.FeedURL, &siteURL, &catHint, &src.ReliabilityScore, &src.Active, &lastFetched, &src.CreatedAt)
		if err != nil {
			return nil, err
		}
		if siteURL != nil {
			src.SiteURL = *siteURL
		}
		if catHint != nil {
			src.CategoryHint = *catHint
		}
		if lastFetched != nil {
			src.LastFetchedAt = *lastFetched
		}
		sources = append(sources, src)
	}
	return sources, nil
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

func (s *PGStore) ConfirmUpload(ctx context.Context, key string, size int64, publicURL string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO uploads (storage_key, file_name, size_bytes, content_type, public_url, created_at)
		VALUES ($1, $1, $2, 'application/octet-stream', $3, NOW())
		ON CONFLICT (storage_key) DO UPDATE
		SET size_bytes = $2, public_url = $3
	`, key, size, publicURL)
	if err != nil {
		return fmt.Errorf("confirm upload: %w", err)
	}
	return nil
}

func (s *PGStore) Publish(ctx context.Context, idOrSlug string, publishedBy string) (*models.Article, error) {
	art, err := s.GetBySlug(ctx, idOrSlug)
	if err != nil {
		row := s.pool.QueryRow(ctx, "SELECT slug FROM articles WHERE id::text = $1", idOrSlug)
		var sl string
		if err := row.Scan(&sl); err != nil {
			return nil, ErrNotFound
		}
		art, err = s.GetBySlug(ctx, sl)
		if err != nil {
			return nil, err
		}
	}

	if art.Title == "" {
		return nil, errors.New("cannot publish article without a title")
	}
	if art.FeaturedImage.URL == "" {
		return nil, errors.New("cannot publish article without a cover photograph")
	}

	now := time.Now().UTC()
	var pubByUUID *string
	if _, err := uuid.Parse(publishedBy); err == nil {
		pubByUUID = &publishedBy
	}

	_, err = s.pool.Exec(ctx, `
		UPDATE articles
		SET status = 'published', published_at = $1, published_by = NULLIF($2, '')::uuid, updated_at = $1
		WHERE id::text = $3 OR slug = $3
	`, now, pubByUUID, idOrSlug)
	if err != nil {
		return nil, fmt.Errorf("publishing article: %w", err)
	}

	art.Status = models.StatusPublished
	art.PublishedAt = now
	art.PublishedBy = publishedBy
	return art, nil
}

func (s *PGStore) GetUserByEmail(ctx context.Context, email string) (*auth.User, string, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, email, password_hash, full_name, role, COALESCE(avatar_url, ''), active, created_at
		FROM users
		WHERE LOWER(email) = LOWER($1)
	`, email)

	var u auth.User
	var hash string
	err := row.Scan(&u.ID, &u.Email, &hash, &u.FullName, &u.Role, &u.AvatarURL, &u.Active, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, "", ErrNotFound
	}
	if err != nil {
		return nil, "", fmt.Errorf("get user by email: %w", err)
	}
	return &u, hash, nil
}

func (s *PGStore) GetUserByID(ctx context.Context, id string) (*auth.User, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, email, full_name, role, COALESCE(avatar_url, ''), active, created_at
		FROM users
		WHERE id = $1
	`, id)

	var u auth.User
	err := row.Scan(&u.ID, &u.Email, &u.FullName, &u.Role, &u.AvatarURL, &u.Active, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &u, nil
}

func (s *PGStore) CreateUser(ctx context.Context, u *auth.User, passwordHash string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, role, avatar_url, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
		ON CONFLICT (email) DO UPDATE
		SET password_hash = $3, full_name = $4, role = $5, active = $7
	`, u.ID, u.Email, passwordHash, u.FullName, u.Role, u.AvatarURL, u.Active, u.CreatedAt)
	return err
}

func (s *PGStore) GetSetting(ctx context.Context, key string) ([]byte, error) {
	row := s.pool.QueryRow(ctx, "SELECT value FROM site_settings WHERE key = $1", key)
	var val []byte
	err := row.Scan(&val)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return val, err
}

func (s *PGStore) SetSetting(ctx context.Context, key string, val []byte) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO site_settings (key, value, updated_at)
		VALUES ($1, $2::jsonb, NOW())
		ON CONFLICT (key) DO UPDATE
		SET value = EXCLUDED.value, updated_at = NOW()
	`, key, string(val))
	return err
}

func (s *PGStore) GetAllSettings(ctx context.Context) (map[string]any, error) {
	rows, err := s.pool.Query(ctx, "SELECT key, value FROM site_settings")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make(map[string]any)
	for rows.Next() {
		var k string
		var raw []byte
		if err := rows.Scan(&k, &raw); err != nil {
			return nil, err
		}
		var parsed any
		if err := json.Unmarshal(raw, &parsed); err == nil {
			res[k] = parsed
		} else {
			res[k] = string(raw)
		}
	}
	return res, rows.Err()
}

func (s *PGStore) GetSection(ctx context.Context, sectionKey string) (*models.CuratedSection, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT section_key, article_ids, metadata, updated_at
		FROM curated_sections
		WHERE section_key = $1
	`, sectionKey)

	var (
		key        string
		articleIDs []string
		metaJSON   []byte
		updatedAt  time.Time
	)

	err := row.Scan(&key, &articleIDs, &metaJSON, &updatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get section %q: %w", sectionKey, err)
	}

	meta := make(map[string]any)
	if len(metaJSON) > 0 {
		_ = json.Unmarshal(metaJSON, &meta)
	}

	return &models.CuratedSection{
		SectionKey: key,
		ArticleIDs: articleIDs,
		Metadata:   meta,
		UpdatedAt:  updatedAt,
	}, nil
}

func (s *PGStore) GetAllSections(ctx context.Context) ([]models.CuratedSection, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT section_key, article_ids, metadata, updated_at
		FROM curated_sections
		ORDER BY section_key ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("get all sections: %w", err)
	}
	defer rows.Close()

	var sections []models.CuratedSection
	for rows.Next() {
		var (
			key        string
			articleIDs []string
			metaJSON   []byte
			updatedAt  time.Time
		)
		if err := rows.Scan(&key, &articleIDs, &metaJSON, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan section: %w", err)
		}
		meta := make(map[string]any)
		if len(metaJSON) > 0 {
			_ = json.Unmarshal(metaJSON, &meta)
		}
		sections = append(sections, models.CuratedSection{
			SectionKey: key,
			ArticleIDs: articleIDs,
			Metadata:   meta,
			UpdatedAt:  updatedAt,
		})
	}
	return sections, rows.Err()
}

func (s *PGStore) SaveSection(ctx context.Context, sectionKey string, articleIDs []string, metadata map[string]any) error {
	if articleIDs == nil {
		articleIDs = []string{}
	}
	metaJSON, err := json.Marshal(metadata)
	if err != nil {
		metaJSON = []byte("{}")
	}

	_, err = s.pool.Exec(ctx, `
		INSERT INTO curated_sections (section_key, article_ids, metadata, updated_at)
		VALUES ($1, $2, $3::jsonb, NOW())
		ON CONFLICT (section_key) DO UPDATE
		SET article_ids = EXCLUDED.article_ids,
		    metadata = EXCLUDED.metadata,
		    updated_at = NOW()
	`, sectionKey, articleIDs, string(metaJSON))
	return err
}

type MemoryStore struct {
	mu         sync.RWMutex
	articles   map[string]models.Article
	candidates map[string]models.SuggestedStory
	sources    []models.Source
	uploads    map[string]string
	users      map[string]*auth.User
	passwords  map[string]string
	settings   map[string][]byte
	sections   map[string]models.CuratedSection
}

func NewMemoryStore() *MemoryStore {
	m := &MemoryStore{
		articles:   make(map[string]models.Article),
		candidates: make(map[string]models.SuggestedStory),
		uploads:    make(map[string]string),
		users:      make(map[string]*auth.User),
		passwords:  make(map[string]string),
		settings:   make(map[string][]byte),
		sections:   make(map[string]models.CuratedSection),
	}

	// Default curated sections
	m.sections["hero_lead"] = models.CuratedSection{
		SectionKey: "hero_lead",
		ArticleIDs: []string{"2f405add-8bba-5261-99af-15feaa62754b"},
		UpdatedAt:  time.Now(),
	}
	m.sections["hero_trending"] = models.CuratedSection{
		SectionKey: "hero_trending",
		ArticleIDs: []string{
			"35588574-ec88-5a81-be90-8c46283c24de",
			"decd08ac-ef71-54ee-ac58-fd2480c007fa",
			"f10fad9c-c601-552e-a7b2-816b8c370ecc",
		},
		UpdatedAt: time.Now(),
	}
	m.sections["featured_updates"] = models.CuratedSection{
		SectionKey: "featured_updates",
		ArticleIDs: []string{
			"25386df5-95e6-50ad-bc4c-3b6db589992d",
			"12f9e7fb-8019-5edb-80e9-7789a9729176",
			"d2065825-0408-5a7f-b4d0-559be9e67a9a",
			"278198f8-3eee-55ce-8736-dcdd55c46cec",
		},
		UpdatedAt: time.Now(),
	}
	m.sections["perspectives"] = models.CuratedSection{
		SectionKey: "perspectives",
		ArticleIDs: []string{
			"278198f8-3eee-55ce-8736-dcdd55c46cec",
			"d2065825-0408-5a7f-b4d0-559be9e67a9a",
			"f10fad9c-c601-552e-a7b2-816b8c370ecc",
		},
		UpdatedAt: time.Now(),
	}
	m.sections["dossier_lead"] = models.CuratedSection{
		SectionKey: "dossier_lead",
		ArticleIDs: []string{"f3033f02-9628-52d2-bef4-8b0814c6d835"},
		UpdatedAt:  time.Now(),
	}
	m.sections["dossier_soundscapes"] = models.CuratedSection{
		SectionKey: "dossier_soundscapes",
		ArticleIDs: []string{"d641aeae-a6be-50f5-a13c-54c67bbf6e75"},
		UpdatedAt:  time.Now(),
	}
	m.sections["dossier_gastronomy"] = models.CuratedSection{
		SectionKey: "dossier_gastronomy",
		ArticleIDs: []string{"0827a7d0-015b-5643-ac3d-4ccced2b933c"},
		UpdatedAt:  time.Now(),
	}
	m.sections["dossier_radar"] = models.CuratedSection{
		SectionKey: "dossier_radar",
		ArticleIDs: []string{
			"f2fc6ae3-16ec-51cc-bc8d-c9d2b4e4271f",
			"ee3b36cb-8ce0-5f3c-bfae-3d077b882004",
			"917d0345-9020-50e5-89eb-9cb1721d21fa",
			"33a00b82-fd1f-5437-9317-0c951151d023",
		},
		UpdatedAt: time.Now(),
	}

	// Default development admin credentials
	initPass := os.Getenv("INITIAL_ADMIN_PASSWORD")
	if initPass == "" {
		initPass = "cherry-" + uuid.New().String()[:8]
	}
	adminHash, _ := auth.HashPassword(initPass)
	adminUser := &auth.User{
		ID:        "usr-admin-01",
		Email:     "admin@cherrylmedia.com",
		FullName:  "Newsroom Editor",
		Role:      "admin",
		Active:    true,
		CreatedAt: time.Now().UTC(),
	}
	m.users[adminUser.Email] = adminUser
	m.passwords[adminUser.Email] = adminHash

	// Sample article
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
		Sources: []models.SourceCitation{
			{
				SourceName:   "State House Official Press Release",
				URL:          "https://statehouse.gov.ng/releases/energy-framework-2026",
				PublishedAt:  "2026-10-14",
				ClaimSummary: "Approval of revised multi-year tariff framework and sovereign backing for state electricity boards.",
			},
			{
				SourceName:   "Federal Ministry of Power Briefing",
				URL:          "https://power.gov.ng/news/grid-code-revisions",
				PublishedAt:  "2026-10-14",
				ClaimSummary: "Technical guidelines for decentralized distribution licensing across 36 states.",
			},
		},
		SocialVariants: models.SocialVariants{
			X:          "Nigeria’s Federal Executive Council has finalized unified energy and electricity distribution guidelines. Here is what state governments and commercial power consumers need to know:",
			Facebook:   "Key update for Nigerian consumers: The Federal Executive Council has officially endorsed the new national electricity and distribution framework.",
			LinkedIn:   "The Federal Executive Council has formalized revised grid codes and decentralized distribution licensing guidelines. Read our full regulatory breakdown for institutions and investors.",
			Newsletter: "Electricity Governance Overhaul: What the new FEC energy guidelines mean for state utility transition.",
		},
		ReadTime:    "4 min read",
		Status:      models.StatusPublished,
		PublishedAt: time.Now(),
		CreatedAt:   time.Now(),
	}

	// Sample AI-discovered story candidates
	m.candidates["cand-01"] = models.SuggestedStory{
		ID:             "cand-01",
		Headline:       "NERC Releases Revised Q4 Tariff Schedule for Commercial and Industrial Feeders",
		Pitch:          "Tariff adjustments focused on manufacturing clusters in Ikeja, Agbara, and Aba to protect production margins.",
		Category:       "nigerian-current-affairs",
		RelevanceScore: 0.94,
		Sources: []models.SourceCitation{
			{
				SourceName:   "NERC Regulatory Order #2026/089",
				URL:          "https://nerc.gov.ng/orders/q4-tariff-update",
				PublishedAt:  "2026-10-15T08:30:00Z",
				ClaimSummary: "Rate stabilization clause activated for industrial clusters consuming >50MW monthly.",
			},
			{
				SourceName:   "Manufacturers Association of Nigeria (MAN)",
				URL:          "https://manufacturersnigeria.org/press/energy-costs",
				PublishedAt:  "2026-10-15T09:15:00Z",
				ClaimSummary: "MAN commendation on targeted power subsidy for key agro-processing zones.",
			},
		},
		ExtractedFacts: []string{
			"Targeted industrial clusters will receive capped transmission charges through December.",
			"State electricity regulatory boards (SERCs) in Lagos, Kaduna, and Enugu assume supervisory authority on Nov 1.",
			"DisCos instructed to deploy smart meters to 100% of industrial feeder points within 60 days.",
		},
		Status:    "discovered",
		CreatedAt: time.Now().Add(-2 * time.Hour),
	}

	m.candidates["cand-02"] = models.SuggestedStory{
		ID:             "cand-02",
		Headline:       "Aviation Ministry Mandates Single-Window Customs Clearance for Lagos & Abuja Cargo",
		Pitch:          "Reduces cargo dwell time from 14 days to 48 hours for perishable and agricultural exports.",
		Category:       "business-institutions",
		RelevanceScore: 0.88,
		Sources: []models.SourceCitation{
			{
				SourceName:   "Federal Airports Authority of Nigeria (FAAN)",
				URL:          "https://faan.gov.ng/cargo-clearance-guidelines",
				PublishedAt:  "2026-10-15T07:00:00Z",
				ClaimSummary: "Digital portal integration with Customs and Quarantine services starting Monday.",
			},
		},
		ExtractedFacts: []string{
			"Single-window system integrates Customs (NCS), Quarantine (NAQS), and FAAN inspection into one physical stop.",
			"Targeting dwell times under 48 hours for perishable agricultural air exports.",
		},
		Status:    "discovered",
		CreatedAt: time.Now().Add(-5 * time.Hour),
	}

	m.sources = []models.Source{
		{
			ID:               "src-01",
			Name:             "State House Newsroom",
			Type:             "rss",
			FeedURL:          "https://statehouse.gov.ng/feed",
			SiteURL:          "https://statehouse.gov.ng",
			CategoryHint:     "nigerian-current-affairs",
			ReliabilityScore: 0.98,
			Active:           true,
			LastFetchedAt:    time.Now().Add(-15 * time.Minute),
			CreatedAt:        time.Now(),
		},
		{
			ID:               "src-02",
			Name:             "Central Bank of Nigeria",
			Type:             "api",
			FeedURL:          "https://cbn.gov.ng/api/press-releases",
			SiteURL:          "https://cbn.gov.ng",
			CategoryHint:     "business-institutions",
			ReliabilityScore: 0.95,
			Active:           true,
			LastFetchedAt:    time.Now().Add(-30 * time.Minute),
			CreatedAt:        time.Now(),
		},
		{
			ID:               "src-03",
			Name:             "Punch RSS",
			Type:             "rss",
			FeedURL:          "https://punchng.com/feed",
			SiteURL:          "https://punchng.com",
			CategoryHint:     "nigerian-current-affairs",
			ReliabilityScore: 0.85,
			Active:           true,
			LastFetchedAt:    time.Now().Add(-10 * time.Minute),
			CreatedAt:        time.Now(),
		},
	}

	return m
}

func (m *MemoryStore) List(ctx context.Context, status string) ([]models.Article, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make([]models.Article, 0, len(m.articles))
	for _, a := range m.articles {
		if status == "" || status == "published" {
			if a.Status == models.StatusPublished {
				res = append(res, a)
			}
		} else if status == "all" || string(a.Status) == status {
			res = append(res, a)
		}
	}
	return res, nil
}

func (m *MemoryStore) GetBySlug(ctx context.Context, slug string) (*models.Article, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	a, ok := m.articles[slug]
	if !ok {
		// Also allow searching by ID
		for _, art := range m.articles {
			if art.ID == slug {
				return &art, nil
			}
		}
		return nil, ErrNotFound
	}
	return &a, nil
}

func (m *MemoryStore) GetByID(ctx context.Context, id string) (*models.Article, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, art := range m.articles {
		if art.ID == id {
			return &art, nil
		}
	}
	return nil, ErrNotFound
}

func (m *MemoryStore) GetByIDs(ctx context.Context, ids []string) ([]models.Article, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	byID := make(map[string]models.Article, len(m.articles))
	for _, art := range m.articles {
		byID[art.ID] = art
	}
	ordered := make([]models.Article, 0, len(ids))
	for _, id := range ids {
		if art, ok := byID[id]; ok {
			ordered = append(ordered, art)
		}
	}
	return ordered, nil
}

func (m *MemoryStore) Search(ctx context.Context, query string, limit int) ([]models.Article, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if limit <= 0 {
		limit = 20
	}
	q := strings.ToLower(strings.TrimSpace(query))
	var matches []models.Article
	for _, art := range m.articles {
		if q == "" {
			matches = append(matches, art)
		} else if strings.Contains(strings.ToLower(art.Title), q) ||
			strings.Contains(strings.ToLower(art.Kicker), q) ||
			strings.Contains(strings.ToLower(art.Deck), q) ||
			strings.Contains(strings.ToLower(string(art.Category)), q) ||
			strings.Contains(strings.ToLower(art.Excerpt), q) {
			matches = append(matches, art)
		}
		if len(matches) >= limit {
			break
		}
	}
	return matches, nil
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

func (m *MemoryStore) Publish(ctx context.Context, idOrSlug string, publishedBy string) (*models.Article, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var targetKey string
	for slug, a := range m.articles {
		if slug == idOrSlug || a.ID == idOrSlug {
			targetKey = slug
			break
		}
	}
	if targetKey == "" {
		return nil, ErrNotFound
	}

	art := m.articles[targetKey]
	if art.Title == "" {
		return nil, errors.New("cannot publish article without a title")
	}
	if art.FeaturedImage.URL == "" {
		return nil, errors.New("cannot publish article without a cover photograph")
	}

	art.Status = models.StatusPublished
	art.PublishedAt = time.Now().UTC()
	art.PublishedBy = publishedBy
	m.articles[targetKey] = art
	return &art, nil
}

func (m *MemoryStore) CheckSlugAvailable(ctx context.Context, slug string, excludeID string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	lower := strings.ToLower(slug)
	for _, a := range m.articles {
		if strings.ToLower(a.Slug) == lower {
			if excludeID != "" && a.ID == excludeID {
				continue
			}
			return false, nil
		}
	}
	return true, nil
}

func (m *MemoryStore) ListTags(ctx context.Context, query string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	tagSet := make(map[string]struct{})
	lower := strings.ToLower(strings.TrimSpace(query))
	for _, a := range m.articles {
		for _, t := range a.Tags {
			if t == "" {
				continue
			}
			if lower == "" || strings.Contains(strings.ToLower(t), lower) {
				tagSet[t] = struct{}{}
			}
		}
	}
	res := make([]string, 0, len(tagSet))
	for t := range tagSet {
		res = append(res, t)
	}
	sort.Strings(res)
	return res, nil
}

func (m *MemoryStore) ListCandidates(ctx context.Context) ([]models.SuggestedStory, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make([]models.SuggestedStory, 0, len(m.candidates))
	for _, c := range m.candidates {
		res = append(res, c)
	}
	return res, nil
}

func (m *MemoryStore) GetCandidate(ctx context.Context, id string) (*models.SuggestedStory, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, ok := m.candidates[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &c, nil
}

func (m *MemoryStore) SaveCandidate(ctx context.Context, sStory *models.SuggestedStory) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if sStory.ID == "" {
		sStory.ID = fmt.Sprintf("cand-%d", time.Now().UnixNano())
	}
	m.candidates[sStory.ID] = *sStory
	return nil
}

func (m *MemoryStore) ListSources(ctx context.Context) ([]models.Source, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sources, nil
}

func (m *MemoryStore) RecordUpload(ctx context.Context, key, name string, size int64, contentType, publicURL string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.uploads[key] = publicURL
	return nil
}

func (m *MemoryStore) ConfirmUpload(ctx context.Context, key string, size int64, publicURL string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.uploads[key] = publicURL
	return nil
}

func (m *MemoryStore) GetUserByEmail(ctx context.Context, email string) (*auth.User, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	u, ok := m.users[email]
	if !ok {
		return nil, "", ErrNotFound
	}
	return u, m.passwords[email], nil
}

func (m *MemoryStore) GetUserByID(ctx context.Context, id string) (*auth.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, ErrNotFound
}

func (m *MemoryStore) CreateUser(ctx context.Context, u *auth.User, passwordHash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users[u.Email] = u
	m.passwords[u.Email] = passwordHash
	return nil
}

func (m *MemoryStore) GetSetting(ctx context.Context, key string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.settings[key]
	if !ok {
		return nil, ErrNotFound
	}
	return val, nil
}

func (m *MemoryStore) SetSetting(ctx context.Context, key string, val []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.settings[key] = val
	return nil
}

func (m *MemoryStore) GetAllSettings(ctx context.Context) (map[string]any, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make(map[string]any)
	for k, raw := range m.settings {
		var parsed any
		if err := json.Unmarshal(raw, &parsed); err == nil {
			res[k] = parsed
		} else {
			res[k] = string(raw)
		}
	}
	return res, nil
}

func (m *MemoryStore) GetSection(ctx context.Context, sectionKey string) (*models.CuratedSection, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	sec, ok := m.sections[sectionKey]
	if !ok {
		return nil, ErrNotFound
	}
	return &sec, nil
}

func (m *MemoryStore) GetAllSections(ctx context.Context) ([]models.CuratedSection, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make([]models.CuratedSection, 0, len(m.sections))
	for _, s := range m.sections {
		res = append(res, s)
	}
	return res, nil
}

func (m *MemoryStore) SaveSection(ctx context.Context, sectionKey string, articleIDs []string, metadata map[string]any) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if articleIDs == nil {
		articleIDs = []string{}
	}
	if metadata == nil {
		metadata = make(map[string]any)
	}
	m.sections[sectionKey] = models.CuratedSection{
		SectionKey: sectionKey,
		ArticleIDs: articleIDs,
		Metadata:   metadata,
		UpdatedAt:  time.Now(),
	}
	return nil
}

