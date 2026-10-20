package services

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/okoye-dev/cherryl-media-engine/internal/store"
)

// SettingsService defines the interface for managing editorial configuration and copy.
type SettingsService interface {
	GetAllSettings(ctx context.Context) (map[string]any, error)
	GetEditorialCopy(ctx context.Context) (map[string]any, error)
	UpdateEditorialCopy(ctx context.Context, payload map[string]any) error
	SetSetting(ctx context.Context, key string, val []byte) error
	DefaultEditorialCopy() map[string]any
}

type settingsService struct {
	store store.SettingsStore
}

// NewSettingsService constructs a new SettingsService.
func NewSettingsService(s store.SettingsStore) SettingsService {
	return &settingsService{store: s}
}

// DefaultEditorialCopy returns the fallback copy matching the canonical website defaults.
func (s *settingsService) DefaultEditorialCopy() map[string]any {
	return map[string]any{
		"hero": map[string]any{
			"editionBadge": "DAILY EDITION",
			"titleLead":    "Independent ",
			"titleAccent":  "journalism",
			"titleTrail":   " & digital media network.",
			"description":  "Leading digital newsroom and new-age media distribution network covering Nigerian state affairs, business, culture, politics, and dispatches across Africa.",
		},
		"ticker": map[string]any{
			"label":        "LIVE WIRE",
			"customNotice": "BREAKING: Federal Executive Council ratifies state electricity market guidelines and sovereign payment assurance frameworks.",
		},
		"dossier": map[string]any{
			"issueBadge":  "WEEKEND EDITION · ISSUE #18",
			"heading":     "Culture, Cinema & In-Depth Features",
			"subheading":  "Features, music chronicles, food heritage, and modern Nigerian narratives.",
		},
		"network": map[string]any{
			"headline":     "Multi-Channel Media Distribution Network",
			"subheadline":  "Reach over 250,000+ readers and professionals daily across West Africa and the diaspora through verified news channels.",
			"subscribers":  "250,000+",
			"reachMonthly": "1.8M+",
		},
		"footer": map[string]any{
			"tagline":         "New-Age Media Distribution & Independent Journalism",
			"locations":       "Abuja · Lagos · London",
			"contactEmail":    "newsroom@cherrylmedia.com",
			"copyrightNotice": "© 2026 Cherryl Media Enterprise. All rights reserved.",
		},
	}
}

func (s *settingsService) GetAllSettings(ctx context.Context) (map[string]any, error) {
	settings, err := s.store.GetAllSettings(ctx)
	if err != nil || len(settings) == 0 {
		settings = map[string]any{
			"editorial_copy": s.DefaultEditorialCopy(),
		}
	}

	if _, ok := settings["editorial_copy"]; !ok {
		settings["editorial_copy"] = s.DefaultEditorialCopy()
	}

	return settings, nil
}

func (s *settingsService) GetEditorialCopy(ctx context.Context) (map[string]any, error) {
	val, err := s.store.GetSetting(ctx, "editorial_copy")
	if err != nil || len(val) == 0 {
		return s.DefaultEditorialCopy(), nil
	}

	var copyMap map[string]any
	if err := json.Unmarshal(val, &copyMap); err != nil {
		return s.DefaultEditorialCopy(), nil
	}

	return copyMap, nil
}

func (s *settingsService) UpdateEditorialCopy(ctx context.Context, payload map[string]any) error {
	rawBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return s.store.SetSetting(ctx, "editorial_copy", rawBytes)
}

func (s *settingsService) SetSetting(ctx context.Context, key string, val []byte) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil
	}
	return s.store.SetSetting(ctx, key, val)
}
