package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/okoye-dev/cherryl-media-engine/internal/models"
	"github.com/okoye-dev/cherryl-media-engine/internal/store"
)

func memWithArticles(t *testing.T, n int) (*store.MemoryStore, *ArticleHandler) {
	t.Helper()
	mem := store.NewMemoryStore()
	for i := 1; i <= n; i++ {
		a := models.Article{
			Slug:        fmt.Sprintf("story-%02d", i),
			Title:       fmt.Sprintf("Story %02d", i),
			Deck:        "A deck",
			Status:      models.StatusPublished,
			PublishedAt: time.Now().Add(-time.Duration(i) * time.Hour),
			CreatedAt:   time.Now(),
			Author:      models.Author{Name: "Ada", Role: "Reporter"},
		}
		if err := mem.Save(context.Background(), &a); err != nil {
			t.Fatal(err)
		}
	}
	return mem, NewArticleHandler(mem, nil)
}

func TestListArticles_IsPagedAndReportsTotals(t *testing.T) {
	mem, h := memWithArticles(t, 45)
	existing, _ := mem.List(context.Background(), "published")
	total := len(existing) // the store starts with a few stories of its own
	pages := (total + 19) / 20
	lastPage := total - (pages-1)*20

	get := func(query string) (*httptest.ResponseRecorder, []models.Article) {
		w := httptest.NewRecorder()
		h.ListArticles(w, httptest.NewRequest(http.MethodGet, "/api/articles"+query, nil))
		var list []models.Article
		_ = json.NewDecoder(w.Body).Decode(&list)
		return w, list
	}

	w, page1 := get("?status=published&pageSize=20")
	if len(page1) != 20 || w.Header().Get("X-Total-Count") != strconv.Itoa(total) || w.Header().Get("X-Total-Pages") != strconv.Itoa(pages) || w.Header().Get("X-Page") != "1" {
		t.Fatalf("page 1 wrong: len=%d headers=%v", len(page1), w.Header())
	}
	_, page3 := get("?status=published&pageSize=20&page=" + strconv.Itoa(pages))
	if len(page3) != lastPage {
		t.Fatalf("last page should hold the remaining %d, got %d", lastPage, len(page3))
	}
	if page1[0].Slug == page3[0].Slug {
		t.Error("different pages must return different stories")
	}
	_, beyond := get("?status=published&pageSize=20&page=9")
	if len(beyond) != 0 {
		t.Errorf("a page past the end is empty, got %d", len(beyond))
	}
	// Nobody can ask for everything at once.
	w, big := get("?status=published&pageSize=5000")
	if len(big) != total || w.Header().Get("X-Page-Size") != "100" {
		t.Errorf("page size should be capped at 100, got size header %q and %d items", w.Header().Get("X-Page-Size"), len(big))
	}
	// The older limit parameter still works as the page size.
	_, limited := get("?status=published&limit=7")
	if len(limited) != 7 {
		t.Errorf("limit=7 should return 7, got %d", len(limited))
	}
}

func patch(h *ArticleHandler, slug string, body any) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPatch, "/api/articles/"+slug, bytes.NewReader(b))
	req.SetPathValue("id", slug)
	w := httptest.NewRecorder()
	h.PatchArticle(w, req)
	return w
}

func TestPatchArticle_ChangesOnlyTheFieldsSent(t *testing.T) {
	mem, h := memWithArticles(t, 3)

	w := patch(h, "story-02", map[string]any{"title": "A new headline"})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var res struct {
		Changed []string       `json:"changed"`
		Article models.Article `json:"article"`
	}
	_ = json.NewDecoder(w.Body).Decode(&res)
	if len(res.Changed) != 1 || res.Changed[0] != "title" {
		t.Errorf("only title should be reported as changed, got %v", res.Changed)
	}

	stored, _ := mem.GetBySlug(context.Background(), "story-02")
	if stored.Title != "A new headline" {
		t.Errorf("title was not saved: %q", stored.Title)
	}
	if stored.Deck != "A deck" || stored.Author.Name != "Ada" || stored.Status != models.StatusPublished {
		t.Errorf("fields that were not sent must stay as they were: %+v", stored)
	}

	// Sending a value that is already stored is not a change.
	w = patch(h, "story-02", map[string]any{"title": "A new headline"})
	_ = json.NewDecoder(w.Body).Decode(&res)
	if w.Code != http.StatusOK || len(res.Changed) != 0 {
		t.Errorf("an identical value changes nothing, got %v", res.Changed)
	}
}

func TestPatchArticle_Refusals(t *testing.T) {
	mem, h := memWithArticles(t, 2)
	ctx := context.Background()

	if w := patch(h, "no-such-story", map[string]any{"title": "x"}); w.Code != http.StatusNotFound {
		t.Errorf("unknown story should be 404, got %d", w.Code)
	}

	// A draft cannot be taken live through a patch: publishing has its own checks.
	d := models.Article{Slug: "a-draft", Title: "A draft", Status: models.StatusDraft, CreatedAt: time.Now(), PublishedAt: time.Now()}
	_ = mem.Save(ctx, &d)
	if w := patch(h, "a-draft", map[string]any{"status": "published"}); w.Code != http.StatusBadRequest {
		t.Errorf("publishing by patch should be refused with 400, got %d", w.Code)
	}
	stored, _ := mem.GetBySlug(ctx, "a-draft")
	if stored.Status != models.StatusDraft {
		t.Error("the draft must stay a draft")
	}

	// Empty titles are refused.
	if w := patch(h, "story-01", map[string]any{"title": ""}); w.Code != http.StatusBadRequest {
		t.Errorf("an empty title should be refused with 400, got %d", w.Code)
	}
}

func TestPatchArticle_RenamesTheSlugWithoutMakingACopy(t *testing.T) {
	mem, h := memWithArticles(t, 2)
	ctx := context.Background()

	before, _ := mem.List(ctx, "all")
	if w := patch(h, "story-01", map[string]any{"slug": "renamed-story"}); w.Code != http.StatusOK {
		t.Fatalf("rename failed: %d %s", w.Code, w.Body.String())
	}
	after, _ := mem.List(ctx, "all")
	if len(after) != len(before) {
		t.Errorf("renaming must not create a copy: %d before, %d after", len(before), len(after))
	}
	if _, err := mem.GetBySlug(ctx, "renamed-story"); err != nil {
		t.Error("the story should now answer to its new address")
	}
	// Taking another story's address is refused.
	if w := patch(h, "renamed-story", map[string]any{"slug": "story-02"}); w.Code != http.StatusBadRequest {
		t.Errorf("a taken address should be refused with 400, got %d", w.Code)
	}
}

func TestPatchEditorialCopy_ChangesOnlyTheFieldsSent(t *testing.T) {
	store := &mockSettingsStore{settings: map[string][]byte{}}
	h := NewSettingsHandler(store)

	// Start from a saved copy with a custom footer tagline.
	in := models.DefaultEditorialCopy()
	in.Footer.Tagline = "Custom tagline"
	if w := put(h, in); w.Code != http.StatusOK {
		t.Fatalf("setup save failed: %d", w.Code)
	}

	body, _ := json.Marshal(map[string]any{"dossier": map[string]any{"issueBadge": "WEEKEND EDITION · ISSUE #20"}})
	req := httptest.NewRequest(http.MethodPatch, "/api/settings/editorial-copy", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.PatchEditorialCopy(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var res struct {
		Copy models.EditorialCopy `json:"editorial_copy"`
	}
	_ = json.NewDecoder(w.Body).Decode(&res)
	if res.Copy.Dossier.IssueBadge != "WEEKEND EDITION · ISSUE #20" {
		t.Errorf("patched field not saved: %q", res.Copy.Dossier.IssueBadge)
	}
	if res.Copy.Footer.Tagline != "Custom tagline" || res.Copy.Hero.EditionBadge != "DAILY EDITION" {
		t.Errorf("fields that were not sent must stay: %+v", res.Copy)
	}

	// A patch that breaks a rule is refused and nothing changes.
	bad, _ := json.Marshal(map[string]any{"footer": map[string]any{"contactEmail": "nope"}})
	w = httptest.NewRecorder()
	h.PatchEditorialCopy(w, httptest.NewRequest(http.MethodPatch, "/api/settings/editorial-copy", bytes.NewReader(bad)))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	var stored models.EditorialCopy
	_ = json.Unmarshal(store.settings["editorial_copy"], &stored)
	if stored.Footer.ContactEmail != models.DefaultEditorialCopy().Footer.ContactEmail {
		t.Errorf("a refused patch must not change the stored copy, got %q", stored.Footer.ContactEmail)
	}
}
