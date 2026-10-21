package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/okoye-dev/cherryl-media-engine/internal/models"
	"github.com/okoye-dev/cherryl-media-engine/internal/services"
	"github.com/okoye-dev/cherryl-media-engine/internal/storage"
	"github.com/okoye-dev/cherryl-media-engine/internal/store"
)

type SectionsHandler struct {
	service services.SectionService
	isAdmin func(*http.Request) bool
}

// WithAdminCheck sets how the handler recognises an admin. Everyone else only sees live stories in sections.
func (h *SectionsHandler) WithAdminCheck(f func(*http.Request) bool) *SectionsHandler {
	h.isAdmin = f
	return h
}

// hideUnpublished removes draft and scheduled stories from a section for the public.
func (h *SectionsHandler) hideUnpublished(r *http.Request, secs ...*models.CuratedSection) {
	if h.isAdmin != nil && h.isAdmin(r) {
		return
	}
	now := time.Now()
	for _, sec := range secs {
		live := make([]models.Article, 0, len(sec.Articles))
		for _, a := range sec.Articles {
			if a.Status == models.StatusPublished && !a.PublishedAt.After(now) {
				live = append(live, a)
			}
		}
		sec.Articles = live
	}
}

func NewSectionsHandler(secStore store.SectionStore, artStore store.ArticleStore, st storage.Storage) *SectionsHandler {
	resolver := services.NewMediaResolver(st)
	svc := services.NewSectionService(secStore, artStore, resolver)
	return &SectionsHandler{service: svc}
}

func NewSectionsHandlerWithService(svc services.SectionService) *SectionsHandler {
	return &SectionsHandler{service: svc}
}

// GetSections handles GET /api/sections
func (h *SectionsHandler) GetSections(w http.ResponseWriter, r *http.Request) {
	sections, byKey, err := h.service.GetSections(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to fetch curated sections"}`, http.StatusInternalServerError)
		return
	}

	for i := range sections {
		h.hideUnpublished(r, &sections[i])
	}
	for k, v := range byKey {
		h.hideUnpublished(r, &v)
		byKey[k] = v
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"sections": sections,
		"byKey":    byKey,
	})
}

// GetSectionByKey handles GET /api/sections/{key}
func (h *SectionsHandler) GetSectionByKey(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/api/sections/")
	key = strings.TrimSpace(key)
	if key == "" {
		h.GetSections(w, r)
		return
	}

	sec, err := h.service.GetSectionByKey(r.Context(), key)
	if errors.Is(err, store.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, `{"error":"failed to fetch section"}`, http.StatusInternalServerError)
		return
	}

	h.hideUnpublished(r, sec)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(sec)
}

// UpdateSections handles POST /api/sections and PUT /api/sections
// Supports single section update or batch { sections: [...] }
func (h *SectionsHandler) UpdateSections(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 5*1024*1024)
	var bodyRaw json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&bodyRaw); err != nil {
		http.Error(w, `{"error":"invalid JSON body or body exceeds 5MB"}`, http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Try batch first: { "sections": [...] }
	var batch struct {
		Sections []models.UpdateSectionRequest `json:"sections"`
	}
	if err := json.Unmarshal(bodyRaw, &batch); err == nil && len(batch.Sections) > 0 {
		if err := h.service.UpdateSections(ctx, batch.Sections); err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"count":   len(batch.Sections),
		})
		return
	}

	// Try direct array: [ { "sectionKey": ... } ]
	var arrayBatch []models.UpdateSectionRequest
	if err := json.Unmarshal(bodyRaw, &arrayBatch); err == nil && len(arrayBatch) > 0 {
		if err := h.service.UpdateSections(ctx, arrayBatch); err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"count":   len(arrayBatch),
		})
		return
	}

	// Fall back to single section object
	var single models.UpdateSectionRequest
	if err := json.Unmarshal(bodyRaw, &single); err != nil || single.SectionKey == "" {
		http.Error(w, `{"error":"invalid section update payload; sectionKey is required"}`, http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateSections(ctx, []models.UpdateSectionRequest{single}); err != nil {
		http.Error(w, `{"error":"failed to update section"}`, http.StatusInternalServerError)
		return
	}

	sec, _ := h.service.GetSectionByKey(ctx, single.SectionKey)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"section": sec,
	})
}
