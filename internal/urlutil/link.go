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
