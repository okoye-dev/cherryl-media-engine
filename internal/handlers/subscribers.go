package handlers

import (
	"encoding/json"
	"net/http"
	"net/mail"
	"strings"

	"github.com/okoye-dev/cherryl-media-engine/internal/httputil"
	"github.com/okoye-dev/cherryl-media-engine/internal/store"
)

type SubscriberHandler struct {
	store store.SubscriberStore
}

func NewSubscriberHandler(s store.SubscriberStore) *SubscriberHandler {
	return &SubscriberHandler{store: s}
}

func (h *SubscriberHandler) ListSubscribers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	activeOnly := r.URL.Query().Get("activeOnly") == "true"
	page := httputil.ParsePage(r, 25, 100)
	subs, total, err := h.store.ListSubscribersPage(r.Context(), activeOnly, page.Size, page.Offset)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "FETCH_FAILED", "Failed to list subscribers: "+err.Error())
		return
	}

	httputil.SetPageHeaders(w, total, page)
	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"subscribers": subs,
		"count":       len(subs),
		"total":       total,
		"page":        page.Number,
		"pageSize":    page.Size,
		"totalPages":  page.TotalPages(total),
	})
}

func (h *SubscriberHandler) AddSubscriber(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 64*1024)
	var payload struct {
		Email      string   `json:"email"`
		Name       string   `json:"name"`
		Categories []string `json:"categories"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON payload or body too large")
		return
	}

	email := strings.ToLower(strings.TrimSpace(payload.Email))
	if _, err := mail.ParseAddress(email); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_EMAIL", "A valid email address is required")
		return
	}

	sub, err := h.store.AddSubscriber(r.Context(), email, strings.TrimSpace(payload.Name), payload.Categories)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "SAVE_FAILED", "Failed to register subscriber: "+err.Error())
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"success":    true,
		"subscriber": sub,
	})
}

func (h *SubscriberHandler) RemoveSubscriber(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	email := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("email")))
	if email == "" {
		r.Body = http.MaxBytesReader(w, r.Body, 64*1024)
		var payload struct {
			Email string `json:"email"`
		}
		_ = json.NewDecoder(r.Body).Decode(&payload)
		email = strings.ToLower(strings.TrimSpace(payload.Email))
	}

	if email == "" {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_EMAIL", "Email parameter is required")
		return
	}

	if err := h.store.RemoveSubscriber(r.Context(), email); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "DELETE_FAILED", "Failed to remove subscriber: "+err.Error())
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"email":   email,
	})
}

func (h *SubscriberHandler) ListDispatches(w http.ResponseWriter, r *http.Request) {
	articleID := r.URL.Query().Get("articleId")
	if articleID == "" {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_PARAM", "articleId is required")
		return
	}

	dispatches, err := h.store.ListDispatches(r.Context(), articleID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "FETCH_FAILED", err.Error())
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"dispatches": dispatches,
		"count":      len(dispatches),
	})
}
