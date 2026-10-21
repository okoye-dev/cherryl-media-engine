package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/okoye-dev/cherryl-media-engine/internal/auth"
	"github.com/okoye-dev/cherryl-media-engine/internal/config"
	"github.com/okoye-dev/cherryl-media-engine/internal/store"
)

type AuthHandler struct {
	userStore store.UserStore
	jwtSecret string
	supabase  config.SupabaseConfig
	limiter   *loginLimiter
}

func NewAuthHandler(u store.UserStore, jwtSecret string, supabase config.SupabaseConfig) *AuthHandler {
	return &AuthHandler{
		userStore: u,
		jwtSecret: jwtSecret,
		supabase:  supabase,
		limiter:   newLoginLimiter(8, 15*time.Minute),
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

	r.Body = http.MaxBytesReader(w, r.Body, 64*1024)
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid JSON payload or body too large"}`, http.StatusBadRequest)
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" || req.Password == "" {
		http.Error(w, `{"error":"email and password are required"}`, http.StatusBadRequest)
		return
	}

	emailKey, ipKey := "email:"+email, "ip:"+clientIP(r)
	if h.limiter.blocked(emailKey, ipKey) {
		w.Header().Set("Retry-After", "900")
		http.Error(w, `{"error":"Too many sign-in attempts. Wait 15 minutes and try again."}`, http.StatusTooManyRequests)
		return
	}

	isSecure := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"

	// 1. Authenticate against Supabase Auth if configured
	if h.supabase.URL != "" && (h.supabase.PublishableKey != "" || h.supabase.SecretKey != "") {
		if sbResp, err := h.loginWithSupabase(r.Context(), email, req.Password); err == nil {
			if sbResp.User.Role != auth.RoleAdmin {
				http.Error(w, `{"error":"This account does not have admin access."}`, http.StatusForbidden)
				return
			}
			h.limiter.clear(emailKey)
			http.SetCookie(w, &http.Cookie{
				Name:     "cherryl_token",
				Value:    sbResp.Token,
				Path:     "/",
				Expires:  time.Now().Add(7 * 24 * time.Hour),
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				Secure:   isSecure,
			})

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(sbResp)
			return
		}
	}

	// 2. Fallback to local user store
	user, hash, err := h.userStore.GetUserByEmail(r.Context(), email)
	if err != nil {
		h.limiter.fail(emailKey, ipKey)
		http.Error(w, `{"error":"Wrong email or password."}`, http.StatusUnauthorized)
		return
	}

	if !auth.CheckPassword(hash, req.Password) {
		h.limiter.fail(emailKey, ipKey)
		http.Error(w, `{"error":"Wrong email or password."}`, http.StatusUnauthorized)
		return
	}
	h.limiter.clear(emailKey)

	if !user.Active {
		http.Error(w, `{"error":"user account is deactivated"}`, http.StatusForbidden)
		return
	}

	if user.Role != auth.RoleAdmin {
		http.Error(w, `{"error":"This account does not have admin access."}`, http.StatusForbidden)
		return
	}

	if h.jwtSecret == "" {
		http.Error(w, `{"error":"Sign-in is not set up on the server yet (JWT_SECRET is missing)."}`, http.StatusServiceUnavailable)
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
		Secure:   isSecure,
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
		user = claims.ToUser()
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(user)
}

type supabaseTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	User        struct {
		ID           string                 `json:"id"`
		Email        string                 `json:"email"`
		AppMetadata  map[string]interface{} `json:"app_metadata"`
		UserMetadata map[string]interface{} `json:"user_metadata"`
	} `json:"user"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	Msg              string `json:"msg"`
}

func (h *AuthHandler) loginWithSupabase(ctx context.Context, email, password string) (*LoginResponse, error) {
	if h.supabase.URL == "" {
		return nil, errors.New("supabase not configured")
	}

	apiKey := h.supabase.PublishableKey
	if apiKey == "" {
		apiKey = h.supabase.SecretKey
	}
	if apiKey == "" {
		return nil, errors.New("supabase api key missing")
	}

	endpoint := strings.TrimRight(h.supabase.URL, "/") + "/auth/v1/token?grant_type=password"
	payload, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("apikey", apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tokenResp supabaseTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK || tokenResp.AccessToken == "" {
		errMsg := tokenResp.ErrorDescription
		if errMsg == "" {
			errMsg = tokenResp.Msg
		}
		if errMsg == "" {
			errMsg = tokenResp.Error
		}
		if errMsg == "" {
			errMsg = "supabase authentication failed"
		}
		return nil, errors.New(errMsg)
	}

	fullName := "Newsroom Editor"
	if fn, ok := tokenResp.User.UserMetadata["full_name"].(string); ok && fn != "" {
		fullName = fn
	}
	// Only app_metadata counts: the person can edit user_metadata themselves.
	role := "editor"
	if r, ok := tokenResp.User.AppMetadata["role"].(string); ok && r != "" {
		role = r
	}

	user := &auth.User{
		ID:        tokenResp.User.ID,
		Email:     tokenResp.User.Email,
		FullName:  fullName,
		Role:      role,
		Active:    true,
		CreatedAt: time.Now().UTC(),
	}

	return &LoginResponse{
		Token: tokenResp.AccessToken,
		User:  user,
	}, nil
}
