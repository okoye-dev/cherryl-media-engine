package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/okoye-dev/cherryl-media-engine/internal/auth"
	"github.com/okoye-dev/cherryl-media-engine/internal/store"
)

type AuthHandler struct {
	userStore store.UserStore
	jwtSecret string
}

func NewAuthHandler(u store.UserStore, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		userStore: u,
		jwtSecret: jwtSecret,
	}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string     `json:"token"`
	User  *auth.User `json:"user"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid JSON payload"}`, http.StatusBadRequest)
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" || req.Password == "" {
		http.Error(w, `{"error":"email and password are required"}`, http.StatusBadRequest)
		return
	}

	user, hash, err := h.userStore.GetUserByEmail(r.Context(), email)
	if err != nil {
		http.Error(w, `{"error":"invalid email or password"}`, http.StatusUnauthorized)
		return
	}

	if !auth.CheckPassword(hash, req.Password) {
		http.Error(w, `{"error":"invalid email or password"}`, http.StatusUnauthorized)
		return
	}

	if !user.Active {
		http.Error(w, `{"error":"user account is deactivated"}`, http.StatusForbidden)
		return
	}

	token, err := auth.GenerateToken(user, h.jwtSecret, 7*24*time.Hour)
	if err != nil {
		http.Error(w, `{"error":"failed to generate session token"}`, http.StatusInternalServerError)
		return
	}

	// Set HttpOnly session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "cherryl_token",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false, // set to true in production with HTTPS
	})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(LoginResponse{
		Token: token,
		User:  user,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "cherryl_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"logged_out"}`))
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	tokenStr := auth.ExtractToken(r)
	if tokenStr == "" {
		http.Error(w, `{"error":"unauthenticated"}`, http.StatusUnauthorized)
		return
	}

	claims, err := auth.ParseToken(tokenStr, h.jwtSecret)
	if err != nil {
		http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
		return
	}

	user, err := h.userStore.GetUserByID(r.Context(), claims.UserID)
	if err != nil {
		// Fallback to claims if user record not found in DB
		user = &auth.User{
			ID:       claims.UserID,
			Email:    claims.Email,
			FullName: claims.Name,
			Role:     claims.Role,
			Active:   true,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(user)
}
