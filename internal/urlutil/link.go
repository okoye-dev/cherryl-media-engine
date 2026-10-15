package urlutil

import (
	"net/url"
	"strings"
)

func IsInternalURL(targetURL string) bool {
	trimmed := strings.TrimSpace(targetURL)
	if trimmed == "" {
		return false
	}

	if strings.HasPrefix(trimmed, "/") || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "./") {
		return true
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return !strings.Contains(trimmed, "://")
	}

	host := strings.ToLower(parsed.Hostname())
	return host == "localhost" ||
		host == "127.0.0.1" ||
		strings.Contains(host, "cherryl") ||
		strings.HasSuffix(host, ".local")
}

func CleanArticleSlug(targetURL string) string {
	trimmed := strings.TrimSpace(targetURL)
	if parsed, err := url.Parse(trimmed); err == nil && parsed.Path != "" {
		trimmed = parsed.Path
	}
	trimmed = strings.TrimPrefix(trimmed, "/article/")
	trimmed = strings.TrimPrefix(trimmed, "/")
	return trimmed
}

// ExtractStorageKey extracts the relative storage path (e.g. "media/uuid.webp")
// from full CDN URLs, absolute paths, or local fallback URLs.
func ExtractStorageKey(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if parsed, err := url.Parse(raw); err == nil && parsed.Path != "" {
		raw = parsed.Path
	}
	raw = strings.TrimPrefix(raw, "/")
	raw = strings.TrimPrefix(raw, "uploads/")
	return raw
}

