package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/okoye-dev/cherryl-media-engine/internal/httputil"
	"github.com/okoye-dev/cherryl-media-engine/internal/models"
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

func (h *SettingsHandler) writeSettingsResponse(w http.ResponseWriter, r *http.Request) {
	settings, err := h.service.GetAllSettings(r.Context())
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "SETTINGS_READ_FAILED", "Could not read settings")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, settings)
}

// GetSettings returns every setting. The editorial copy is always complete (defaults fill any gaps).
func (h *SettingsHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	h.writeSettingsResponse(w, r)
}

// SaveEditorialCopy checks and stores the site text. It answers:
//
//	200 {"editorial_copy": {...}}                     saved
//	400 {"error": "...", "fields": {"path": "msg"}}  something is invalid, nothing was saved
//	500                                               the database failed, nothing was saved
func (h *SettingsHandler) SaveEditorialCopy(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 256*1024)
	var in models.EditorialCopy
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "The request body is not valid JSON")
		return
	}
	h.saveCopy(w, r, in)
}

func (h *SettingsHandler) saveCopy(w http.ResponseWriter, r *http.Request, in models.EditorialCopy) {
	saved, fieldErrs, err := h.service.SaveEditorialCopy(r.Context(), in)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "SETTINGS_SAVE_FAILED", "Could not save the site text")
		return
	}
	if len(fieldErrs) > 0 {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"error":  "Some fields need attention",
			"code":   "VALIDATION_FAILED",
			"fields": fieldErrs,
		})
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]any{"editorial_copy": saved})
}

// UpdateSettings keeps the older single endpoint working: {"editorial_copy": {...}} or {"key": "...", "value": ...}.
func (h *SettingsHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 2*1024*1024)
	var payload map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "The request body is not valid JSON, or is over 2MB")
		return
	}

	if raw, ok := payload["editorial_copy"]; ok {
		var in models.EditorialCopy
		if err := json.Unmarshal(raw, &in); err != nil {
			httputil.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "editorial_copy is not valid")
			return
		}
		h.saveCopy(w, r, in)
		return
	}

	var key string
	if raw, ok := payload["key"]; ok {
		_ = json.Unmarshal(raw, &key)
	}
	if strings.TrimSpace(key) == "" {
		httputil.WriteError(w, http.StatusBadRequest, "KEY_REQUIRED", "Send editorial_copy, or a key and value")
		return
	}
	if err := h.service.SetSetting(r.Context(), key, payload["value"]); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "SETTINGS_SAVE_FAILED", err.Error())
		return
	}
	h.writeSettingsResponse(w, r)
}

// PatchEditorialCopy changes only the fields sent, for example {"footer": {"tagline": "..."}}.
// It answers like SaveEditorialCopy: 200 with the whole saved copy, or 400 with the fields that need attention.
func (h *SettingsHandler) PatchEditorialCopy(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 256*1024)
	var patch map[string]any
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil || len(patch) == 0 {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "Send a JSON object with the fields to change")
		return
	}
	saved, fieldErrs, err := h.service.PatchEditorialCopy(r.Context(), patch)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "SETTINGS_SAVE_FAILED", "Could not save the site text")
		return
	}
	if len(fieldErrs) > 0 {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"error":  "Some fields need attention",
			"code":   "VALIDATION_FAILED",
			"fields": fieldErrs,
		})
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]any{"editorial_copy": saved})
}
