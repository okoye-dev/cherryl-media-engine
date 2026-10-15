# Cherryl Media: Go Backend & Infrastructure Specification

## 1. Architectural Strategy
A unified, modular Go backend consisting of a single deployable binary with two run modes:
- `api`: HTTP REST/JSON service serving the Next.js frontend and editorial `/admin`.
- `worker`: Background cron worker orchestrating ingestion, deduplication, AI drafting, and social broadcasts.

No microservice complexity during initial rollout. Runs comfortably on a minimal \$4/month Linux VPS or container.

---

## 2. PostgreSQL Schema Specification

```sql
-- Sources allowlist
CREATE TABLE sources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(120) NOT NULL,
    feed_url TEXT NOT NULL UNIQUE,
    category VARCHAR(60) NOT NULL,
    reliability_tier INT DEFAULT 1,
    is_active BOOLEAN DEFAULT TRUE,
    last_polled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Raw ingested source articles
CREATE TABLE raw_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id UUID REFERENCES sources(id) ON DELETE CASCADE,
    external_id TEXT,
    title TEXT NOT NULL,
    url TEXT NOT NULL UNIQUE,
    content_hash CHAR(64) NOT NULL,
    raw_text TEXT NOT NULL,
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_raw_items_hash ON raw_items(content_hash);

-- Canonical curated stories
CREATE TABLE canonical_articles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(255) NOT NULL UNIQUE,
    title TEXT NOT NULL,
    kicker VARCHAR(120),
    deck TEXT,
    category VARCHAR(60) NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'draft', -- draft, in_review, approved, published, rejected
    author_name VARCHAR(120) NOT NULL,
    author_role VARCHAR(120) NOT NULL,
    read_time VARCHAR(30) DEFAULT '4 min read',
    body_blocks JSONB NOT NULL,
    source_links JSONB DEFAULT '[]'::jsonb,
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_canonical_slug ON canonical_articles(slug);
CREATE INDEX idx_canonical_category_status ON canonical_articles(category, status);

-- Platform-specific social variants
CREATE TABLE channel_variants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    article_id UUID REFERENCES canonical_articles(id) ON DELETE CASCADE,
    channel VARCHAR(30) NOT NULL, -- x, facebook, linkedin, newsletter
    headline TEXT,
    body TEXT NOT NULL,
    status VARCHAR(30) DEFAULT 'draft', -- draft, ready, published, failed
    external_post_id VARCHAR(255),
    error_message TEXT,
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_article_channel ON channel_variants(article_id, channel);
```

---

## 3. Core Modules & Engine Pipelines

### 3.1 Source Ingestion & Deduplication
- **Worker**: Concurrently polls active RSS feeds and official gazettes via Go channels (`sync.WaitGroup`).
- **Deduplication**: 
  - Generates SHA-256 hash of normalized headline and text body.
  - Computes SimHash / cosine similarity against stories published in the preceding 48 hours to group coverage duplicates.
  - Inserts distinct items into `raw_items`.

### 3.2 AI Grounded Drafting Worker
- Calls LLM with structured output schema (JSON mode).
- **Strict Grounding Rule**: Generation prompt enforces zero invention; only facts with verifiable sources are included.
- Outputs:
  1. Catchy editorial headline + kicker + deck.
  2. Medium-style content blocks (Drop-cap intro, section headings, pull quotes, inline photo suggestions).
  3. Tailored channel variants (X hook/thread, Facebook post, LinkedIn perspective, Newsletter digest).
- Sets initial status to `in_review`. **Never allows automated publishing directly to live.**

### 3.3 WordPress 1.2 GB DB & Media Migration Tool
- **DB Parser**: Go streaming parser for WordPress `wp_posts` and `wp_postmeta`.
  - Filters out `post_type != 'post'` and `post_status != 'publish'`.
  - Discards revisions, auto-drafts, and transients (shrinks 1.2 GB dump down to ~80 MB of active records).
- **Media Migration**:
  - Migrates images from legacy cPanel `wp-content/uploads/` to server storage.
  - Rewrites all legacy image path strings to `/uploads/` URLs.

---

## 4. Hosting & Infrastructure Cost Optimization

| Layer | Provider | Estimated Monthly Cost | Rationale |
| :--- | :--- | :--- | :--- |
| **Public Site & Admin** | Vercel Hobby / Self-hosted | **$0.00** | Edge SSR/SSG caching, unlimited builds |
| **Media Assets & Storage** | Direct Server / Volume Storage | **Included with VPS** | Direct file upload handling |
| **Go Backend + Postgres** | Hetzner Cloud (CX22 VPS) | **€3.79 (~$4.10)** | 2 vCPU, 4GB RAM, 40GB NVMe. Runs Go + Postgres with ample headroom |
| **Alternative DB/Host** | Neon Postgres Free + Fly.io / Railway | **$0.00 – $5.00** | Serverless scale-to-zero options for minimal traffic |
| **Total Operational Cost** | — | **~$4.10 / month** | 97% cheaper than Sanity ($99+/mo) |
