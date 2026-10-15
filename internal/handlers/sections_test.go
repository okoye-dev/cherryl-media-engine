package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/okoye-dev/cherryl-media-engine/internal/models"
	"github.com/okoye-dev/cherryl-media-engine/internal/seed"
	"github.com/okoye-dev/cherryl-media-engine/internal/store"
)

func setupTestStore(t *testing.T) (*store.MemoryStore, *ArticleHandler, *SectionsHandler) {
	mem := store.NewMemoryStore()
	ctx := context.Background()

	// Seed articles into memory store
	for _, art := range seed.GetAllWebsiteArticles() {
		a := art
		if err := mem.Save(ctx, &a); err != nil {
			t.Fatalf("failed to save seed article: %v", err)
		}
	}

	artHandler := NewArticleHandler(mem, nil)
	secHandler := NewSectionsHandler(mem, mem, nil)
	return mem, artHandler, secHandler
}

func TestSectionsHandler_GetSections(t *testing.T) {
	_, _, secHandler := setupTestStore(t)

	req := httptest.NewRequest(http.MethodGet, "/api/sections", nil)
	w := httptest.NewRecorder()

	secHandler.GetSections(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Sections []models.CuratedSection          `json:"sections"`
		ByKey    map[string]models.CuratedSection `json:"byKey"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Sections) < 7 {
		t.Errorf("expected at least 7 sections, got %d", len(resp.Sections))
	}

	heroLead, ok := resp.ByKey["hero_lead"]
	if !ok {
		t.Fatalf("expected hero_lead section in byKey")
	}

	if len(heroLead.Articles) != 1 {
		t.Fatalf("expected 1 populated article in hero_lead, got %d", len(heroLead.Articles))
	}

	if heroLead.Articles[0].Slug != "federal-executive-council-approves-unified-energy-framework" {
		t.Errorf("unexpected hero lead article: %s", heroLead.Articles[0].Slug)
	}
}

func TestSectionsHandler_UpdateSections(t *testing.T) {
	_, _, secHandler := setupTestStore(t)

	// Update hero_lead to another article
	newArticleID := "f3033f02-9628-52d2-bef4-8b0814c6d835" // Nollywood Studio Architecture
	payload := models.UpdateSectionRequest{
		SectionKey: "hero_lead",
		ArticleIDs: []string{newArticleID},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/sections", bytes.NewReader(body))
	w := httptest.NewRecorder()

	secHandler.UpdateSections(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Success bool                  `json:"success"`
		Section models.CuratedSection `json:"section"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal update response: %v", err)
	}

	if !resp.Success {
		t.Errorf("expected success true")
	}
	if len(resp.Section.Articles) != 1 || resp.Section.Articles[0].ID != newArticleID {
		t.Errorf("expected updated hero_lead to have article %s, got: %+v", newArticleID, resp.Section.Articles)
	}
}

func TestArticleHandler_SearchAndGetByIDs(t *testing.T) {
	_, artHandler, _ := setupTestStore(t)

	// Test Search
	req := httptest.NewRequest(http.MethodGet, "/api/articles?q=energy", nil)
	w := httptest.NewRecorder()
	artHandler.ListArticles(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for search, got %d", w.Code)
	}

	var searchResults []models.Article
	if err := json.Unmarshal(w.Body.Bytes(), &searchResults); err != nil {
		t.Fatalf("failed to unmarshal search results: %v", err)
	}
	if len(searchResults) == 0 {
		t.Fatalf("expected at least 1 search result for 'energy'")
	}

	// Test GetByIDs
	id1 := "2f405add-8bba-5261-99af-15feaa62754b"
	id2 := "f3033f02-9628-52d2-bef4-8b0814c6d835"
	reqIDs := httptest.NewRequest(http.MethodGet, "/api/articles?ids="+id1+","+id2, nil)
	wIDs := httptest.NewRecorder()
	artHandler.ListArticles(wIDs, reqIDs)

	var byIDs []models.Article
	if err := json.Unmarshal(wIDs.Body.Bytes(), &byIDs); err != nil {
		t.Fatalf("failed to unmarshal ids result: %v", err)
	}
	if len(byIDs) != 2 {
		t.Fatalf("expected 2 articles, got %d", len(byIDs))
	}
	if byIDs[0].ID != id1 || byIDs[1].ID != id2 {
		t.Errorf("expected articles ordered [%s, %s], got [%s, %s]", id1, id2, byIDs[0].ID, byIDs[1].ID)
	}
}
