package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/okoye-dev/cherryl-media-engine/internal/store"
)

type SettingsHandler struct {
	store store.SettingsStore
}

func NewSettingsHandler(s store.SettingsStore) *SettingsHandler {
	return &SettingsHandler{store: s}
}

// DefaultEditorialCopy returns fallback copy matching website defaults
func DefaultEditorialCopy() map[string]any {
	return map[string]any{
		"hero": map[string]any{
			"editionBadge": "CHERRYL MEDIA · DAILY EDITION",
			"titleLead":    "Independent ",
			"titleAccent":  "journalism",
			"titleTrail":   " & digital media network.",
			"description":  "Leading digital newsroom and new-age media distribution network covering Nigerian state affairs, business, culture, politics, and dispatches across Africa.",
		},
		"ticker": map[string]any{
			"label":        "LIVE WIRE",
			"customNotice": "BREAKING: Federal Executive Council ratifies state electricity market guidelines and sovereign payment assurance frameworks.",
		},
		"dossier": map[string]any{
			"issueBadge":  "WEEKEND DOSSIER · ISSUE Nº 18",
			"heading":     "Culture, Cinema & Critical Essays.",
			"subheading":  "Investigative photo essays, music production chronicles, and contemporary West African thought.",
		},
		"network": map[string]any{
			"headline":     "Multi-Channel Media Distribution Network",
			"subheadline":  "Reach over 250,000+ readers and professionals daily across West Africa and the diaspora through verified news channels.",
			"subscribers":  "250,000+",
			"reachMonthly": "1.8M+",
		},
		"footer": map[string]any{
			"tagline":         "New-Age Media Distribution & Independent Journalism",
			"locations":       "Abuja · Lagos · London",
			"contactEmail":    "newsroom@cherrylmedia.com",
			"copyrightNotice": "© 2026 Cherryl Media Enterprise. All rights reserved.",
		},
	}
}

func (h *SettingsHandler) writeSettingsResponse(w http.ResponseWriter, r *http.Request) {
	settings, err := h.store.GetAllSettings(r.Context())
	if err != nil || len(settings) == 0 {
		settings = map[string]any{
			"editorial_copy": DefaultEditorialCopy(),
		}
	}

	// Ensure editorial_copy key exists
	if _, ok := settings["editorial_copy"]; !ok {
		settings["editorial_copy"] = DefaultEditorialCopy()
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(settings)
}

func (h *SettingsHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	h.writeSettingsResponse(w, r)
}

func (h *SettingsHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, `{"error":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}

	// If the payload has "key" and "value", store that specific key
	if k, ok := payload["key"].(string); ok && k != "" {
		valBytes, _ := json.Marshal(payload["value"])
		_ = h.store.SetSetting(r.Context(), k, valBytes)
	} else if edCopy, ok := payload["editorial_copy"]; ok {
		valBytes, _ := json.Marshal(edCopy)
		_ = h.store.SetSetting(r.Context(), "editorial_copy", valBytes)
	} else {
		// Treat entire payload as editorial_copy if it contains hero/ticker keys
		if _, hasHero := payload["hero"]; hasHero {
			valBytes, _ := json.Marshal(payload)
			_ = h.store.SetSetting(r.Context(), "editorial_copy", valBytes)
		} else {
			for k, v := range payload {
				if strings.TrimSpace(k) != "" {
					valBytes, _ := json.Marshal(v)
					_ = h.store.SetSetting(r.Context(), k, valBytes)
				}
			}
		}
	}

	// Return full updated settings
	h.writeSettingsResponse(w, r)
}
