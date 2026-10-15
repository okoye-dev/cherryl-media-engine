-- Migration 0002: Sources Ingestion, AI Story Discovery, and Multichannel Distribution

-- 1. Trusted news & information sources
CREATE TABLE IF NOT EXISTS sources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    type TEXT NOT NULL DEFAULT 'rss', -- 'rss', 'scraper', 'api', 'youtube', 'social'
    feed_url TEXT NOT NULL UNIQUE,
    site_url TEXT,
    category_hint TEXT,
    reliability_score NUMERIC(3,2) NOT NULL DEFAULT 0.85,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    last_fetched_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sources_active ON sources(active);

-- 2. Raw ingested stories from source feeds/crawlers
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

-- 3. AI-suggested story candidates for the editorial discovery inbox
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
    article_id UUID REFERENCES articles(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_suggested_stories_status ON suggested_stories(status);
CREATE INDEX IF NOT EXISTS idx_suggested_stories_category ON suggested_stories(category);
CREATE INDEX IF NOT EXISTS idx_suggested_stories_relevance ON suggested_stories(relevance_score DESC);

-- 4. Extend articles table for newsroom traceability and multichannel publishing
ALTER TABLE articles
    ADD COLUMN IF NOT EXISTS kicker TEXT,
    ADD COLUMN IF NOT EXISTS author_name TEXT DEFAULT 'Cherryl Staff',
    ADD COLUMN IF NOT EXISTS author_role TEXT DEFAULT 'Editorial Desk',
    ADD COLUMN IF NOT EXISTS sources JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS social_variants JSONB NOT NULL DEFAULT '{}'::jsonb,
    ADD COLUMN IF NOT EXISTS candidate_id UUID REFERENCES suggested_stories(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS ai_review_notes TEXT;

-- 5. Multichannel publication tracking (X, Facebook, LinkedIn, Newsletter)
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
