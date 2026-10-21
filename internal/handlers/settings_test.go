package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/okoye-dev/cherryl-media-engine/internal/models"
)

type mockSettingsStore struct {
	settings map[string][]byte
}

func (m *mockSettingsStore) GetSetting(ctx context.Context, key string) ([]byte, error) {
	val, ok := m.settings[key]
	if !ok {
		return nil, nil
	}
	return val, nil
}

func (m *mockSettingsStore) SetSetting(ctx context.Context, key string, val []byte) error {
	m.settings[key] = val
	return nil
}

func (m *mockSettingsStore) GetAllSettings(ctx context.Context) (map[string]any, error) {
	res := make(map[string]any)
	for k, v := range m.settings {
		var parsed any
		if err := json.Unmarshal(v, &parsed); err == nil {
			res[k] = parsed
		}
	}
	return res, nil
}

func put(h *SettingsHandler, body any) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPut, "/api/settings/editorial-copy", bytes.NewReader(b))
	w := httptest.NewRecorder()
	h.SaveEditorialCopy(w, req)
	return w
}

func TestSettings_GetReturnsFullDefaultsWhenNothingSaved(t *testing.T) {
	h := NewSettingsHandler(&mockSettingsStore{settings: map[string][]byte{}})
	w := httptest.NewRecorder()
	h.GetSettings(w, httptest.NewRequest(http.MethodGet, "/api/settings", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var res struct {
		Copy models.EditorialCopy `json:"editorial_copy"`
	}
	if err := json.NewDecoder(w.Body).Decode(&res); err != nil {
		t.Fatal(err)
	}
	if res.Copy.Hero.TitleAccent != "journalism" || res.Copy.Footer.ContactEmail == "" {
		t.Errorf("expected defaults, got %+v", res.Copy)
	}
}

func TestSettings_SaveRoundTripsAndGetKeepsUnsavedDefaults(t *testing.T) {
	store := &mockSettingsStore{settings: map[string][]byte{}}
	h := NewSettingsHandler(store)

	in := models.DefaultEditorialCopy()
	in.Dossier.IssueBadge = "WEEKEND EDITION · ISSUE #19"
	in.Ticker.CustomNotice = ""
	w := put(h, in)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// A fresh read returns what was saved.
	gw := httptest.NewRecorder()
	h.GetSettings(gw, httptest.NewRequest(http.MethodGet, "/api/settings", nil))
	var res struct {
		Copy models.EditorialCopy `json:"editorial_copy"`
	}
	_ = json.NewDecoder(gw.Body).Decode(&res)
	if res.Copy.Dossier.IssueBadge != "WEEKEND EDITION · ISSUE #19" {
		t.Errorf("saved badge not returned: %q", res.Copy.Dossier.IssueBadge)
	}

	// Keys missing from storage fall back to defaults instead of coming back empty.
	store.settings["editorial_copy"] = []byte(`{"footer":{"tagline":"Only this"}}`)
	gw = httptest.NewRecorder()
	h.GetSettings(gw, httptest.NewRequest(http.MethodGet, "/api/settings", nil))
	_ = json.NewDecoder(gw.Body).Decode(&res)
	if res.Copy.Footer.Tagline != "Only this" || res.Copy.Hero.EditionBadge != "DAILY EDITION" {
		t.Errorf("expected merge over defaults, got %+v", res.Copy)
	}
}

func TestSettings_InvalidCopyIsRejectedAndNothingIsSaved(t *testing.T) {
	store := &mockSettingsStore{settings: map[string][]byte{}}
	h := NewSettingsHandler(store)

	in := models.DefaultEditorialCopy()
	in.Footer.ContactEmail = "not-an-email"
	in.Hero.Description = strings.Repeat("x", 500)
	in.Dossier.Heading = "   "
	w := put(h, in)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	var res struct {
		Fields map[string]string `json:"fields"`
	}
	_ = json.NewDecoder(w.Body).Decode(&res)
	for _, k := range []string{"footer.contactEmail", "hero.description", "dossier.heading"} {
		if res.Fields[k] == "" {
			t.Errorf("expected a message for %s, got %v", k, res.Fields)
		}
	}
	if _, saved := store.settings["editorial_copy"]; saved {
		t.Error("invalid copy must not be saved")
	}
}

func TestSettings_LegacyEndpointStillSavesEditorialCopy(t *testing.T) {
	store := &mockSettingsStore{settings: map[string][]byte{}}
	h := NewSettingsHandler(store)
	body, _ := json.Marshal(map[string]any{"editorial_copy": models.DefaultEditorialCopy()})
	w := httptest.NewRecorder()
	h.UpdateSettings(w, httptest.NewRequest(http.MethodPost, "/api/settings", bytes.NewReader(body)))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if _, ok := store.settings["editorial_copy"]; !ok {
		t.Error("expected editorial_copy to be stored")
	}
}
