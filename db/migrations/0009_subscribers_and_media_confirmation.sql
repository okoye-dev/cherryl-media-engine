-- Migration 0009: Subscribers, email notifications, media confirmation, and featured flags

-- 1. Ensure uploads status exists
ALTER TABLE uploads ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'confirmed';

-- 2. Add is_featured and media_confirmed to articles
ALTER TABLE articles ADD COLUMN IF NOT EXISTS is_featured BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE articles ADD COLUMN IF NOT EXISTS media_confirmed BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_articles_is_featured ON articles(is_featured, published_at DESC);

-- 3. Notification Subscribers table
CREATE TABLE IF NOT EXISTS post_subscribers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL DEFAULT '',
    categories TEXT[] NOT NULL DEFAULT '{}', -- Empty array denotes all categories
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_post_subscribers_active ON post_subscribers(active);
CREATE INDEX IF NOT EXISTS idx_post_subscribers_email ON post_subscribers(email);

-- 4. Post Email Dispatches queue / logs table
CREATE TABLE IF NOT EXISTS post_email_dispatches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    article_id UUID NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    recipient_email TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending', -- 'pending', 'queued', 'sent', 'failed'
    error_message TEXT DEFAULT '',
    sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_email_dispatches_article ON post_email_dispatches(article_id);
CREATE INDEX IF NOT EXISTS idx_email_dispatches_status ON post_email_dispatches(status);
