package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/okoye-dev/cherryl-media-engine/internal/services"
	"github.com/okoye-dev/cherryl-media-engine/internal/store"
)

type SettingsHandler struct {
	service services.SettingsService
}

func NewSettingsHandler(s store.SettingsStore) *SettingsHandler {
	return &SettingsHandler{service: services.NewSettingsService(s)}
}

func NewSettingsHandlerWithService(svc services.SettingsService) *SettingsHandler {
	return &SettingsHandler{service: svc}
}

// DefaultEditorialCopy returns fallback copy matching website defaults
func DefaultEditorialCopy() map[string]any {
	svc := services.NewSettingsService(nil)
	return svc.DefaultEditorialCopy()
}

func (h *SettingsHandler) writeSettingsResponse(w http.ResponseWriter, r *http.Request) {
	settings, err := h.service.GetAllSettings(r.Context())
	if err != nil {
		settings = map[string]any{
			"editorial_copy": DefaultEditorialCopy(),
		}
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

	r.Body = http.MaxBytesReader(w, r.Body, 2*1024*1024)
	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, `{"error":"invalid JSON body or body exceeds 2MB"}`, http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// If the payload has "key" and "value", store that specific key
	if k, ok := payload["key"].(string); ok && k != "" {
		valBytes, _ := json.Marshal(payload["value"])
		_ = h.service.SetSetting(ctx, k, valBytes)
	} else if edCopy, ok := payload["editorial_copy"].(map[string]any); ok {
		_ = h.service.UpdateEditorialCopy(ctx, edCopy)
	} else {
		// Treat entire payload as editorial_copy if it contains hero/ticker keys
		if _, hasHero := payload["hero"]; hasHero {
			_ = h.service.UpdateEditorialCopy(ctx, payload)
		} else {
			for k, v := range payload {
				if strings.TrimSpace(k) != "" {
					valBytes, _ := json.Marshal(v)
					_ = h.service.SetSetting(ctx, k, valBytes)
				}
			}
		}
	}

	// Return full updated settings
	h.writeSettingsResponse(w, r)
}
