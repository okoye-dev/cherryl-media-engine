package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/okoye-dev/cherryl-media-engine/internal/models"
	"github.com/okoye-dev/cherryl-media-engine/internal/store"
	"github.com/okoye-dev/cherryl-media-engine/internal/urlutil"
)

type ArticleHandler struct {
	store store.ArticleStore
}

func NewArticleHandler(s store.ArticleStore) *ArticleHandler {
	return &ArticleHandler{
		store: s,
	}
}

func (h *ArticleHandler) ListArticles(w http.ResponseWriter, r *http.Request) {
	articles, err := h.store.List(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to fetch articles"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(articles)
}

func (h *ArticleHandler) GetArticle(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/api/articles/")
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

	if err := h.store.Save(r.Context(), &art); err != nil {
		http.Error(w, `{"error":"failed to save article: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"article": art,
	})
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
