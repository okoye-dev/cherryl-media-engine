package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/okoye-dev/cherryl-media-engine/internal/models"
	"github.com/okoye-dev/cherryl-media-engine/internal/store"
	"github.com/okoye-dev/cherryl-media-engine/internal/urlutil"
)

// ArticleService defines the application service operations for articles.
type ArticleService interface {
	// ListArticlesPage returns one page of articles and the total number that match.
	ListArticlesPage(ctx context.Context, status, query, idsParam string, limit, offset int) ([]models.Article, int, error)
	GetArticle(ctx context.Context, slugOrID string) (*models.Article, error)
	SaveArticle(ctx context.Context, art *models.Article) (*models.Article, error)
	// PatchArticle changes only the fields in patch. It returns the saved story and the names of the fields that really changed.
	PatchArticle(ctx context.Context, slugOrID string, patch map[string]json.RawMessage) (*models.Article, []string, error)
	PublishArticle(ctx context.Context, idOrSlug string, publishedBy string) (*models.Article, error)
	CheckSlug(ctx context.Context, slug, excludeID string) (bool, error)
	ListTags(ctx context.Context, query string) ([]string, error)
	ResolveLink(targetURL string) map[string]any
	ListCandidatesPage(ctx context.Context, limit, offset int) ([]models.SuggestedStory, int, error)
	GetCandidate(ctx context.Context, id string) (*models.SuggestedStory, error)
	ListSources(ctx context.Context) ([]models.Source, error)
	SetFeatured(ctx context.Context, idOrSlug string, isFeatured bool) (*models.Article, error)
	GetFeatured(ctx context.Context) (*models.Article, error)
	SearchCuratedArticles(ctx context.Context, category, query, sort string, limit, offset int) ([]models.Article, int, error)
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

func (s *articleService) ListArticlesPage(ctx context.Context, status, query, idsParam string, limit, offset int) ([]models.Article, int, error) {
	var articles []models.Article
	var total int
	var err error

	switch {
	case idsParam != "":
		var rawIDs []string
		for _, id := range strings.Split(idsParam, ",") {
			if trimmed := strings.TrimSpace(id); trimmed != "" {
				rawIDs = append(rawIDs, trimmed)
			}
		}
		// A lookup by id is bounded by the ids asked for, so it is not paged.
		articles, err = s.store.GetByIDs(ctx, rawIDs)
		total = len(articles)
	case query != "":
		articles, total, err = s.store.SearchPage(ctx, query, limit, offset)
	default:
		articles, total, err = s.store.ListPage(ctx, status, limit, offset)
	}
	if err != nil {
		return nil, 0, err
	}
	if s.resolver != nil {
		s.resolver.ResolveArticles(articles)
	}
	return articles, total, nil
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
		art.ID = uuid.New().String()
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

// Fields a patch may never change. The id and creation time belong to the database.
var immutableArticleFields = map[string]bool{"id": true, "createdAt": true}

// ErrPatchRejected marks a patch that was refused for a reason the person can act on.
var ErrPatchRejected = errors.New("patch rejected")

func (s *articleService) PatchArticle(ctx context.Context, slugOrID string, patch map[string]json.RawMessage) (*models.Article, []string, error) {
	cur, err := s.GetArticle(ctx, slugOrID)
	if err != nil {
		return nil, nil, err
	}
	curRaw, err := json.Marshal(cur)
	if err != nil {
		return nil, nil, err
	}
	merged := map[string]json.RawMessage{}
	if err := json.Unmarshal(curRaw, &merged); err != nil {
		return nil, nil, err
	}

	// Put in only what really differs from what is stored.
	var changed []string
	for k, v := range patch {
		if immutableArticleFields[k] {
			continue
		}
		if old, ok := merged[k]; ok && jsonEqual(old, v) {
			continue
		}
		merged[k] = v
		changed = append(changed, k)
	}
	sort.Strings(changed)
	if len(changed) == 0 {
		return cur, nil, nil
	}

	// Going live has its own checks (the cover photo must be uploaded), so a patch cannot skip them.
	if raw, ok := patch["status"]; ok {
		var st string
		_ = json.Unmarshal(raw, &st)
		if st == string(models.StatusPublished) && cur.Status != models.StatusPublished {
			return nil, nil, fmt.Errorf("%w: use the publish step to take a story live", ErrPatchRejected)
		}
	}

	mergedRaw, err := json.Marshal(merged)
	if err != nil {
		return nil, nil, err
	}
	var next models.Article
	if err := json.Unmarshal(mergedRaw, &next); err != nil {
		return nil, nil, fmt.Errorf("%w: a field has the wrong type: %v", ErrPatchRejected, err)
	}
	next.ID = cur.ID
	next.CreatedAt = cur.CreatedAt

	// A new slug renames the story. Save alone would create a second copy, so rename first.
	if next.Slug != "" && next.Slug != cur.Slug {
		ok, err := s.store.CheckSlugAvailable(ctx, next.Slug, cur.ID)
		if err != nil {
			return nil, nil, err
		}
		if !ok {
			return nil, nil, fmt.Errorf("%w: the address %q is already used by another story", ErrPatchRejected, next.Slug)
		}
		if err := s.store.UpdateSlug(ctx, cur.ID, next.Slug); err != nil {
			return nil, nil, err
		}
	}

	saved, err := s.SaveArticle(ctx, &next)
	if err != nil {
		return nil, nil, err
	}
	return saved, changed, nil
}

func jsonEqual(a, b json.RawMessage) bool {
	var x, y any
	if json.Unmarshal(a, &x) != nil || json.Unmarshal(b, &y) != nil {
		return string(a) == string(b)
	}
	xb, _ := json.Marshal(x)
	yb, _ := json.Marshal(y)
	return string(xb) == string(yb)
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

func (s *articleService) SearchCuratedArticles(ctx context.Context, category, query, sortBy string, limit, offset int) ([]models.Article, int, error) {
	if limit <= 0 {
		limit = 30
	}

	// Fetch all articles matching status
	allArticles, err := s.store.List(ctx, "all")
	if err != nil {
		return nil, 0, err
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

	// Newest first (published date, or created date for drafts)
	when := func(a models.Article) time.Time {
		if a.PublishedAt.IsZero() {
			return a.CreatedAt
		}
		return a.PublishedAt
	}
	sort.SliceStable(filtered, func(i, j int) bool { return when(filtered[i]).After(when(filtered[j])) })

	total := len(filtered)
	if offset >= total {
		filtered = []models.Article{}
	} else {
		end := offset + limit
		if end > total {
			end = total
		}
		filtered = filtered[offset:end]
	}

	if s.resolver != nil {
		s.resolver.ResolveArticles(filtered)
	}

	return filtered, total, nil
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

func (s *articleService) ListCandidatesPage(ctx context.Context, limit, offset int) ([]models.SuggestedStory, int, error) {
	return s.store.ListCandidatesPage(ctx, limit, offset)
}

func (s *articleService) GetCandidate(ctx context.Context, id string) (*models.SuggestedStory, error) {
	return s.store.GetCandidate(ctx, id)
}

func (s *articleService) ListSources(ctx context.Context) ([]models.Source, error) {
	return s.store.ListSources(ctx)
}
