package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/okoye-dev/cherryl-media-engine/internal/httputil"
	"github.com/okoye-dev/cherryl-media-engine/internal/models"
	"github.com/okoye-dev/cherryl-media-engine/internal/storage"
	"github.com/okoye-dev/cherryl-media-engine/internal/store"
	"github.com/okoye-dev/cherryl-media-engine/internal/urlutil"
)

type ArticleHandler struct {
	store   store.ArticleStore
	storage storage.Storage
}

func NewArticleHandler(s store.ArticleStore, st storage.Storage) *ArticleHandler {
	return &ArticleHandler{
		store:   s,
		storage: st,
	}
}

// ResolveMedia dynamically builds public URLs using the active storage configuration
// so the database only needs to store the location-independent storage key.
func (h *ArticleHandler) ResolveMedia(art *models.Article) {
	if art == nil || h.storage == nil {
		return
	}

	if art.FeaturedImage.StorageKey != "" {
		art.FeaturedImage.URL = h.storage.GetPublicURL(art.FeaturedImage.StorageKey)
	} else if key := urlutil.ExtractStorageKey(art.FeaturedImage.URL); strings.HasPrefix(key, "media/") {
		art.FeaturedImage.StorageKey = key
		art.FeaturedImage.URL = h.storage.GetPublicURL(key)
	}

	for i := range art.Body {
		block := &art.Body[i]
		if block.Type == "image" || block.Type == "video" {
			if block.StorageKey != "" {
				block.Src = h.storage.GetPublicURL(block.StorageKey)
			} else if key := urlutil.ExtractStorageKey(block.Src); strings.HasPrefix(key, "media/") {
				block.StorageKey = key
				block.Src = h.storage.GetPublicURL(key)
			}
		}
	}
}

// NormalizeMediaKeys ensures storage keys are extracted from URLs so that
// articles remain decoupled from the physical CDN/host location.
func (h *ArticleHandler) NormalizeMediaKeys(art *models.Article) {
	if art.FeaturedImage.StorageKey == "" && art.FeaturedImage.URL != "" {
		key := urlutil.ExtractStorageKey(art.FeaturedImage.URL)
		if strings.HasPrefix(key, "media/") {
			art.FeaturedImage.StorageKey = key
		}
	}
	for i := range art.Body {
		block := &art.Body[i]
		if (block.Type == "image" || block.Type == "video") && block.StorageKey == "" && block.Src != "" {
			key := urlutil.ExtractStorageKey(block.Src)
			if strings.HasPrefix(key, "media/") {
				block.StorageKey = key
			}
		}
	}
}

func (h *ArticleHandler) ListArticles(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	query := r.URL.Query().Get("q")
	if query == "" {
		query = r.URL.Query().Get("search")
	}
	idsParam := r.URL.Query().Get("ids")

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
		articles, err = h.store.GetByIDs(r.Context(), rawIDs)
	} else if query != "" {
		limit := 20
		if lStr := r.URL.Query().Get("limit"); lStr != "" {
			if l, lErr := strconv.Atoi(lStr); lErr == nil && l > 0 {
				limit = l
			}
		}
		articles, err = h.store.Search(r.Context(), query, limit)
	} else {
		articles, err = h.store.List(r.Context(), status)
	}

	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "FETCH_FAILED", "Failed to fetch articles")
		return
	}

	for i := range articles {
		h.ResolveMedia(&articles[i])
	}

	httputil.WriteJSON(w, http.StatusOK, articles)
}

func (h *ArticleHandler) PublishArticle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/articles/")
	id = strings.TrimSuffix(id, "/publish")
	id = strings.TrimSpace(id)
	if id == "" {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_PARAM", "Article ID or slug is required")
		return
	}

	publishedBy := "editor"
	art, err := h.store.Publish(r.Context(), id, publishedBy)
	if errors.Is(err, store.ErrNotFound) {
		httputil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Article not found")
		return
	}
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "PUBLISH_FAILED", err.Error())
		return
	}

	h.ResolveMedia(art)
	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"article": art,
	})
}

func (h *ArticleHandler) GetArticle(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/api/articles/")
	slug = strings.TrimSuffix(slug, "/publish")
	slug = strings.TrimSpace(slug)
	if slug == "" {
		httputil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Article not found")
		return
	}

	art, err := h.store.GetBySlug(r.Context(), slug)
	if errors.Is(err, store.ErrNotFound) {
		// Fallback to lookup by ID
		art, err = h.store.GetByID(r.Context(), slug)
	}
	if errors.Is(err, store.ErrNotFound) {
		httputil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Article not found")
		return
	}
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "FETCH_FAILED", "Failed to fetch article")
		return
	}

	h.ResolveMedia(art)
	httputil.WriteJSON(w, http.StatusOK, art)
}

func (h *ArticleHandler) SaveArticle(w http.ResponseWriter, r *http.Request) {
	var art models.Article
	if err := json.NewDecoder(r.Body).Decode(&art); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON request body")
		return
	}

	if art.Title == "" {
		httputil.WriteError(w, http.StatusBadRequest, "VALIDATION_FAILED", "Post title cannot be empty")
		return
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

	h.NormalizeMediaKeys(&art)

	if err := h.store.Save(r.Context(), &art); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "SAVE_FAILED", "Failed to save article: "+err.Error())
		return
	}

	h.ResolveMedia(&art)
	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"article": art,
	})
}

func (h *ArticleHandler) ListCandidates(w http.ResponseWriter, r *http.Request) {
	candidates, err := h.store.ListCandidates(r.Context())
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "FETCH_FAILED", "Failed to fetch candidates")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, candidates)
}

func (h *ArticleHandler) GetCandidate(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/candidates/")
	id = strings.TrimSpace(id)
	if id == "" {
		httputil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Candidate not found")
		return
	}

	candidate, err := h.store.GetCandidate(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		httputil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Candidate not found")
		return
	}
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "FETCH_FAILED", "Failed to fetch candidate")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, candidate)
}

func (h *ArticleHandler) ListSources(w http.ResponseWriter, r *http.Request) {
	sources, err := h.store.ListSources(r.Context())
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "FETCH_FAILED", "Failed to fetch sources")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, sources)
}

func (h *ArticleHandler) ResolveLink(w http.ResponseWriter, r *http.Request) {
	targetURL := r.URL.Query().Get("url")
	if targetURL == "" {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_PARAM", "URL parameter is required")
		return
	}

	isInternal := urlutil.IsInternalURL(targetURL)
	slug := urlutil.CleanArticleSlug(targetURL)

	response := map[string]any{
		"url":        targetURL,
		"isInternal": isInternal,
		"slug":       slug,
	}

	httputil.WriteJSON(w, http.StatusOK, response)
}

func (h *ArticleHandler) CheckSlug(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimSpace(r.URL.Query().Get("slug"))
	if slug == "" {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_PARAM", "Slug parameter is required")
		return
	}
	excludeID := strings.TrimSpace(r.URL.Query().Get("excludeId"))
	available, err := h.store.CheckSlugAvailable(r.Context(), slug, excludeID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "CHECK_FAILED", "Failed to check slug availability: "+err.Error())
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"slug":      slug,
		"available": available,
	})
}

func (h *ArticleHandler) ListTags(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	tags, err := h.store.ListTags(r.Context(), q)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "FETCH_FAILED", "Failed to list tags: "+err.Error())
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"tags": tags,
	})
}

// GenerateAIContext formats a complete editorial bundle for external LLM prompt context
func FormatAIContext(art *models.Article) string {
	var sb strings.Builder
	sb.WriteString("# EDITORIAL CONTEXT & SOURCE BRIEF FOR FACT CHECKING / REFINEMENT\n\n")
	sb.WriteString(fmt.Sprintf("## TITLE: %s\n", art.Title))
	if art.Deck != "" {
		sb.WriteString(fmt.Sprintf("## DECK / SUBTITLE: %s\n", art.Deck))
	}
	sb.WriteString(fmt.Sprintf("## CATEGORY: %s\n", art.CategoryName))
	sb.WriteString(fmt.Sprintf("## ESTIMATED READ TIME: %s\n\n", art.ReadTime))

	sb.WriteString("### PRIMARY SOURCES & CITATIONS:\n")
	if len(art.Sources) == 0 {
		sb.WriteString("- No external source links attached yet.\n")
	} else {
		for i, s := range art.Sources {
			sb.WriteString(fmt.Sprintf("%d. **%s** (%s)\n   - Source URL: %s\n", i+1, s.SourceName, s.PublishedAt, s.URL))
			if s.ClaimSummary != "" {
				sb.WriteString(fmt.Sprintf("   - Key Claim / Fact: %s\n", s.ClaimSummary))
			}
		}
	}
	sb.WriteString("\n")

	sb.WriteString("### FULL ARTICLE BODY:\n")
	for _, block := range art.Body {
		switch block.Type {
		case "paragraph":
			sb.WriteString(block.Text + "\n\n")
		case "heading":
			sb.WriteString(fmt.Sprintf("#### %s\n\n", block.Text))
		case "quote":
			sb.WriteString(fmt.Sprintf("> \"%s\"\n> — %s (%s)\n\n", block.Quote, block.Author, block.Context))
		case "link":
			sb.WriteString(fmt.Sprintf("[Link Card: %s - %s (%s)]\n\n", block.Title, block.URL, block.Publisher))
		case "image":
			sb.WriteString(fmt.Sprintf("[Image: %s | Caption: %s]\n\n", block.Src, block.Caption))
		}
	}

	sb.WriteString("### SOCIAL MEDIA CHANNEL VARIANTS:\n")
	if art.SocialVariants.X != "" {
		sb.WriteString(fmt.Sprintf("- **X / Twitter**: %s\n", art.SocialVariants.X))
	}
	if art.SocialVariants.Facebook != "" {
		sb.WriteString(fmt.Sprintf("- **Facebook**: %s\n", art.SocialVariants.Facebook))
	}
	if art.SocialVariants.LinkedIn != "" {
		sb.WriteString(fmt.Sprintf("- **LinkedIn**: %s\n", art.SocialVariants.LinkedIn))
	}
	if art.SocialVariants.Newsletter != "" {
		sb.WriteString(fmt.Sprintf("- **Newsletter Blurb**: %s\n", art.SocialVariants.Newsletter))
	}

	sb.WriteString("\n---\n")
	sb.WriteString("### EDITORIAL INSTRUCTIONS FOR AI:\n")
	sb.WriteString("1. Verify consistency between cited primary sources and the claims in the article body.\n")
	sb.WriteString("2. Check for journalistic objectivity and crisp, engaging Nigerian newsroom tone.\n")
	sb.WriteString("3. Highlight any factual inconsistencies, ungrounded speculation, or spelling of Nigerian names and public entities.\n")
	sb.WriteString("4. Provide refined body paragraphs or improved social media angles if appropriate.\n")

	return sb.String()
}
