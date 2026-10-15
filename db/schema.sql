-- ==============================================================================
-- Cherryl Media Database Schema (Supabase PostgreSQL)
-- Run this in your Supabase SQL Editor or apply via Go migration runner
-- ==============================================================================

-- 1. Enable UUID generator
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- 2. Schema migrations table
CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 3. Trusted news & information sources
CREATE TABLE IF NOT EXISTS sources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    type TEXT NOT NULL DEFAULT 'rss', -- 'rss', 'scraper', 'api', 'social'
    feed_url TEXT NOT NULL UNIQUE,
    site_url TEXT,
    category_hint TEXT,
    reliability_score NUMERIC(3,2) NOT NULL DEFAULT 0.85,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    last_fetched_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sources_active ON sources(active);

-- 4. Raw ingested stories from source feeds/crawlers
CREATE TABLE IF NOT EXISTS ingested_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id UUID REFERENCES sources(id) ON DELETE SET NULL,
    external_id TEXT NOT NULL,
    title TEXT NOT NULL,
    raw_content TEXT,
    source_url TEXT NOT NULL,
    author TEXT,
    published_at TIMESTAMPTZ,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_source_external UNIQUE (source_id, external_id)
);

CREATE INDEX IF NOT EXISTS idx_ingested_items_created_at ON ingested_items(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_ingested_items_published_at ON ingested_items(published_at DESC);

-- 5. AI-suggested story candidates for the editorial discovery inbox
CREATE TABLE IF NOT EXISTS suggested_stories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    headline TEXT NOT NULL,
    pitch TEXT,
    category TEXT NOT NULL,
    relevance_score NUMERIC(3,2) NOT NULL DEFAULT 0.75,
    sources JSONB NOT NULL DEFAULT '[]'::jsonb,
    extracted_facts JSONB NOT NULL DEFAULT '[]'::jsonb,
    suggested_draft JSONB,
    status TEXT NOT NULL DEFAULT 'discovered', -- 'discovered', 'selected', 'drafted', 'in_review', 'rejected', 'published'
    article_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_suggested_stories_status ON suggested_stories(status);
CREATE INDEX IF NOT EXISTS idx_suggested_stories_category ON suggested_stories(category);
CREATE INDEX IF NOT EXISTS idx_suggested_stories_relevance ON suggested_stories(relevance_score DESC);

-- 3. Newsroom Users & Staff
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    full_name TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'editor', -- 'admin', 'editor', 'reporter'
    avatar_url TEXT,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

-- 4. Newsroom Auth Sessions
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

-- 5. Canonical publication articles
CREATE TABLE IF NOT EXISTS articles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    kicker TEXT,
    deck TEXT,
    category TEXT NOT NULL,
    cover_image TEXT,
    author_id UUID REFERENCES users(id) ON DELETE SET NULL,
    author_name TEXT DEFAULT 'Cherryl Staff',
    author_role TEXT DEFAULT 'Editorial Desk',
    blocks JSONB NOT NULL DEFAULT '[]'::jsonb,
    sources JSONB NOT NULL DEFAULT '[]'::jsonb,
    social_variants JSONB NOT NULL DEFAULT '{}'::jsonb,
    tags TEXT[] NOT NULL DEFAULT '{}',
    read_time TEXT,
    status TEXT NOT NULL DEFAULT 'draft', -- 'draft', 'in_review', 'approved', 'published'
    published_by UUID REFERENCES users(id) ON DELETE SET NULL,
    candidate_id UUID REFERENCES suggested_stories(id) ON DELETE SET NULL,
    ai_review_notes TEXT,
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_articles_slug ON articles(slug);
CREATE INDEX IF NOT EXISTS idx_articles_category ON articles(category);
CREATE INDEX IF NOT EXISTS idx_articles_status ON articles(status);
CREATE INDEX IF NOT EXISTS idx_articles_published_at ON articles(published_at DESC);

-- Link candidate back to article
ALTER TABLE suggested_stories
    DROP CONSTRAINT IF EXISTS fk_suggested_article,
    ADD CONSTRAINT fk_suggested_article FOREIGN KEY (article_id) REFERENCES articles(id) ON DELETE SET NULL;

-- 7. Media uploads tracking (Cloudflare R2 / S3)
CREATE TABLE IF NOT EXISTS uploads (
    storage_key TEXT PRIMARY KEY,
    file_name TEXT NOT NULL,
    size_bytes BIGINT NOT NULL,
    content_type TEXT NOT NULL,
    public_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_uploads_created_at ON uploads(created_at DESC);

-- 8. Multichannel distribution logs (X, Facebook, LinkedIn, Newsletter)
CREATE TABLE IF NOT EXISTS channel_publications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    article_id UUID NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    channel TEXT NOT NULL, -- 'x', 'facebook', 'linkedin', 'newsletter'
    status TEXT NOT NULL DEFAULT 'draft', -- 'draft', 'scheduled', 'publishing', 'published', 'failed'
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
