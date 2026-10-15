package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

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
	articles, err := h.store.List(r.Context(), status)
	if err != nil {
		http.Error(w, `{"error":"failed to fetch articles"}`, http.StatusInternalServerError)
		return
	}

	for i := range articles {
		h.ResolveMedia(&articles[i])
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(articles)
}

func (h *ArticleHandler) PublishArticle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/articles/")
	id = strings.TrimSuffix(id, "/publish")
	id = strings.TrimSpace(id)
	if id == "" {
		http.Error(w, `{"error":"article id or slug is required"}`, http.StatusBadRequest)
		return
	}

	publishedBy := "editor"
	art, err := h.store.Publish(r.Context(), id, publishedBy)
	if errors.Is(err, store.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusBadRequest)
		return
	}

	h.ResolveMedia(art)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"article": art,
	})
}

func (h *ArticleHandler) GetArticle(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/api/articles/")
	slug = strings.TrimSuffix(slug, "/publish")
	slug = strings.TrimSpace(slug)
	if slug == "" {
		http.NotFound(w, r)
		return
	}

	art, err := h.store.GetBySlug(r.Context(), slug)
	if errors.Is(err, store.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, `{"error":"failed to fetch article"}`, http.StatusInternalServerError)
		return
	}

	h.ResolveMedia(art)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(art)
}

func (h *ArticleHandler) SaveArticle(w http.ResponseWriter, r *http.Request) {
	var art models.Article
	if err := json.NewDecoder(r.Body).Decode(&art); err != nil {
		http.Error(w, `{"error":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}

	if art.Title == "" {
		http.Error(w, `{"error":"title is required"}`, http.StatusBadRequest)
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
		http.Error(w, `{"error":"failed to save article: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	h.ResolveMedia(&art)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"article": art,
	})
}

func (h *ArticleHandler) ListCandidates(w http.ResponseWriter, r *http.Request) {
	candidates, err := h.store.ListCandidates(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to fetch suggested candidates"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(candidates)
}

func (h *ArticleHandler) GetCandidate(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/candidates/")
	id = strings.TrimSpace(id)
	if id == "" {
		http.NotFound(w, r)
		return
	}

	candidate, err := h.store.GetCandidate(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, `{"error":"failed to fetch candidate"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(candidate)
}

func (h *ArticleHandler) ListSources(w http.ResponseWriter, r *http.Request) {
	sources, err := h.store.ListSources(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to fetch sources"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(sources)
}

func (h *ArticleHandler) ResolveLink(w http.ResponseWriter, r *http.Request) {
	targetURL := r.URL.Query().Get("url")
	if targetURL == "" {
		http.Error(w, `{"error":"url parameter is required"}`, http.StatusBadRequest)
		return
	}

	isInternal := urlutil.IsInternalURL(targetURL)
	slug := urlutil.CleanArticleSlug(targetURL)

	response := map[string]any{
		"url":        targetURL,
		"isInternal": isInternal,
		"slug":       slug,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
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
