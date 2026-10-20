-- Migration 0007: Post Revamp, SEO, Subtitle, Webpushr, and Related Stories
ALTER TABLE articles ADD COLUMN IF NOT EXISTS subtitle TEXT;
ALTER TABLE articles ADD COLUMN IF NOT EXISTS seo_title TEXT;
ALTER TABLE articles ADD COLUMN IF NOT EXISTS seo_slug TEXT;
ALTER TABLE articles ADD COLUMN IF NOT EXISTS seo_description TEXT;
ALTER TABLE articles ADD COLUMN IF NOT EXISTS send_webpushr BOOLEAN DEFAULT FALSE;
ALTER TABLE articles ADD COLUMN IF NOT EXISTS related_stories JSONB DEFAULT '[]'::jsonb;
CREATE INDEX IF NOT EXISTS idx_articles_tags ON articles USING GIN(tags);
