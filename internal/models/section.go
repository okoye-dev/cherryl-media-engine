package models

import "time"

// CuratedSection maps a site layout section to an ordered list of article IDs and full resolved articles.
type CuratedSection struct {
	SectionKey string         `json:"sectionKey"`
	ArticleIDs []string       `json:"articleIds"`
	Articles   []Article      `json:"articles,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
	UpdatedAt  time.Time      `json:"updatedAt"`
}

// UpdateSectionRequest represents a request to update the articles assigned to a section.
type UpdateSectionRequest struct {
	SectionKey string         `json:"sectionKey"`
	ArticleIDs []string       `json:"articleIds"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}
