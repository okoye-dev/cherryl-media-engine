package services

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/okoye-dev/cherryl-media-engine/internal/models"
	"github.com/okoye-dev/cherryl-media-engine/internal/store"
)

// FieldErrors maps a field path (for example "footer.contactEmail") to a message a person can read.
type FieldErrors map[string]string

// SettingsService manages the editable site text and other settings.
type SettingsService interface {
	GetAllSettings(ctx context.Context) (map[string]any, error)
	GetEditorialCopy(ctx context.Context) (models.EditorialCopy, error)
	// SaveEditorialCopy checks and stores the copy. It returns the saved copy, or field errors when something is invalid.
	SaveEditorialCopy(ctx context.Context, in models.EditorialCopy) (models.EditorialCopy, FieldErrors, error)
	// PatchEditorialCopy changes only the fields in patch (nested, for example {"footer":{"tagline":"..."}}) and checks the result.
	PatchEditorialCopy(ctx context.Context, patch map[string]any) (models.EditorialCopy, FieldErrors, error)
	SetSetting(ctx context.Context, key string, val []byte) error
}

type settingsService struct {
	store store.SettingsStore
}

// NewSettingsService constructs a new SettingsService.
func NewSettingsService(s store.SettingsStore) SettingsService {
	return &settingsService{store: s}
}

const editorialCopyKey = "editorial_copy"

// GetEditorialCopy returns the stored copy laid over the defaults, so a field that was never saved still has a value.
func (s *settingsService) GetEditorialCopy(ctx context.Context) (models.EditorialCopy, error) {
	copy := models.DefaultEditorialCopy()
	if s.store == nil {
		return copy, nil
	}
	val, err := s.store.GetSetting(ctx, editorialCopyKey)
	if err != nil {
		return copy, err
	}
	if len(val) == 0 {
		return copy, nil
	}
	// Decoding over the defaults only replaces keys that were saved; keys that are absent keep their default.
	if err := json.Unmarshal(val, &copy); err != nil {
		return models.DefaultEditorialCopy(), nil
	}
	return copy, nil
}

type limit struct {
	path string
	max  int
	get  func(c *models.EditorialCopy) *string
	trim bool
}

var emailPattern = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

var copyLimits = []limit{
	{"hero.editionBadge", 60, func(c *models.EditorialCopy) *string { return &c.Hero.EditionBadge }, true},
	{"hero.titleLead", 120, func(c *models.EditorialCopy) *string { return &c.Hero.TitleLead }, false},
	{"hero.titleAccent", 80, func(c *models.EditorialCopy) *string { return &c.Hero.TitleAccent }, true},
	{"hero.titleTrail", 120, func(c *models.EditorialCopy) *string { return &c.Hero.TitleTrail }, false},
	{"hero.description", 400, func(c *models.EditorialCopy) *string { return &c.Hero.Description }, true},
	{"ticker.label", 40, func(c *models.EditorialCopy) *string { return &c.Ticker.Label }, true},
	{"ticker.customNotice", 280, func(c *models.EditorialCopy) *string { return &c.Ticker.CustomNotice }, true},
	{"dossier.issueBadge", 80, func(c *models.EditorialCopy) *string { return &c.Dossier.IssueBadge }, true},
	{"dossier.heading", 120, func(c *models.EditorialCopy) *string { return &c.Dossier.Heading }, true},
	{"dossier.subheading", 300, func(c *models.EditorialCopy) *string { return &c.Dossier.Subheading }, true},
	{"footer.tagline", 160, func(c *models.EditorialCopy) *string { return &c.Footer.Tagline }, true},
	{"footer.locations", 120, func(c *models.EditorialCopy) *string { return &c.Footer.Locations }, true},
	{"footer.contactEmail", 120, func(c *models.EditorialCopy) *string { return &c.Footer.ContactEmail }, true},
	{"footer.copyrightNotice", 200, func(c *models.EditorialCopy) *string { return &c.Footer.CopyrightNotice }, true},
}

// requiredCopy lists the fields that must not be empty, because the page would look broken without them.
var requiredCopy = map[string]string{
	"hero.titleLead":         "Write the start of the headline",
	"hero.description":       "Write a short description",
	"dossier.heading":        "Write a heading for the Culture section",
	"footer.contactEmail":    "Enter a contact email",
	"footer.copyrightNotice": "Write the copyright line",
}

// CleanEditorialCopy trims spaces where they do not matter and returns a message for every problem found.
func CleanEditorialCopy(in models.EditorialCopy) (models.EditorialCopy, FieldErrors) {
	out := in
	errs := FieldErrors{}
	for _, l := range copyLimits {
		p := l.get(&out)
		if l.trim {
			*p = strings.TrimSpace(*p)
		}
		if utf8.RuneCountInString(*p) > l.max {
			errs[l.path] = "Keep this under " + strconv.Itoa(l.max) + " characters"
		}
		if msg, ok := requiredCopy[l.path]; ok && strings.TrimSpace(*p) == "" {
			errs[l.path] = msg
		}
	}
	if _, bad := errs["footer.contactEmail"]; !bad && !emailPattern.MatchString(out.Footer.ContactEmail) {
		errs["footer.contactEmail"] = "Enter a valid email address"
	}
	return out, errs
}

func (s *settingsService) SaveEditorialCopy(ctx context.Context, in models.EditorialCopy) (models.EditorialCopy, FieldErrors, error) {
	clean, errs := CleanEditorialCopy(in)
	if len(errs) > 0 {
		return in, errs, nil
	}
	raw, err := json.Marshal(clean)
	if err != nil {
		return in, nil, err
	}
	if s.store == nil {
		return clean, nil, errors.New("settings store is not configured")
	}
	if err := s.store.SetSetting(ctx, editorialCopyKey, raw); err != nil {
		return in, nil, err
	}
	return clean, nil, nil
}

// GetAllSettings returns every stored setting, with the editorial copy always complete.
func (s *settingsService) GetAllSettings(ctx context.Context) (map[string]any, error) {
	settings := map[string]any{}
	if s.store != nil {
		if all, err := s.store.GetAllSettings(ctx); err == nil && all != nil {
			settings = all
		}
	}
	copy, err := s.GetEditorialCopy(ctx)
	if err != nil {
		copy = models.DefaultEditorialCopy()
	}
	settings[editorialCopyKey] = copy
	return settings, nil
}

func (s *settingsService) SetSetting(ctx context.Context, key string, val []byte) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return errors.New("setting key is required")
	}
	if key == editorialCopyKey {
		return errors.New("use the editorial copy endpoint for this setting")
	}
	if s.store == nil {
		return errors.New("settings store is not configured")
	}
	return s.store.SetSetting(ctx, key, val)
}

// PatchEditorialCopy lays only the sent fields over what is stored, then checks and saves the whole.
func (s *settingsService) PatchEditorialCopy(ctx context.Context, patch map[string]any) (models.EditorialCopy, FieldErrors, error) {
	cur, err := s.GetEditorialCopy(ctx)
	if err != nil {
		return cur, nil, err
	}
	curRaw, err := json.Marshal(cur)
	if err != nil {
		return cur, nil, err
	}
	base := map[string]any{}
	if err := json.Unmarshal(curRaw, &base); err != nil {
		return cur, nil, err
	}
	mergeInto(base, patch)
	mergedRaw, err := json.Marshal(base)
	if err != nil {
		return cur, nil, err
	}
	var next models.EditorialCopy
	if err := json.Unmarshal(mergedRaw, &next); err != nil {
		return cur, FieldErrors{"body": "A field has the wrong type"}, nil
	}
	return s.SaveEditorialCopy(ctx, next)
}

// mergeInto copies src onto dst. Objects merge field by field; everything else replaces.
func mergeInto(dst, src map[string]any) {
	for k, v := range src {
		if sm, ok := v.(map[string]any); ok {
			if dm, ok := dst[k].(map[string]any); ok {
				mergeInto(dm, sm)
				continue
			}
		}
		dst[k] = v
	}
}
