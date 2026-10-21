package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/okoye-dev/cherryl-media-engine/internal/models"
	"github.com/okoye-dev/cherryl-media-engine/internal/store"
)

func withDraft(t *testing.T, asAdmin bool) (*ArticleHandler, string) {
	t.Helper()
	mem := store.NewMemoryStore()
	save := func(slug string, status models.PublishingStatus, at time.Time) {
		a := models.Article{Slug: slug, Title: slug, Status: status, PublishedAt: at, CreatedAt: time.Now()}
		if err := mem.Save(context.Background(), &a); err != nil {
			t.Fatal(err)
		}
	}
	save("live-story", models.StatusPublished, time.Now().Add(-time.Hour))
	save("secret-draft", models.StatusDraft, time.Now())
	h := NewArticleHandler(mem, nil).WithAdminCheck(func(*http.Request) bool { return asAdmin })
	return h, "secret-draft"
}

func TestPublicNeverSeesDrafts(t *testing.T) {
	h, _ := withDraft(t, false)

	// Asking for drafts, or everything, still returns live stories only.
	for _, status := range []string{"draft", "all", ""} {
		w := httptest.NewRecorder()
		h.ListArticles(w, httptest.NewRequest(http.MethodGet, "/api/articles?status="+status, nil))
		if strings.Contains(w.Body.String(), "secret-draft") {
			t.Errorf("status=%q leaked a draft: %s", status, w.Body.String())
		}
	}
	// A search must not find a draft either.
	w := httptest.NewRecorder()
	h.ListArticles(w, httptest.NewRequest(http.MethodGet, "/api/articles?q=secret", nil))
	if strings.Contains(w.Body.String(), "secret-draft") {
		t.Errorf("search leaked a draft: %s", w.Body.String())
	}
	// Opening a draft directly is a plain "not found".
	w = httptest.NewRecorder()
	h.GetArticle(w, httptest.NewRequest(http.MethodGet, "/api/articles/secret-draft", nil))
	if w.Code != http.StatusNotFound {
		t.Errorf("draft by slug: want 404, got %d", w.Code)
	}
	w = httptest.NewRecorder()
	h.GetArticle(w, httptest.NewRequest(http.MethodGet, "/api/articles/live-story", nil))
	if w.Code != http.StatusOK {
		t.Errorf("live story by slug: want 200, got %d", w.Code)
	}
}

func TestAdminSeesDrafts(t *testing.T) {
	h, _ := withDraft(t, true)
	w := httptest.NewRecorder()
	h.ListArticles(w, httptest.NewRequest(http.MethodGet, "/api/articles?status=draft", nil))
	if !strings.Contains(w.Body.String(), "secret-draft") {
		t.Errorf("admin should see drafts: %s", w.Body.String())
	}
	w = httptest.NewRecorder()
	h.GetArticle(w, httptest.NewRequest(http.MethodGet, "/api/articles/secret-draft", nil))
	if w.Code != http.StatusOK {
		t.Errorf("admin draft by slug: want 200, got %d", w.Code)
	}
}

func TestLoginLimiter(t *testing.T) {
	l := newLoginLimiter(3, time.Minute)
	for i := 0; i < 3; i++ {
		if l.blocked("email:a") {
			t.Fatal("blocked too early")
		}
		l.fail("email:a", "ip:1")
	}
	if !l.blocked("email:a") || !l.blocked("ip:1") {
		t.Error("should be blocked after 3 failures")
	}
	if l.blocked("email:other") {
		t.Error("another email must not be blocked")
	}
	l.clear("email:a")
	if l.blocked("email:a") {
		t.Error("clearing should unblock")
	}
}
