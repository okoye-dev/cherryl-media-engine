package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/okoye-dev/cherryl-media-engine/internal/models"
)

// ArticleSelectFields defines the canonical SQL column list for selecting articles.
const ArticleSelectFields = `
	id, slug, title, kicker, deck, category, cover_image, author_name, author_role, 
	blocks, sources, social_variants, read_time, status, candidate_id, ai_review_notes, 
	excerpt, original_url, is_lead, urgent, published_at, created_at, tags,
	COALESCE(subtitle, ''), COALESCE(seo_title, ''), COALESCE(seo_slug, ''), 
	COALESCE(seo_description, ''), COALESCE(send_webpushr, false), COALESCE(related_stories, '[]'::jsonb),
	COALESCE(is_featured, false), COALESCE(media_confirmed, false)
`

// RowScanner abstracts single-row scanning across pgx.Row and pgx.Rows.
type RowScanner interface {
	Scan(dest ...any) error
}

// scanArticle extracts a single Article model from an SQL row scanner.
func scanArticle(scanner RowScanner) (*models.Article, error) {
	var (
		id, slug, title, category, status string
		kicker, deck, coverImage, authorName, authorRole, readTime, candidateID, aiNotes, excerpt, origURL *string
		isLead, urgent bool
		blocksJSON, sourcesJSON, socialJSON []byte
		publishedAt, createdAt *time.Time
		tags []string
		subtitle, seoTitle, seoSlug, seoDescription string
		sendWebpushr, isFeatured, mediaConfirmed bool
		relatedJSON []byte
	)

	err := scanner.Scan(
		&id, &slug, &title, &kicker, &deck, &category, &coverImage, &authorName, &authorRole,
		&blocksJSON, &sourcesJSON, &socialJSON, &readTime, &status, &candidateID, &aiNotes,
		&excerpt, &origURL, &isLead, &urgent, &publishedAt, &createdAt, &tags,
		&subtitle, &seoTitle, &seoSlug, &seoDescription, &sendWebpushr, &relatedJSON,
		&isFeatured, &mediaConfirmed,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
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
		IsFeatured:     isFeatured,
		MediaConfirmed: mediaConfirmed,
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

// scanArticles iterates over pgx.Rows and collects Article models.
func scanArticles(rows pgx.Rows) ([]models.Article, error) {
	var articles []models.Article
	for rows.Next() {
		art, err := scanArticle(rows)
		if err != nil {
			return nil, fmt.Errorf("scan article row: %w", err)
		}
		articles = append(articles, *art)
	}
	return articles, rows.Err()
}
