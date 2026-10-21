package auth

import (
	"encoding/json"
	"net/http"
)

// RoleAdmin is the only role allowed to change anything or read private data.
const RoleAdmin = "admin"

func writeAuthError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// CurrentUser reads the signed-in person from the request without refusing it.
// It returns nil when there is no valid session.
func CurrentUser(r *http.Request, secret string) *User {
	tokenStr := ExtractToken(r)
	if tokenStr == "" {
		return nil
	}
	claims, err := ParseToken(tokenStr, secret)
	if err != nil {
		return nil
	}
	return claims.ToUser()
}

// IsAdmin reports whether the request carries a valid admin session.
// Public handlers use it to decide how much to show, for example whether drafts are visible.
func IsAdmin(r *http.Request, secret string) bool {
	u := CurrentUser(r, secret)
	return u != nil && u.Role == RoleAdmin
}

// RequireAdmin refuses the request unless it carries a valid admin session.
// 401 means "not signed in, or the session ran out". 403 means "signed in, but not an admin".
func RequireAdmin(secret string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenStr := ExtractToken(r)
		if tokenStr == "" {
			writeAuthError(w, http.StatusUnauthorized, "Sign in to continue.")
			return
		}
		claims, err := ParseToken(tokenStr, secret)
		if err != nil {
			writeAuthError(w, http.StatusUnauthorized, "Your session has expired. Sign in again.")
			return
		}
		user := claims.ToUser()
		if user.Role != RoleAdmin {
			writeAuthError(w, http.StatusForbidden, "This account does not have admin access.")
			return
		}
		next(w, WithUser(r, user))
	}
}
