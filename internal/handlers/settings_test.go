package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
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

func TestSettingsHandler(t *testing.T) {
	store := &mockSettingsStore{settings: make(map[string][]byte)}
	h := NewSettingsHandler(store)

	// 1. GET default settings
	req := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	w := httptest.NewRecorder()
	h.GetSettings(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	var res map[string]any
	if err := json.NewDecoder(w.Body).Decode(&res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if _, ok := res["editorial_copy"]; !ok {
		t.Errorf("expected editorial_copy in settings response")
	}

	// 2. POST updated copy
	newCopy := map[string]any{
		"editorial_copy": map[string]any{
			"hero": map[string]any{
				"editionBadge": "CHERRYL MEDIA · SPECIAL EDITION",
			},
		},
	}
	bodyBytes, _ := json.Marshal(newCopy)
	req2 := httptest.NewRequest(http.MethodPost, "/api/settings", bytes.NewReader(bodyBytes))
	w2 := httptest.NewRecorder()
	h.UpdateSettings(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w2.Code)
	}
}
