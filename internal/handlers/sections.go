package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/okoye-dev/cherryl-media-engine/internal/models"
	"github.com/okoye-dev/cherryl-media-engine/internal/storage"
	"github.com/okoye-dev/cherryl-media-engine/internal/store"
	"github.com/okoye-dev/cherryl-media-engine/internal/urlutil"
)

type SectionsHandler struct {
	sectionStore store.SectionStore
	articleStore store.ArticleStore
	storage      storage.Storage
}

func NewSectionsHandler(secStore store.SectionStore, artStore store.ArticleStore, st storage.Storage) *SectionsHandler {
	return &SectionsHandler{
		sectionStore: secStore,
		articleStore: artStore,
		storage:      st,
	}
}

// resolveArticlesMedia resolves storage keys to full CDN URLs for each article
func (h *SectionsHandler) resolveArticlesMedia(articles []models.Article) {
	if h.storage == nil {
		return
	}
	for i := range articles {
		art := &articles[i]
		if art.FeaturedImage.StorageKey != "" {
			art.FeaturedImage.URL = h.storage.GetPublicURL(art.FeaturedImage.StorageKey)
		} else if key := urlutil.ExtractStorageKey(art.FeaturedImage.URL); strings.HasPrefix(key, "media/") {
			art.FeaturedImage.StorageKey = key
			art.FeaturedImage.URL = h.storage.GetPublicURL(key)
		}
		for j := range art.Body {
			b := &art.Body[j]
			if (b.Type == "image" || b.Type == "video") && b.StorageKey != "" {
				b.Src = h.storage.GetPublicURL(b.StorageKey)
			} else if (b.Type == "image" || b.Type == "video") && strings.HasPrefix(urlutil.ExtractStorageKey(b.Src), "media/") {
				b.StorageKey = urlutil.ExtractStorageKey(b.Src)
				b.Src = h.storage.GetPublicURL(b.StorageKey)
			}
		}
	}
}

// populateSection resolves the articles for the section from articleStore
func (h *SectionsHandler) populateSection(r *http.Request, sec *models.CuratedSection) error {
	if len(sec.ArticleIDs) == 0 {
		sec.Articles = []models.Article{}
		return nil
	}
	arts, err := h.articleStore.GetByIDs(r.Context(), sec.ArticleIDs)
	if err != nil {
		return err
	}
	h.resolveArticlesMedia(arts)
	sec.Articles = arts
	return nil
}

// GetSections handles GET /api/sections
func (h *SectionsHandler) GetSections(w http.ResponseWriter, r *http.Request) {
	sections, err := h.sectionStore.GetAllSections(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to fetch curated sections"}`, http.StatusInternalServerError)
		return
	}

	byKey := make(map[string]models.CuratedSection, len(sections))
	for i := range sections {
		_ = h.populateSection(r, &sections[i])
		byKey[sections[i].SectionKey] = sections[i]
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

	sec, err := h.sectionStore.GetSection(r.Context(), key)
	if errors.Is(err, store.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, `{"error":"failed to fetch section"}`, http.StatusInternalServerError)
		return
	}

	_ = h.populateSection(r, sec)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(sec)
}

// UpdateSections handles POST /api/sections and PUT /api/sections
// Supports single section update or batch { sections: [...] }
func (h *SectionsHandler) UpdateSections(w http.ResponseWriter, r *http.Request) {
	var bodyRaw json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&bodyRaw); err != nil {
		http.Error(w, `{"error":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}

	// Try batch first: { "sections": [...] }
	var batch struct {
		Sections []models.UpdateSectionRequest `json:"sections"`
	}
	if err := json.Unmarshal(bodyRaw, &batch); err == nil && len(batch.Sections) > 0 {
		for _, s := range batch.Sections {
			if s.SectionKey == "" {
				continue
			}
			if err := h.sectionStore.SaveSection(r.Context(), s.SectionKey, s.ArticleIDs, s.Metadata); err != nil {
				http.Error(w, `{"error":"failed to save section: `+s.SectionKey+`"}`, http.StatusInternalServerError)
				return
			}
		}
		h.GetSections(w, r)
		return
	}

	// Single section update: { "sectionKey": "...", "articleIds": [...] }
	var single models.UpdateSectionRequest
	if err := json.Unmarshal(bodyRaw, &single); err != nil || single.SectionKey == "" {
		http.Error(w, `{"error":"sectionKey and articleIds are required"}`, http.StatusBadRequest)
		return
	}

	if err := h.sectionStore.SaveSection(r.Context(), single.SectionKey, single.ArticleIDs, single.Metadata); err != nil {
		http.Error(w, `{"error":"failed to save section"}`, http.StatusInternalServerError)
		return
	}

	sec, err := h.sectionStore.GetSection(r.Context(), single.SectionKey)
	if err == nil {
		_ = h.populateSection(r, sec)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"section": sec,
	})
}

// UpdateSectionByKey handles PUT /api/sections/{key}
func (h *SectionsHandler) UpdateSectionByKey(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/api/sections/")
	key = strings.TrimSpace(key)
	if key == "" {
		http.Error(w, `{"error":"section key in path is required"}`, http.StatusBadRequest)
		return
	}

	var req models.UpdateSectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}
	req.SectionKey = key

	if err := h.sectionStore.SaveSection(r.Context(), req.SectionKey, req.ArticleIDs, req.Metadata); err != nil {
		http.Error(w, `{"error":"failed to save section"}`, http.StatusInternalServerError)
		return
	}

	sec, err := h.sectionStore.GetSection(r.Context(), key)
	if err == nil {
		_ = h.populateSection(r, sec)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"section": sec,
	})
}
