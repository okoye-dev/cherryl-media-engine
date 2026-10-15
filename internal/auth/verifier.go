package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWK struct {
	Alg    string   `json:"alg"`
	Crv    string   `json:"crv"`
	Ext    bool     `json:"ext"`
	KeyOps []string `json:"key_ops"`
	Kid    string   `json:"kid"`
	Kty    string   `json:"kty"`
	Use    string   `json:"use"`
	X      string   `json:"x"`
	Y      string   `json:"y"`
}

type JWKS struct {
	Keys []JWK `json:"keys"`
}

// TokenVerifier supports dual validation:
// 1. Asymmetric ECDSA ES256 tokens issued by Supabase Auth (via JWKS discovery & preconfigured EC public keys)
// 2. Symmetric HMAC HS256 tokens issued by internal Cherryl Media Engine
type TokenVerifier struct {
	mu         sync.RWMutex
	jwtSecret  string
	jwksURL    string
	publicKeys map[string]*ecdsa.PublicKey // kid -> *ecdsa.PublicKey
	lastFetch  time.Time
}

var (
	defaultVerifierMu sync.RWMutex
	defaultVerifier   *TokenVerifier
)

// InitDefaultVerifier sets up the package-level default verifier
func InitDefaultVerifier(secret, jwksURL string) *TokenVerifier {
	v := NewTokenVerifier(secret, jwksURL)
	defaultVerifierMu.Lock()
	defaultVerifier = v
	defaultVerifierMu.Unlock()
	return v
}

// GetDefaultVerifier returns the package-level default verifier
func GetDefaultVerifier() *TokenVerifier {
	defaultVerifierMu.RLock()
	defer defaultVerifierMu.RUnlock()
	return defaultVerifier
}

func NewTokenVerifier(jwtSecret, jwksURL string) *TokenVerifier {
	return &TokenVerifier{
		jwtSecret:  jwtSecret,
		jwksURL:    jwksURL,
		publicKeys: make(map[string]*ecdsa.PublicKey),
	}
}

// RegisterECKey parses and registers an EC public key by kid
func (v *TokenVerifier) RegisterECKey(kid, xStr, yStr string) error {
	pub, err := ParseECDSAPublicKey(xStr, yStr)
	if err != nil {
		return fmt.Errorf("failed to parse EC key %s: %w", kid, err)
	}

	v.mu.Lock()
	v.publicKeys[kid] = pub
	v.mu.Unlock()
	return nil
}

// ParseECDSAPublicKey converts Base64URL-encoded X and Y coordinate strings into a P-256 ecdsa.PublicKey
func ParseECDSAPublicKey(xStr, yStr string) (*ecdsa.PublicKey, error) {
	xBytes, err := decodeBase64URL(xStr)
	if err != nil {
		return nil, fmt.Errorf("invalid x coordinate: %w", err)
	}
	yBytes, err := decodeBase64URL(yStr)
	if err != nil {
		return nil, fmt.Errorf("invalid y coordinate: %w", err)
	}

	curve := elliptic.P256()
	return &ecdsa.PublicKey{
		Curve: curve,
		X:     new(big.Int).SetBytes(xBytes),
		Y:     new(big.Int).SetBytes(yBytes),
	}, nil
}

func decodeBase64URL(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	// Try RawURLEncoding first (unpadded)
	if b, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(s, "=")); err == nil {
		return b, nil
	}
	// Try URLEncoding with padding
	padding := len(s) % 4
	if padding > 0 {
		s += strings.Repeat("=", 4-padding)
	}
	return base64.URLEncoding.DecodeString(s)
}

// FetchJWKS pulls fresh public keys from the discovery endpoint
func (v *TokenVerifier) FetchJWKS() error {
	if v.jwksURL == "" {
		return errors.New("no JWKS URL configured")
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(v.jwksURL)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("JWKS endpoint returned status %d", resp.StatusCode)
	}

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("failed to decode JWKS: %w", err)
	}

	v.mu.Lock()
	defer v.mu.Unlock()
	for _, k := range jwks.Keys {
		if k.Kty == "EC" && k.Crv == "P-256" && k.X != "" && k.Y != "" {
			if pub, err := ParseECDSAPublicKey(k.X, k.Y); err == nil {
				v.publicKeys[k.Kid] = pub
			}
		}
	}
	v.lastFetch = time.Now()
	return nil
}

// VerifyToken validates an incoming JWT token (supporting both Supabase ES256 and local HS256)
func (v *TokenVerifier) VerifyToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		switch method := token.Method.(type) {
		case *jwt.SigningMethodECDSA:
			// Supabase ES256 Token
			kid, _ := token.Header["kid"].(string)

			v.mu.RLock()
			key, found := v.publicKeys[kid]
			v.mu.RUnlock()

			if found {
				return key, nil
			}

			// If kid not found in memory, try refreshing JWKS from Supabase
			if v.jwksURL != "" && time.Since(v.lastFetch) > 30*time.Second {
				_ = v.FetchJWKS()
				v.mu.RLock()
				key, found = v.publicKeys[kid]
				v.mu.RUnlock()
				if found {
					return key, nil
				}
			}

			// Fallback: If only 1 public key is registered, verify against it
			v.mu.RLock()
			defer v.mu.RUnlock()
			if len(v.publicKeys) == 1 {
				for _, k := range v.publicKeys {
					return k, nil
				}
			}

			return nil, fmt.Errorf("unrecognized EC key kid: %s", kid)

		case *jwt.SigningMethodHMAC:
			// Internal Engine HS256 Token
			if v.jwtSecret == "" {
				return nil, errors.New("HMAC secret not configured")
			}
			return []byte(v.jwtSecret), nil

		default:
			return nil, fmt.Errorf("unsupported token signing algorithm: %v", method.Alg())
		}
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}
