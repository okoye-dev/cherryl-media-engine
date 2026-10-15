package auth

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type userContextKey struct{}

type Claims struct {
	UserID       string                 `json:"user_id,omitempty"`
	Email        string                 `json:"email,omitempty"`
	Role         string                 `json:"role,omitempty"`
	Name         string                 `json:"name,omitempty"`
	UserMetadata map[string]interface{} `json:"user_metadata,omitempty"`
	AppMetadata  map[string]interface{} `json:"app_metadata,omitempty"`
	jwt.RegisteredClaims
}

func (c *Claims) ToUser() *User {
	id := c.UserID
	if id == "" {
		id = c.Subject
	}

	name := c.Name
	if name == "" && c.UserMetadata != nil {
		if fn, ok := c.UserMetadata["full_name"].(string); ok && fn != "" {
			name = fn
		} else if n, ok := c.UserMetadata["name"].(string); ok && n != "" {
			name = n
		}
	}

	role := c.Role
	if role == "" || role == "authenticated" {
		if c.AppMetadata != nil {
			if r, ok := c.AppMetadata["role"].(string); ok && r != "" {
				role = r
			}
		}
		if role == "" || role == "authenticated" {
			if c.UserMetadata != nil {
				if r, ok := c.UserMetadata["role"].(string); ok && r != "" {
					role = r
				}
			}
		}
		if role == "" || role == "authenticated" {
			role = "editor"
		}
	}

	return &User{
		ID:        id,
		Email:     c.Email,
		FullName:  name,
		Role:      role,
		Active:    true,
		CreatedAt: time.Now().UTC(),
	}
}

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	Role      string    `json:"role"`
	AvatarURL string    `json:"avatar_url,omitempty"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GenerateToken(user *User, secret string, ttl time.Duration) (string, error) {
	claims := Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		Name:   user.FullName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ParseToken(tokenStr, secret string) (*Claims, error) {
	if v := GetDefaultVerifier(); v != nil {
		return v.VerifyToken(tokenStr)
	}

	v := NewTokenVerifier(secret, "")
	return v.VerifyToken(tokenStr)
}

func ExtractToken(r *http.Request) string {
	// 1. Authorization: Bearer <token>
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}
	// 2. Cookie: cherryl_token=<token>
	if cookie, err := r.Cookie("cherryl_token"); err == nil && cookie.Value != "" {
		return cookie.Value
	}
	return ""
}

func WithUser(r *http.Request, u *User) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), userContextKey{}, u))
}

func UserFrom(r *http.Request) (*User, bool) {
	u, ok := r.Context().Value(userContextKey{}).(*User)
	return u, ok
}

// RequireAuth wraps a handler ensuring valid JWT is present
func RequireAuth(secret string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenStr := ExtractToken(r)
		if tokenStr == "" {
			http.Error(w, `{"error":"unauthorized: authentication required"}`, http.StatusUnauthorized)
			return
		}

		claims, err := ParseToken(tokenStr, secret)
		if err != nil {
			http.Error(w, `{"error":"unauthorized: invalid or expired session"}`, http.StatusUnauthorized)
			return
		}

		user := claims.ToUser()
		next(w, WithUser(r, user))
	}
}
