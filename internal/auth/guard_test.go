package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret-for-guard"

func tokenFor(t *testing.T, role string, meta map[string]interface{}) string {
	t.Helper()
	c := Claims{
		UserID:       "u1",
		Email:        "a@b.co",
		Role:         role,
		UserMetadata: meta,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			Subject:   "u1",
		},
	}
	s, err := jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString([]byte(testSecret))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func call(t *testing.T, token string) int {
	t.Helper()
	InitDefaultVerifier(testSecret, "")
	h := RequireAdmin(testSecret, func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	req := httptest.NewRequest(http.MethodPost, "/x", nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	h(w, req)
	return w.Code
}

func TestRequireAdmin(t *testing.T) {
	if got := call(t, ""); got != http.StatusUnauthorized {
		t.Errorf("no token: want 401, got %d", got)
	}
	if got := call(t, "not.a.token"); got != http.StatusUnauthorized {
		t.Errorf("bad token: want 401, got %d", got)
	}
	if got := call(t, tokenFor(t, "editor", nil)); got != http.StatusForbidden {
		t.Errorf("editor: want 403, got %d", got)
	}
	if got := call(t, tokenFor(t, "admin", nil)); got != http.StatusOK {
		t.Errorf("admin: want 200, got %d", got)
	}
}

func TestRoleCannotBeSetThroughUserMetadata(t *testing.T) {
	// A Supabase user can edit their own user_metadata, so it must never grant admin.
	tok := tokenFor(t, "authenticated", map[string]interface{}{"role": "admin"})
	if got := call(t, tok); got != http.StatusForbidden {
		t.Errorf("user_metadata role: want 403, got %d", got)
	}
}

func TestExpiredTokenIsRefused(t *testing.T) {
	InitDefaultVerifier(testSecret, "")
	c := Claims{Role: "admin", RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute))}}
	s, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString([]byte(testSecret))
	if got := call(t, s); got != http.StatusUnauthorized {
		t.Errorf("expired: want 401, got %d", got)
	}
}

type fakeUsers struct {
	u    *User
	hash string
}

func (f *fakeUsers) GetUserByEmail(_ context.Context, email string) (*User, string, error) {
	if f.u == nil {
		return nil, "", context.Canceled
	}
	return f.u, f.hash, nil
}
func (f *fakeUsers) CreateUser(_ context.Context, u *User, h string) error {
	f.u, f.hash = u, h
	return nil
}

func TestEnsureAdmin(t *testing.T) {
	f := &fakeUsers{}
	if changed, err := EnsureAdmin(context.Background(), f, "", ""); changed || err != nil {
		t.Fatalf("empty env should do nothing, got %v %v", changed, err)
	}
	if _, err := EnsureAdmin(context.Background(), f, "a@b.co", "short"); err == nil {
		t.Fatal("a short password must be refused")
	}
	changed, err := EnsureAdmin(context.Background(), f, " Admin@B.co ", "a-long-enough-password")
	if err != nil || !changed || f.u.Role != RoleAdmin || f.u.Email != "admin@b.co" || !CheckPassword(f.hash, "a-long-enough-password") {
		t.Fatalf("admin not created properly: %v %v %+v", changed, err, f.u)
	}
	if changed, _ := EnsureAdmin(context.Background(), f, "admin@b.co", "a-long-enough-password"); changed {
		t.Error("second run with the same values should change nothing")
	}
	if changed, _ := EnsureAdmin(context.Background(), f, "admin@b.co", "another-long-password"); !changed || !CheckPassword(f.hash, "another-long-password") {
		t.Error("a new password in the environment should replace the old one")
	}
}
