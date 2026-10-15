-- Migration 0004: Parity for article excerpt, original URL, lead flag, and urgency
-- Purpose: Support newsroom archiving, external crawls, homepage feature pinning, and breaking news banners

ALTER TABLE articles
    ADD COLUMN IF NOT EXISTS excerpt TEXT,
    ADD COLUMN IF NOT EXISTS original_url TEXT,
    ADD COLUMN IF NOT EXISTS is_lead BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS urgent BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_articles_is_lead ON articles(is_lead);
CREATE INDEX IF NOT EXISTS idx_articles_original_url ON articles(original_url);
