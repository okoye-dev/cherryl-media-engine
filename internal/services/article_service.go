package services

import (
	"context"
	"errors"
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
	MediaResolver() MediaResolver
}

type articleService struct {
	store    store.ArticleStore
	resolver MediaResolver
}

// NewArticleService constructs a new ArticleService with injected store and media resolver.
func NewArticleService(s store.ArticleStore, resolver MediaResolver) ArticleService {
	return &articleService{
		store:    s,
		resolver: resolver,
	}
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
	art, err := s.store.Publish(ctx, idOrSlug, publishedBy)
	if err != nil {
		return nil, err
	}
	if s.resolver != nil {
		s.resolver.ResolveMedia(art)
	}
	return art, nil
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
