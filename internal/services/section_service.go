package services

import (
	"context"
	"errors"
	"strings"

	"github.com/okoye-dev/cherryl-media-engine/internal/models"
	"github.com/okoye-dev/cherryl-media-engine/internal/store"
)

// SectionService defines operations for curated homepage layout sections.
type SectionService interface {
	GetSections(ctx context.Context) ([]models.CuratedSection, map[string]models.CuratedSection, error)
	GetSectionByKey(ctx context.Context, key string) (*models.CuratedSection, error)
	UpdateSections(ctx context.Context, reqs []models.UpdateSectionRequest) error
}

type sectionService struct {
	sectionStore store.SectionStore
	articleStore store.ArticleStore
	resolver     MediaResolver
}

// NewSectionService constructs a new SectionService instance.
func NewSectionService(secStore store.SectionStore, artStore store.ArticleStore, res MediaResolver) SectionService {
	return &sectionService{
		sectionStore: secStore,
		articleStore: artStore,
		resolver:     res,
	}
}

func (s *sectionService) populateSection(ctx context.Context, sec *models.CuratedSection) error {
	if sec == nil {
		return nil
	}
	if len(sec.ArticleIDs) == 0 {
		sec.Articles = []models.Article{}
		return nil
	}
	arts, err := s.articleStore.GetByIDs(ctx, sec.ArticleIDs)
	if err != nil {
		return err
	}
	if s.resolver != nil {
		s.resolver.ResolveArticles(arts)
	}
	sec.Articles = arts
	return nil
}

func (s *sectionService) GetSections(ctx context.Context) ([]models.CuratedSection, map[string]models.CuratedSection, error) {
	sections, err := s.sectionStore.GetAllSections(ctx)
	if err != nil {
		return nil, nil, err
	}

	byKey := make(map[string]models.CuratedSection, len(sections))
	for i := range sections {
		_ = s.populateSection(ctx, &sections[i])
		byKey[sections[i].SectionKey] = sections[i]
	}

	return sections, byKey, nil
}

func (s *sectionService) GetSectionByKey(ctx context.Context, key string) (*models.CuratedSection, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, errors.New("section key cannot be empty")
	}

	sec, err := s.sectionStore.GetSection(ctx, key)
	if err != nil {
		return nil, err
	}

	_ = s.populateSection(ctx, sec)
	return sec, nil
}

func (s *sectionService) UpdateSections(ctx context.Context, reqs []models.UpdateSectionRequest) error {
	for _, req := range reqs {
		if req.SectionKey == "" {
			return errors.New("sectionKey is required")
		}
		if req.ArticleIDs == nil {
			req.ArticleIDs = []string{}
		}
		if err := s.sectionStore.SaveSection(ctx, req.SectionKey, req.ArticleIDs, req.Metadata); err != nil {
			return err
		}
	}
	return nil
}
