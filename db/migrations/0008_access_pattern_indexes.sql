-- Migration 0008: Optimize access pattern indexes for digital publishing workload
-- 1. Category beat feeds sorted by publication date
CREATE INDEX IF NOT EXISTS idx_articles_category_published ON articles(category, published_at DESC);

-- 2. Status feeds (draft vs published) sorted by publication date
CREATE INDEX IF NOT EXISTS idx_articles_status_published ON articles(status, published_at DESC);

-- 3. Full-text search index for editorial title, kicker, deck, and excerpt
CREATE INDEX IF NOT EXISTS idx_articles_fulltext_search ON articles 
USING GIN(to_tsvector('english', title || ' ' || COALESCE(kicker, '') || ' ' || COALESCE(deck, '') || ' ' || COALESCE(excerpt, '')));
