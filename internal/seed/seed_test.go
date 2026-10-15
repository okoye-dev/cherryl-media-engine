package seed

import (
	"testing"
)

func TestGetAllWebsiteArticles(t *testing.T) {
	arts := GetAllWebsiteArticles()
	if len(arts) != 22 {
		t.Fatalf("expected 22 articles, got %d", len(arts))
	}

	seenIDs := make(map[string]string)
	seenSlugs := make(map[string]string)

	for _, a := range arts {
		if a.ID == "" {
			t.Errorf("article %s has empty ID", a.Slug)
		}
		if a.Slug == "" {
			t.Errorf("article has empty slug")
		}
		if a.Title == "" {
			t.Errorf("article %s has empty title", a.Slug)
		}
		if a.Category == "" {
			t.Errorf("article %s has empty category", a.Slug)
		}
		if prev, exists := seenIDs[a.ID]; exists {
			t.Errorf("duplicate ID %s for slugs %s and %s", a.ID, prev, a.Slug)
		}
		seenIDs[a.ID] = a.Slug

		if prev, exists := seenSlugs[a.Slug]; exists {
			t.Errorf("duplicate slug %s", prev)
		}
		seenSlugs[a.Slug] = a.ID

		t.Logf("slug: %-60s -> id: %s", a.Slug, a.ID)
	}
}
