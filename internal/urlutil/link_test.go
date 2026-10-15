package urlutil

import "testing"

func TestIsInternalURL(t *testing.T) {
	tests := []struct {
		url      string
		expected bool
	}{
		{"/article/energy-framework", true},
		{"/category/nigerian-current-affairs", true},
		{"http://localhost:3000/article/energy", true},
		{"https://cherrylmedia.com/post", true},
		{"https://nerc.gov.ng/gazette.pdf", false},
		{"https://punchng.com/news/123", false},
		{"https://reuters.com/africa", false},
	}

	for _, tt := range tests {
		got := IsInternalURL(tt.url)
		if got != tt.expected {
			t.Errorf("IsInternalURL(%q) = %v; want %v", tt.url, got, tt.expected)
		}
	}
}
