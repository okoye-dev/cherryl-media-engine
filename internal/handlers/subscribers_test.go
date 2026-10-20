package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/okoye-dev/cherryl-media-engine/internal/models"
	"github.com/okoye-dev/cherryl-media-engine/internal/services"
	"github.com/okoye-dev/cherryl-media-engine/internal/store"
)

func TestSubscriberHandler_CRUD(t *testing.T) {
	mem := store.NewMemoryStore()
	h := NewSubscriberHandler(mem)

	// 1. Add subscriber
	payload := map[string]any{
		"email":      "newsroom-reader@example.com",
		"name":       "Adaeze Okon",
		"categories": []string{"state-affairs-environs", "newsbeat"},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/subscribers", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.AddSubscriber(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// 2. List subscribers
	req = httptest.NewRequest(http.MethodGet, "/api/subscribers", nil)
	w = httptest.NewRecorder()
	h.ListSubscribers(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var listResp struct {
		Subscribers []models.PostSubscriber `json:"subscribers"`
		Count       int                     `json:"count"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("failed to decode list response: %v", err)
	}
	if listResp.Count != 1 || listResp.Subscribers[0].Email != "newsroom-reader@example.com" {
		t.Fatalf("unexpected subscriber list: %+v", listResp)
	}

	// 3. Remove subscriber
	req = httptest.NewRequest(http.MethodDelete, "/api/subscribers?email=newsroom-reader@example.com", nil)
	w = httptest.NewRecorder()
	h.RemoveSubscriber(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify removal
	subs, _ := mem.ListSubscribers(context.Background(), false)
	if len(subs) != 0 {
		t.Fatalf("expected 0 subscribers after deletion, got %d", len(subs))
	}
}

func TestArticleHandler_FeaturedAndCuratedSearch(t *testing.T) {
	mem := store.NewMemoryStore()
	res := services.NewMediaResolver(nil)
	svc := services.NewArticleService(mem, res)
	h := NewArticleHandlerWithService(svc)

	ctx := context.Background()
	art := models.Article{
		ID:           "test-art-01",
		Slug:         "lagos-clean-energy-summit",
		Title:        "Lagos State Concludes Clean Energy Summit",
		Deck:         "Key takeaways for sub-national renewable microgrids.",
		Category:     models.PillarCategory("state-affairs-environs"),
		CategoryName: "STATE AFFAIRS & ENVIRONS",
		Status:       models.StatusPublished,
		IsFeatured:   false,
	}
	_ = mem.Save(ctx, &art)

	// 1. Search curated articles
	req := httptest.NewRequest(http.MethodGet, "/api/curation/search?category=state-affairs-environs&q=summit", nil)
	w := httptest.NewRecorder()
	h.SearchCurated(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var searchResp struct {
		Articles []models.Article `json:"articles"`
		Count    int              `json:"count"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &searchResp); err != nil {
		t.Fatalf("failed to unmarshal search response: %v", err)
	}
	if searchResp.Count != 1 || searchResp.Articles[0].Slug != art.Slug {
		t.Fatalf("expected to find article by search, got %+v", searchResp)
	}

	// 2. Set as featured
	featuredPayload := map[string]any{"isFeatured": true}
	fBody, _ := json.Marshal(featuredPayload)
	req = httptest.NewRequest(http.MethodPost, "/api/articles/"+art.ID+"/featured", bytes.NewReader(fBody))
	w = httptest.NewRecorder()
	h.SetFeatured(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for set featured, got %d: %s", w.Code, w.Body.String())
	}

	// 3. Get featured
	req = httptest.NewRequest(http.MethodGet, "/api/articles/featured", nil)
	w = httptest.NewRecorder()
	h.GetFeatured(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for get featured, got %d: %s", w.Code, w.Body.String())
	}

	var featuredArt models.Article
	if err := json.Unmarshal(w.Body.Bytes(), &featuredArt); err != nil {
		t.Fatalf("failed to decode featured article: %v", err)
	}
	if featuredArt.ID != art.ID || !featuredArt.IsFeatured {
		t.Fatalf("expected featured article to match test-art-01 with IsFeatured=true, got %+v", featuredArt)
	}
}
