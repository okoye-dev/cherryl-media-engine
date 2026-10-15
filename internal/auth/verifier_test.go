package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestParseECDSAPublicKey(t *testing.T) {
	testKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate test key: %v", err)
	}

	xStr := base64.RawURLEncoding.EncodeToString(testKey.X.Bytes())
	yStr := base64.RawURLEncoding.EncodeToString(testKey.Y.Bytes())

	pub, err := ParseECDSAPublicKey(xStr, yStr)
	if err != nil {
		t.Fatalf("failed to parse EC key: %v", err)
	}

	if pub == nil {
		t.Fatal("expected non-nil public key")
	}

	if !pub.Curve.IsOnCurve(pub.X, pub.Y) {
		t.Fatal("parsed coordinates are not on P-256 curve")
	}
}

func TestDualVerifier(t *testing.T) {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate test EC key: %v", err)
	}

	kid := "test-kid"
	verifier := NewTokenVerifier("test-hmac-secret", "")
	verifier.publicKeys[kid] = &privKey.PublicKey

	// 1. Test ES256 signing and verification
	claims := Claims{
		UserID: "usr-supabase-123",
		Email:  "editor@cherrylmedia.com",
		Role:   "authenticated",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   "usr-supabase-123",
		},
	}

	esToken := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	esToken.Header["kid"] = kid
	signedES, err := esToken.SignedString(privKey)
	if err != nil {
		t.Fatalf("failed to sign ES256 token: %v", err)
	}

	parsedClaims, err := verifier.VerifyToken(signedES)
	if err != nil {
		t.Fatalf("failed to verify ES256 token: %v", err)
	}

	user := parsedClaims.ToUser()
	if user.ID != "usr-supabase-123" {
		t.Errorf("expected user ID usr-supabase-123, got %s", user.ID)
	}
	if user.Email != "editor@cherrylmedia.com" {
		t.Errorf("expected email editor@cherrylmedia.com, got %s", user.Email)
	}
	if user.Role != "editor" { // mapped from 'authenticated'
		t.Errorf("expected default role editor, got %s", user.Role)
	}

	// 2. Test HS256 HMAC verification
	hmacClaims := Claims{
		UserID: "usr-local-456",
		Email:  "local@cherrylmedia.com",
		Role:   "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			Subject:   "usr-local-456",
		},
	}
	hsToken := jwt.NewWithClaims(jwt.SigningMethodHS256, hmacClaims)
	signedHS, err := hsToken.SignedString([]byte("test-hmac-secret"))
	if err != nil {
		t.Fatalf("failed to sign HS256 token: %v", err)
	}

	parsedHS, err := verifier.VerifyToken(signedHS)
	if err != nil {
		t.Fatalf("failed to verify HS256 token: %v", err)
	}
	if parsedHS.UserID != "usr-local-456" {
		t.Errorf("expected usr-local-456, got %s", parsedHS.UserID)
	}
}
