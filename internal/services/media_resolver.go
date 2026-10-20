package services

import (
	"strings"

	"github.com/okoye-dev/cherryl-media-engine/internal/models"
	"github.com/okoye-dev/cherryl-media-engine/internal/storage"
	"github.com/okoye-dev/cherryl-media-engine/internal/urlutil"
)

// MediaResolver defines methods for hydrating media URLs from storage keys and vice versa.
type MediaResolver interface {
	ResolveMedia(art *models.Article)
	ResolveArticles(articles []models.Article)
	NormalizeMediaKeys(art *models.Article)
}

type mediaResolver struct {
	storage storage.Storage
}

// NewMediaResolver constructs a new MediaResolver instance.
func NewMediaResolver(st storage.Storage) MediaResolver {
	return &mediaResolver{storage: st}
}

// ResolveMedia dynamically builds public URLs using the active storage configuration
// so the database only needs to store location-independent storage keys.
func (m *mediaResolver) ResolveMedia(art *models.Article) {
	if art == nil || m.storage == nil {
		return
	}

	if art.FeaturedImage.StorageKey != "" {
		art.FeaturedImage.URL = m.storage.GetPublicURL(art.FeaturedImage.StorageKey)
	} else if key := urlutil.ExtractStorageKey(art.FeaturedImage.URL); strings.HasPrefix(key, "media/") {
		art.FeaturedImage.StorageKey = key
		art.FeaturedImage.URL = m.storage.GetPublicURL(key)
	}

	for i := range art.Body {
		block := &art.Body[i]
		if block.Type == "image" || block.Type == "video" {
			if block.StorageKey != "" {
				block.Src = m.storage.GetPublicURL(block.StorageKey)
			} else if key := urlutil.ExtractStorageKey(block.Src); strings.HasPrefix(key, "media/") {
				block.StorageKey = key
				block.Src = m.storage.GetPublicURL(key)
			}
		}
	}
}

// ResolveArticles applies media resolution to a slice of articles.
func (m *mediaResolver) ResolveArticles(articles []models.Article) {
	for i := range articles {
		m.ResolveMedia(&articles[i])
	}
}

// NormalizeMediaKeys ensures storage keys are extracted from URLs so that
// articles remain decoupled from the physical CDN/host location.
func (m *mediaResolver) NormalizeMediaKeys(art *models.Article) {
	if art == nil {
		return
	}

	if art.FeaturedImage.StorageKey == "" && art.FeaturedImage.URL != "" {
		if key := urlutil.ExtractStorageKey(art.FeaturedImage.URL); strings.HasPrefix(key, "media/") {
			art.FeaturedImage.StorageKey = key
		}
	}

	for i := range art.Body {
		block := &art.Body[i]
		if block.Type == "image" || block.Type == "video" {
			if block.StorageKey == "" && block.Src != "" {
				if key := urlutil.ExtractStorageKey(block.Src); strings.HasPrefix(key, "media/") {
					block.StorageKey = key
				}
			}
		}
	}
}
