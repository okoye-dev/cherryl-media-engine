package auth

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// MinPasswordLength is the shortest admin password the server will set up.
const MinPasswordLength = 12

type adminStore interface {
	GetUserByEmail(ctx context.Context, email string) (*User, string, error)
	CreateUser(ctx context.Context, u *User, passwordHash string) error
}

// EnsureAdmin makes sure the admin account named in the environment exists, is active and
// has the password from the environment. The environment is the source of truth: change
// ADMIN_PASSWORD and restart to change the password. It does nothing when either value is empty.
// It returns true when it created or changed the account.
func EnsureAdmin(ctx context.Context, s adminStore, email, password string) (bool, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || password == "" {
		return false, nil
	}
	if len(password) < MinPasswordLength {
		return false, fmt.Errorf("ADMIN_PASSWORD must be at least %d characters", MinPasswordLength)
	}

	existing, hash, err := s.GetUserByEmail(ctx, email)
	if err == nil && existing != nil && existing.Active && existing.Role == RoleAdmin && CheckPassword(hash, password) {
		return false, nil
	}

	newHash, err2 := HashPassword(password)
	if err2 != nil {
		return false, err2
	}
	u := &User{Email: email, FullName: "Newsroom Admin", Role: RoleAdmin, Active: true, CreatedAt: time.Now().UTC()}
	if existing != nil {
		u.ID = existing.ID
		if existing.FullName != "" {
			u.FullName = existing.FullName
		}
		u.AvatarURL = existing.AvatarURL
		u.CreatedAt = existing.CreatedAt
	}
	if err := s.CreateUser(ctx, u, newHash); err != nil {
		return false, err
	}
	return true, nil
}
