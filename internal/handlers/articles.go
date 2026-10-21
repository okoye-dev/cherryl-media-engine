package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/okoye-dev/cherryl-media-engine/internal/httputil"
	"github.com/okoye-dev/cherryl-media-engine/internal/models"
	"github.com/okoye-dev/cherryl-media-engine/internal/services"
	"github.com/okoye-dev/cherryl-media-engine/internal/storage"
	"github.com/okoye-dev/cherryl-media-engine/internal/store"
)

type ArticleHandler struct {
	service services.ArticleService
}

// NewArticleHandler constructs an ArticleHandler from store and storage for backwards compatibility.
func NewArticleHandler(s store.ArticleStore, st storage.Storage) *ArticleHandler {
	resolver := services.NewMediaResolver(st)
	svc := services.NewArticleService(s, resolver)
	return &ArticleHandler{service: svc}
}

// NewArticleHandlerWithService constructs an ArticleHandler with an injected ArticleService.
func NewArticleHandlerWithService(svc services.ArticleService) *ArticleHandler {
	return &ArticleHandler{service: svc}
}

func (h *ArticleHandler) ResolveMedia(art *models.Article) {
	if h.service != nil && h.service.MediaResolver() != nil {
		h.service.MediaResolver().ResolveMedia(art)
	}
}

func (h *ArticleHandler) NormalizeMediaKeys(art *models.Article) {
	if h.service != nil && h.service.MediaResolver() != nil {
		h.service.MediaResolver().NormalizeMediaKeys(art)
	}
}

func (h *ArticleHandler) GetFeatured(w http.ResponseWriter, r *http.Request) {
	featured, err := h.service.GetFeatured(r.Context())
	if errors.Is(err, store.ErrNotFound) {
		httputil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "No featured article found")
		return
	}
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "FETCH_FAILED", err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, featured)
}

func (h *ArticleHandler) SetFeatured(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/articles/")
	id = strings.TrimSuffix(id, "/featured")
	id = strings.TrimSpace(id)
	if id == "" {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_PARAM", "Article ID or slug is required")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 64*1024)
	var body struct {
		IsFeatured bool `json:"isFeatured"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	art, err := h.service.SetFeatured(r.Context(), id, body.IsFeatured)
	if errors.Is(err, store.ErrNotFound) {
		httputil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Article not found")
		return
	}
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "UPDATE_FAILED", err.Error())
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"article": art,
	})
}

func (h *ArticleHandler) SearchCurated(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page := httputil.ParsePage(r, 30, 100)

	results, total, err := h.service.SearchCuratedArticles(r.Context(), q.Get("category"), q.Get("q"), q.Get("sort"), page.Size, page.Offset)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "SEARCH_FAILED", err.Error())
		return
	}

	httputil.SetPageHeaders(w, total, page)
	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"articles":   results,
		"count":      len(results),
		"total":      total,
		"page":       page.Number,
		"pageSize":   page.Size,
		"totalPages": page.TotalPages(total),
	})
}

// ListArticles returns one page of articles as a plain array. The page position is in the
// X-Total-Count, X-Page, X-Page-Size and X-Total-Pages headers. Ask for ?page=2&pageSize=20.
func (h *ArticleHandler) ListArticles(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	query := r.URL.Query().Get("q")
	if query == "" {
		query = r.URL.Query().Get("search")
	}
	idsParam := r.URL.Query().Get("ids")
	page := httputil.ParsePage(r, 20, 100)

	articles, total, err := h.service.ListArticlesPage(r.Context(), status, query, idsParam, page.Size, page.Offset)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "FETCH_FAILED", "Failed to fetch articles")
		return
	}

	httputil.SetPageHeaders(w, total, page)
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
	art, err := h.service.PublishArticle(r.Context(), id, publishedBy)
	if errors.Is(err, store.ErrNotFound) {
		httputil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Article not found")
		return
	}
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "PUBLISH_FAILED", err.Error())
		return
	}

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

	art, err := h.service.GetArticle(r.Context(), slug)
	if errors.Is(err, store.ErrNotFound) {
		httputil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Article not found")
		return
	}
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "FETCH_FAILED", "Failed to fetch article")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, art)
}

func (h *ArticleHandler) SaveArticle(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 10*1024*1024)
	var art models.Article
	if err := json.NewDecoder(r.Body).Decode(&art); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON request body or payload exceeds 10MB")
		return
	}

	saved, err := h.service.SaveArticle(r.Context(), &art)
	if err != nil {
		if strings.Contains(err.Error(), "cannot be empty") {
			httputil.WriteError(w, http.StatusBadRequest, "VALIDATION_FAILED", err.Error())
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "SAVE_FAILED", "Failed to save article: "+err.Error())
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"article": saved,
	})
}

// PatchArticle changes only the fields sent. The reply lists which fields really changed.
func (h *ArticleHandler) PatchArticle(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_PARAM", "Article ID or slug is required")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 10*1024*1024)
	var patch map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil || len(patch) == 0 {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Send a JSON object with the fields to change")
		return
	}

	saved, changed, err := h.service.PatchArticle(r.Context(), id, patch)
	switch {
	case errors.Is(err, store.ErrNotFound):
		httputil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Article not found")
		return
	case errors.Is(err, services.ErrPatchRejected):
		httputil.WriteError(w, http.StatusBadRequest, "PATCH_REJECTED", err.Error())
		return
	case err != nil && strings.Contains(err.Error(), "cannot be empty"):
		httputil.WriteError(w, http.StatusBadRequest, "VALIDATION_FAILED", err.Error())
		return
	case err != nil:
		httputil.WriteError(w, http.StatusInternalServerError, "SAVE_FAILED", "Failed to save article: "+err.Error())
		return
	}

	if changed == nil {
		changed = []string{}
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"article": saved,
		"changed": changed,
	})
}

func (h *ArticleHandler) ListCandidates(w http.ResponseWriter, r *http.Request) {
	page := httputil.ParsePage(r, 20, 100)
	candidates, total, err := h.service.ListCandidatesPage(r.Context(), page.Size, page.Offset)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "FETCH_FAILED", "Failed to fetch candidates")
		return
	}
	httputil.SetPageHeaders(w, total, page)
	httputil.WriteJSON(w, http.StatusOK, candidates)
}

func (h *ArticleHandler) GetCandidate(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/candidates/")
	id = strings.TrimSpace(id)
	if id == "" {
		httputil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Candidate not found")
		return
	}

	candidate, err := h.service.GetCandidate(r.Context(), id)
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
	sources, err := h.service.ListSources(r.Context())
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

	response := h.service.ResolveLink(targetURL)
	httputil.WriteJSON(w, http.StatusOK, response)
}

func (h *ArticleHandler) CheckSlug(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimSpace(r.URL.Query().Get("slug"))
	if slug == "" {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_PARAM", "Slug parameter is required")
		return
	}
	excludeID := strings.TrimSpace(r.URL.Query().Get("excludeId"))
	available, err := h.service.CheckSlug(r.Context(), slug, excludeID)
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
	tags, err := h.service.ListTags(r.Context(), q)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "FETCH_FAILED", "Failed to fetch tags: "+err.Error())
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"tags": tags,
	})
}

// FormatAIContext renders a structured text digest of an article and its sources for AI ingestion.
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
		case "video":
			sb.WriteString(fmt.Sprintf("[Video: %s]\n\n", block.Src))
		}
	}

	return sb.String()
}
