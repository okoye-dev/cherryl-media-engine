-- ==============================================================================
-- Cherryl Media Engine — Complete Combined Supabase Schema & Initial Migrations
-- Generated: 2026-10-15
-- Instructions: Paste and run directly in Supabase Dashboard -> SQL Editor -> Run
-- ==============================================================================

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ------------------------------------------------------------------------------
-- 1. Base Articles & Ingestion Tables (0001)
-- ------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS articles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    deck TEXT,
    category TEXT NOT NULL,
    cover_image TEXT,
    blocks JSONB NOT NULL DEFAULT '[]'::jsonb,
    read_time TEXT,
    status TEXT NOT NULL DEFAULT 'draft',
    tags TEXT[] NOT NULL DEFAULT '{}',
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_articles_slug ON articles(slug);
CREATE INDEX IF NOT EXISTS idx_articles_status ON articles(status);
CREATE INDEX IF NOT EXISTS idx_articles_category ON articles(category);
CREATE INDEX IF NOT EXISTS idx_articles_published_at ON articles(published_at DESC);

CREATE TABLE IF NOT EXISTS uploads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    storage_key TEXT NOT NULL UNIQUE,
    original_name TEXT NOT NULL,
    file_size BIGINT NOT NULL,
    content_type TEXT NOT NULL,
    public_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_uploads_key ON uploads(storage_key);

-- ------------------------------------------------------------------------------
-- 2. Discovery Sources & AI Story Candidates (0002)
-- ------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS sources (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    feed_url TEXT NOT NULL,
    site_url TEXT,
    category_hint TEXT,
    reliability_score DOUBLE PRECISION NOT NULL DEFAULT 0.8,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    last_fetched_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS suggested_stories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    headline TEXT NOT NULL,
    pitch TEXT NOT NULL,
    category TEXT NOT NULL,
    relevance_score DOUBLE PRECISION NOT NULL DEFAULT 0.5,
    sources JSONB NOT NULL DEFAULT '[]'::jsonb,
    extracted_facts JSONB NOT NULL DEFAULT '[]'::jsonb,
    suggested_draft JSONB NOT NULL DEFAULT '[]'::jsonb,
    status TEXT NOT NULL DEFAULT 'discovered',
    article_id UUID REFERENCES articles(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_suggested_stories_status ON suggested_stories(status);
CREATE INDEX IF NOT EXISTS idx_suggested_stories_category ON suggested_stories(category);
CREATE INDEX IF NOT EXISTS idx_suggested_stories_relevance ON suggested_stories(relevance_score DESC);

ALTER TABLE articles
    ADD COLUMN IF NOT EXISTS kicker TEXT,
    ADD COLUMN IF NOT EXISTS author_name TEXT DEFAULT 'Cherryl Staff',
    ADD COLUMN IF NOT EXISTS author_role TEXT DEFAULT 'Editorial Desk',
    ADD COLUMN IF NOT EXISTS sources JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS social_variants JSONB NOT NULL DEFAULT '{}'::jsonb,
    ADD COLUMN IF NOT EXISTS candidate_id UUID REFERENCES suggested_stories(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS ai_review_notes TEXT;

CREATE TABLE IF NOT EXISTS channel_publications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    article_id UUID NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    channel TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'draft',
    payload TEXT NOT NULL,
    external_post_id TEXT,
    external_post_url TEXT,
    error_message TEXT,
    scheduled_at TIMESTAMPTZ,
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_channel_publications_article ON channel_publications(article_id);
CREATE INDEX IF NOT EXISTS idx_channel_publications_status ON channel_publications(status);

-- ------------------------------------------------------------------------------
-- 3. Users & Editorial Sessions (0003)
-- ------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    full_name TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'editor',
    avatar_url TEXT,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    user_agent TEXT,
    ip_address TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token_hash);
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);

ALTER TABLE articles
    ADD COLUMN IF NOT EXISTS author_id UUID REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS published_by UUID REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_articles_author ON articles(author_id);
CREATE INDEX IF NOT EXISTS idx_articles_published_by ON articles(published_by);

-- ------------------------------------------------------------------------------
-- 4. Excerpt, Original URL & Urgency (0004)
-- ------------------------------------------------------------------------------
ALTER TABLE articles
    ADD COLUMN IF NOT EXISTS excerpt TEXT,
    ADD COLUMN IF NOT EXISTS original_url TEXT,
    ADD COLUMN IF NOT EXISTS is_lead BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS urgent BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_articles_is_lead ON articles(is_lead);
CREATE INDEX IF NOT EXISTS idx_articles_original_url ON articles(original_url);

-- ------------------------------------------------------------------------------
-- 5. Site Settings & Broadcast Copy (0005)
-- ------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS site_settings (
    key TEXT PRIMARY KEY,
    value JSONB NOT NULL DEFAULT '{}'::jsonb,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO site_settings (key, value)
VALUES (
    'editorial_copy',
    '{
        "hero": {
            "editionBadge": "DAILY EDITION",
            "titleLead": "Independent ",
            "titleAccent": "journalism",
            "titleTrail": " & digital media network.",
            "description": "Leading digital newsroom and new-age media distribution network covering Nigerian state affairs, business, culture, politics, and dispatches across Africa."
        },
        "ticker": {
            "label": "LIVE WIRE",
            "customNotice": "BREAKING: Federal Executive Council ratifies state electricity market guidelines and sovereign payment assurance frameworks."
        },
        "dossier": {
            "issueBadge": "WEEKEND EDITION · ISSUE #18",
            "heading": "Culture, Cinema & In-Depth Features",
            "subheading": "Features, music chronicles, food heritage, and modern Nigerian narratives."
        },
        "network": {
            "headline": "Multi-Channel Media Distribution Network",
            "subheadline": "Reach over 250,000+ readers and professionals daily across West Africa and the diaspora through verified news channels.",
            "subscribers": "250,000+",
            "reachMonthly": "1.8M+"
        },
        "footer": {
            "tagline": "New-Age Media Distribution & Independent Journalism",
            "locations": "Abuja · Lagos · London",
            "contactEmail": "newsroom@cherrylmedia.com",
            "copyrightNotice": "© 2026 Cherryl Media Enterprise. All rights reserved."
        }
    }'::jsonb
)
ON CONFLICT (key) DO NOTHING;

-- ------------------------------------------------------------------------------
-- 6. Curated Homepage Sections (0006)
-- ------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS curated_sections (
    section_key TEXT PRIMARY KEY,
    article_ids TEXT[] NOT NULL DEFAULT '{}',
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_curated_sections_updated_at ON curated_sections(updated_at DESC);

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
