package services

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/okoye-dev/cherryl-media-engine/internal/models"
	"github.com/okoye-dev/cherryl-media-engine/internal/store"
	"github.com/okoye-dev/cherryl-media-engine/internal/urlutil"
)

// ArticleService defines the application service operations for articles.
type ArticleService interface {
	ListArticles(ctx context.Context, status, query, idsParam string, limit int) ([]models.Article, error)
	GetArticle(ctx context.Context, slugOrID string) (*models.Article, error)
	SaveArticle(ctx context.Context, art *models.Article) (*models.Article, error)
	PublishArticle(ctx context.Context, idOrSlug string, publishedBy string) (*models.Article, error)
	CheckSlug(ctx context.Context, slug, excludeID string) (bool, error)
	ListTags(ctx context.Context, query string) ([]string, error)
	ResolveLink(targetURL string) map[string]any
	ListCandidates(ctx context.Context) ([]models.SuggestedStory, error)
	GetCandidate(ctx context.Context, id string) (*models.SuggestedStory, error)
	ListSources(ctx context.Context) ([]models.Source, error)
	SetFeatured(ctx context.Context, idOrSlug string, isFeatured bool) (*models.Article, error)
	GetFeatured(ctx context.Context) (*models.Article, error)
	SearchCuratedArticles(ctx context.Context, category, query, sort string, limit int) ([]models.Article, error)
	MediaResolver() MediaResolver
}

type articleService struct {
	store           store.ArticleStore
	uploadStore     store.UploadStore
	subscriberStore store.SubscriberStore
	resolver        MediaResolver
}

// NewArticleService constructs a new ArticleService with injected store and media resolver.
func NewArticleService(s store.ArticleStore, resolver MediaResolver) ArticleService {
	svc := &articleService{
		store:    s,
		resolver: resolver,
	}
	if u, ok := s.(store.UploadStore); ok {
		svc.uploadStore = u
	}
	if sub, ok := s.(store.SubscriberStore); ok {
		svc.subscriberStore = sub
	}
	return svc
}

func (s *articleService) MediaResolver() MediaResolver {
	return s.resolver
}

func (s *articleService) ListArticles(ctx context.Context, status, query, idsParam string, limit int) ([]models.Article, error) {
	var articles []models.Article
	var err error

	if idsParam != "" {
		var rawIDs []string
		for _, id := range strings.Split(idsParam, ",") {
			trimmed := strings.TrimSpace(id)
			if trimmed != "" {
				rawIDs = append(rawIDs, trimmed)
			}
		}
		articles, err = s.store.GetByIDs(ctx, rawIDs)
	} else if query != "" {
		if limit <= 0 {
			limit = 20
		}
		articles, err = s.store.Search(ctx, query, limit)
	} else {
		articles, err = s.store.List(ctx, status)
	}

	if err != nil {
		return nil, err
	}

	if s.resolver != nil {
		s.resolver.ResolveArticles(articles)
	}
	return articles, nil
}

func (s *articleService) GetArticle(ctx context.Context, slugOrID string) (*models.Article, error) {
	art, err := s.store.GetBySlug(ctx, slugOrID)
	if err == nil && art != nil {
		if s.resolver != nil {
			s.resolver.ResolveMedia(art)
		}
		return art, nil
	}

	art, err = s.store.GetByID(ctx, slugOrID)
	if err != nil {
		return nil, err
	}
	if s.resolver != nil {
		s.resolver.ResolveMedia(art)
	}
	return art, nil
}

func (s *articleService) SaveArticle(ctx context.Context, art *models.Article) (*models.Article, error) {
	if art == nil {
		return nil, errors.New("article cannot be nil")
	}
	if art.Title == "" {
		return nil, errors.New("post title cannot be empty")
	}

	if art.Slug == "" {
		art.Slug = strings.ToLower(strings.ReplaceAll(art.Title, " ", "-"))
	}

	if art.ID == "" {
		art.ID = "art-" + strconv.FormatInt(time.Now().UnixMilli(), 10)
	}

	if art.CreatedAt.IsZero() {
		art.CreatedAt = time.Now()
	}
	if art.PublishedAt.IsZero() {
		art.PublishedAt = time.Now()
	}

	// Status defaults to draft if not set
	if art.Status == "" {
		art.Status = models.StatusDraft
	}

	if s.resolver != nil {
		s.resolver.NormalizeMediaKeys(art)
	}

	if err := s.store.Save(ctx, art); err != nil {
		return nil, err
	}

	if s.resolver != nil {
		s.resolver.ResolveMedia(art)
	}

	return art, nil
}

func (s *articleService) PublishArticle(ctx context.Context, idOrSlug string, publishedBy string) (*models.Article, error) {
	if publishedBy == "" {
		publishedBy = "editor"
	}

	// 1. Fetch the target article
	art, err := s.GetArticle(ctx, idOrSlug)
	if err != nil {
		return nil, err
	}

	// 2. Media confirmation check
	// "before anything is published, media needs to be confirmed uploaded."
	if s.uploadStore != nil {
		var keysToCheck []string
		if art.FeaturedImage.StorageKey != "" {
			keysToCheck = append(keysToCheck, art.FeaturedImage.StorageKey)
		} else if key := urlutil.ExtractStorageKey(art.FeaturedImage.URL); strings.HasPrefix(key, "media/") {
			keysToCheck = append(keysToCheck, key)
		}

		for _, b := range art.Body {
			if b.Type == "image" || b.Type == "video" {
				if b.StorageKey != "" {
					keysToCheck = append(keysToCheck, b.StorageKey)
				} else if key := urlutil.ExtractStorageKey(b.Src); strings.HasPrefix(key, "media/") {
					keysToCheck = append(keysToCheck, key)
				}
			}
		}

		for _, key := range keysToCheck {
			confirmed, checkErr := s.uploadStore.IsConfirmed(ctx, key)
			if checkErr != nil || !confirmed {
				return nil, fmt.Errorf("cannot publish article: media asset %q is not confirmed uploaded", key)
			}
		}
	}

	// 3. Mark media confirmed and publish
	art.MediaConfirmed = true
	publishedArt, err := s.store.Publish(ctx, idOrSlug, publishedBy)
	if err != nil {
		return nil, err
	}

	if s.resolver != nil {
		s.resolver.ResolveMedia(publishedArt)
	}

	// 4. Post email notification dispatch
	// "we should be able to have emails that need a post sent to them when post is made. all this should be modelled in the DB."
	if s.subscriberStore != nil {
		go func(a models.Article) {
			bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			subs, err := s.subscriberStore.ListSubscribers(bgCtx, true)
			if err == nil && len(subs) > 0 {
				var targetEmails []string
				cat := strings.ToLower(string(a.Category))
				for _, sub := range subs {
					if len(sub.Categories) == 0 {
						targetEmails = append(targetEmails, sub.Email)
						continue
					}
					for _, c := range sub.Categories {
						if strings.EqualFold(c, cat) || strings.EqualFold(c, a.CategoryName) {
							targetEmails = append(targetEmails, sub.Email)
							break
						}
					}
				}
				if len(targetEmails) > 0 {
					_ = s.subscriberStore.RecordDispatches(bgCtx, a.ID, targetEmails)
				}
			}
		}(*publishedArt)
	}

	return publishedArt, nil
}

func (s *articleService) SetFeatured(ctx context.Context, idOrSlug string, isFeatured bool) (*models.Article, error) {
	art, err := s.GetArticle(ctx, idOrSlug)
	if err != nil {
		return nil, err
	}
	art.IsFeatured = isFeatured
	if err := s.store.Save(ctx, art); err != nil {
		return nil, err
	}
	if s.resolver != nil {
		s.resolver.ResolveMedia(art)
	}
	return art, nil
}

// GetFeatured implements the decision engine:
// 1. Manually pinned featured article (is_featured = true, published_at DESC)
// 2. Otherwise automated fallback: most recent published story with is_lead = true
// 3. Otherwise latest published article with cover photo & deck
func (s *articleService) GetFeatured(ctx context.Context) (*models.Article, error) {
	published, err := s.store.List(ctx, "published")
	if err != nil {
		return nil, err
	}
	if len(published) == 0 {
		return nil, store.ErrNotFound
	}

	// 1. Check for manually flagged featured story
	for _, art := range published {
		if art.IsFeatured {
			featured := art
			if s.resolver != nil {
				s.resolver.ResolveMedia(&featured)
			}
			return &featured, nil
		}
	}

	// 2. Check for lead story
	for _, art := range published {
		if art.IsLead {
			lead := art
			if s.resolver != nil {
				s.resolver.ResolveMedia(&lead)
			}
			return &lead, nil
		}
	}

	// 3. Automated decision engine: highest editorial completeness
	for _, art := range published {
		if art.FeaturedImage.URL != "" && (art.Deck != "" || art.Subtitle != "") {
			complete := art
			if s.resolver != nil {
				s.resolver.ResolveMedia(&complete)
			}
			return &complete, nil
		}
	}

	// Fallback to first published story
	fallback := published[0]
	if s.resolver != nil {
		s.resolver.ResolveMedia(&fallback)
	}
	return &fallback, nil
}

func (s *articleService) SearchCuratedArticles(ctx context.Context, category, query, sort string, limit int) ([]models.Article, error) {
	if limit <= 0 {
		limit = 30
	}

	// Fetch all articles matching status
	allArticles, err := s.store.List(ctx, "all")
	if err != nil {
		return nil, err
	}

	q := strings.ToLower(strings.TrimSpace(query))
	cat := strings.ToLower(strings.TrimSpace(category))

	var filtered []models.Article
	for _, art := range allArticles {
		// Category match
		if cat != "" && !strings.EqualFold(string(art.Category), cat) && !strings.EqualFold(art.CategoryName, cat) {
			continue
		}

		// Query match
		if q != "" {
			inTitle := strings.Contains(strings.ToLower(art.Title), q)
			inKicker := strings.Contains(strings.ToLower(art.Kicker), q)
			inDeck := strings.Contains(strings.ToLower(art.Deck), q) || strings.Contains(strings.ToLower(art.Subtitle), q)
			inExcerpt := strings.Contains(strings.ToLower(art.Excerpt), q)
			if !inTitle && !inKicker && !inDeck && !inExcerpt {
				continue
			}
		}

		filtered = append(filtered, art)
	}

	// Sort by recency (published_at DESC, or created_at DESC)
	for i := 0; i < len(filtered)-1; i++ {
		for j := i + 1; j < len(filtered); j++ {
			tI := filtered[i].PublishedAt
			if tI.IsZero() {
				tI = filtered[i].CreatedAt
			}
			tJ := filtered[j].PublishedAt
			if tJ.IsZero() {
				tJ = filtered[j].CreatedAt
			}
			if tJ.After(tI) {
				filtered[i], filtered[j] = filtered[j], filtered[i]
			}
		}
	}

	if len(filtered) > limit {
		filtered = filtered[:limit]
	}

	if s.resolver != nil {
		s.resolver.ResolveArticles(filtered)
	}

	return filtered, nil
}

func (s *articleService) CheckSlug(ctx context.Context, slug, excludeID string) (bool, error) {
	return s.store.CheckSlugAvailable(ctx, slug, excludeID)
}

func (s *articleService) ListTags(ctx context.Context, query string) ([]string, error) {
	return s.store.ListTags(ctx, query)
}

func (s *articleService) ResolveLink(targetURL string) map[string]any {
	isInternal := urlutil.IsInternalURL(targetURL)
	slug := urlutil.CleanArticleSlug(targetURL)
	return map[string]any{
		"url":        targetURL,
		"isInternal": isInternal,
		"slug":       slug,
	}
}

func (s *articleService) ListCandidates(ctx context.Context) ([]models.SuggestedStory, error) {
	return s.store.ListCandidates(ctx)
}

func (s *articleService) GetCandidate(ctx context.Context, id string) (*models.SuggestedStory, error) {
	return s.store.GetCandidate(ctx, id)
}

func (s *articleService) ListSources(ctx context.Context) ([]models.Source, error) {
	return s.store.ListSources(ctx)
}
