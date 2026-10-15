-- Migration 0006: Curated Sections Table
-- Purpose: Persist section-to-article mappings for homepage and editorial layouts in the database

CREATE TABLE IF NOT EXISTS curated_sections (
    section_key TEXT PRIMARY KEY,
    article_ids TEXT[] NOT NULL DEFAULT '{}',
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_curated_sections_updated_at ON curated_sections(updated_at DESC);

-- Seed default initial section allocations mapping to canonical published articles
INSERT INTO curated_sections (section_key, article_ids)
VALUES
    ('hero_lead', ARRAY['2f405add-8bba-5261-99af-15feaa62754b']),
    ('hero_trending', ARRAY[
        '35588574-ec88-5a81-be90-8c46283c24de',
        'decd08ac-ef71-54ee-ac58-fd2480c007fa',
        'f10fad9c-c601-552e-a7b2-816b8c370ecc'
    ]),
    ('featured_updates', ARRAY[
        '25386df5-95e6-50ad-bc4c-3b6db589992d',
        '12f9e7fb-8019-5edb-80e9-7789a9729176',
        'd2065825-0408-5a7f-b4d0-559be9e67a9a',
        '278198f8-3eee-55ce-8736-dcdd55c46cec'
    ]),
    ('perspectives', ARRAY[
        '278198f8-3eee-55ce-8736-dcdd55c46cec',
        'd2065825-0408-5a7f-b4d0-559be9e67a9a',
        'f10fad9c-c601-552e-a7b2-816b8c370ecc'
    ]),
    ('dossier_lead', ARRAY['f3033f02-9628-52d2-bef4-8b0814c6d835']),
    ('dossier_soundscapes', ARRAY['d641aeae-a6be-50f5-a13c-54c67bbf6e75']),
    ('dossier_gastronomy', ARRAY['0827a7d0-015b-5643-ac3d-4ccced2b933c']),
    ('dossier_radar', ARRAY[
        'f2fc6ae3-16ec-51cc-bc8d-c9d2b4e4271f',
        'ee3b36cb-8ce0-5f3c-bfae-3d077b882004',
        '917d0345-9020-50e5-89eb-9cb1721d21fa',
        '33a00b82-fd1f-5437-9317-0c951151d023'
    ])
ON CONFLICT (section_key) DO UPDATE
SET article_ids = EXCLUDED.article_ids, updated_at = NOW();
