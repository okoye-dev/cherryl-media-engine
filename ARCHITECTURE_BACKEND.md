# Cherryl Media — Backend Architecture & Engine Engineering Blueprint

---

## 1. Product Mission & Editorial Vision

**Cherryl Media** is designed as a premier, high-conviction digital newsroom and editorial intelligence publication focused on Nigerian current affairs, state governance, commerce, culture, youth innovations, and lifestyle.

Modern African digital media often struggles with two extremes: sensationalist aggregation lacking verifiable source citations on one end, or stale institutional bureaucracy on the other. Cherryl Media resolves this gap with:

1. **Unapologetic Editorial Depth & Verification**: Every article supports structured inline and footer source citations, linking directly to official ministerial releases, regulatory gazettes (NERC, CBN, FEC), or direct interviews.
2. **Modular Block-Based Storytelling**: Long-form journalistic dispatches interspersed naturally with high-resolution imagery, direct MP4 video, YouTube embeds, pull quotes, and internal story cards.
3. **Structured Editorial Taxonomy**: Ten dedicated editorial pillars that reflect real reader interest without crowding navigation surfaces.
4. **Sub-10ms Go Engine**: A compiled, multi-threaded Go media engine backed by Supabase PostgreSQL and edge-ready Next.js 15 delivery.

---

## 2. Layered Clean Architecture & Dependency Injection

The backend is organized into four distinct architectural layers with strict **Dependency Inversion (DIP)**:

```
┌────────────────────────────────────────────────────────────────────────┐
│                   Layer 1: HTTP Transport & Routing                   │
│   (cmd/server/main.go & internal/handlers/*)                          │
│   - Go 1.22+ ServeMux with method routing                              │
│   - CORS, Auth middleware, and Standard HTTP Envelopes (httputil)     │
│   - Handlers perform zero direct SQL; only HTTP decoding & formatting  │
└────────────────────────────────────┬───────────────────────────────────┘
                                     │ Calls Service Interfaces
                                     ▼
┌────────────────────────────────────────────────────────────────────────┐
│                  Layer 2: Application & Domain Services                │
│   (internal/services/*)                                                │
│   - ArticleService: Slugs, tags, candidate triage, publishing rules    │
│   - SectionService: Curated slot population, reordering, validation    │
│   - SettingsService: Live Wire ticker, editorial fallback defaults     │
│   - MediaResolver: Dynamic CDN resolution & storage key normalization  │
└───────────────────┬────────────────────────────────┬───────────────────┘
                    │ Depends on Store Interfaces     │ Depends on Storage
                    ▼                                ▼
┌──────────────────────────────────────┐  ┌──────────────────────────────┐
│ Layer 3: Persistence Repositories    │  │ Layer 4: Object Storage      │
│ (internal/store/*)                   │  │ (internal/storage/*)         │
│ - PGStore (Supabase pgxpool)         │  │ - S3Storage (Cloudflare R2)  │
│ - MemoryStore (for tests & fallback) │  │ - Presigned Upload URLs      │
│ - Centralized RowScanners (scan.go)  │  │ - SHA256 Key Generation      │
└──────────────────────────────────────┘  └──────────────────────────────┘
```

### 2.1 Constructor Dependency Injection in `main.go`
Dependencies are created once at boot time and injected down the hierarchy:

```go
// cmd/server/main.go
pool, _ := store.NewPool(ctx, cfg.Database.URL)
pgStore := store.NewPGStore(pool)
s3Storage, _ := storage.NewS3Storage(&cfg.S3)

// Domain services receive stores and storage via interfaces
mediaResolver := services.NewMediaResolver(s3Storage)
articleService := services.NewArticleService(pgStore, mediaResolver)
sectionService := services.NewSectionService(pgStore, pgStore, mediaResolver)
settingsService := services.NewSettingsService(pgStore)

// HTTP Handlers receive domain services
articleHandler := handlers.NewArticleHandlerWithService(articleService)
sectionsHandler := handlers.NewSectionsHandlerWithService(sectionService)
settingsHandler := handlers.NewSettingsHandlerWithService(settingsService)
```

---

## 3. Official Editorial Taxonomy (10 Pillars)

The editorial system categorizes every dispatch into one of ten official pillars:

| # | Category Slug | Display Name | Editorial Scope |
|---|---|---|---|
| 1 | `state-affairs-environs` | **STATE AFFAIRS & ENVIRONS** | Sub-national policies, state government budgets, environmental issues, infrastructure projects across Nigeria. |
| 2 | `newsbeat` | **NEWSBEAT** | Breaking news, national alerts, high-urgency dispatches, legislative updates, and security investigations. |
| 3 | `brands-smes` | **BRANDS & SMEs** | Enterprise growth, fintech innovations, market consumer analysis, local manufacturing, startups. |
| 4 | `opinions` | **OPINIONS** | Thought leadership, guest columns, policy critiques, intellectual discourse, and economic essays. |
| 5 | `sports` | **SPORTS** | Nigerian Premier Football League (NPFL), Super Eagles, African athletics, combat sports, European football. |
| 6 | `entertainment-lifestyle` | **ENTERTAINMENT & LIFESTYLE** | Nollywood, Afrobeats music, red carpets, creative arts, fashion, dining, and cultural critiques. |
| 7 | `education` | **EDUCATION** | Universities, polytechnics, WAEC/JAMB reforms, academic research, polytechnic upgrades, scholarships. |
| 8 | `thrills-tricks` | **THRILLS & TRICKS** | Practical consumer advice, life hacks, digital navigation, curious phenomena, viral culture. |
| 9 | `youth-corner` | **YOUTH CORNER** | Gen-Z voices, student creators, campus technology, hackathons, emerging youth leadership. |
| 10 | `tour-hotels` | **TOUR & HOTELS** | Hospitality reviews, domestic travel destinations, eco-resorts, cultural heritage sites, weekend guides. |

---

## 4. Database Schema & Indexing Strategy (Supabase PostgreSQL)

### 4.1 Relational Schema (`articles`)
The article storage uses a hybrid relational + document model:

```sql
CREATE TABLE IF NOT EXISTS articles (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    slug TEXT UNIQUE NOT NULL,
    title TEXT NOT NULL,
    subtitle TEXT DEFAULT '',
    kicker TEXT DEFAULT '',
    deck TEXT DEFAULT '',
    excerpt TEXT DEFAULT '',
    category TEXT NOT NULL,
    category_name TEXT DEFAULT '',
    author_name TEXT NOT NULL,
    author_role TEXT DEFAULT 'Staff Reporter',
    author_avatar TEXT DEFAULT '',
    featured_image_url TEXT NOT NULL,
    featured_image_caption TEXT DEFAULT '',
    featured_image_credit TEXT DEFAULT 'Cherryl Media',
    read_time TEXT DEFAULT '4 min read',
    status TEXT NOT NULL DEFAULT 'draft',
    tags TEXT[] DEFAULT '{}',
    is_lead BOOLEAN DEFAULT FALSE,
    urgent BOOLEAN DEFAULT FALSE,
    send_webpushr BOOLEAN DEFAULT FALSE,
    seo_title TEXT DEFAULT '',
    seo_slug TEXT DEFAULT '',
    seo_description TEXT DEFAULT '',
    body JSONB NOT NULL DEFAULT '[]'::jsonb,
    related_stories JSONB NOT NULL DEFAULT '[]'::jsonb,
    sources JSONB DEFAULT '[]'::jsonb,
    social_variants JSONB DEFAULT '{}'::jsonb,
    published_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

### 4.2 High-Throughput Composite Indexing (Migration 0008)
To guarantee sub-10ms response times under concurrent reader queries, composite indexes target exact query access patterns:

```sql
-- Category feeds sorted by publication recency:
CREATE INDEX IF NOT EXISTS idx_articles_category_published_at
ON articles (category, published_at DESC);

-- Public timeline queries (status = 'published' ORDER BY published_at DESC):
CREATE INDEX IF NOT EXISTS idx_articles_status_published_at
ON articles (status, published_at DESC);

-- Slug lookup (instant O(1) B-tree lookup):
CREATE UNIQUE INDEX IF NOT EXISTS idx_articles_slug ON articles(slug);

-- Full-Text Search on article title, subtitle, and kicker:
CREATE INDEX IF NOT EXISTS idx_articles_fts_title
ON articles USING GIN (to_tsvector('english', title || ' ' || COALESCE(subtitle, '') || ' ' || COALESCE(kicker, '')));

-- Tag search typeahead & autocomplete:
CREATE INDEX IF NOT EXISTS idx_articles_tags_gin ON articles USING GIN(tags);
```

---

## 5. Media Pipeline & Storage Key Decoupling

Articles stored in PostgreSQL remain completely decoupled from specific CDN domain names or cloud hosts:

1. **Storage Keys in Database**: Image and video blocks store relative storage keys (e.g. `media/articles/2026/energy-grid.webp`).
2. **Dynamic Ingress / Egress via `MediaResolver`**:
   - **Ingress (`NormalizeMediaKeys`)**: Strips absolute CDN URLs from client payloads before database persistence.
   - **Egress (`ResolveMedia`)**: Dynamically attaches the configured Cloudflare R2 / S3 public base URL when serving JSON to readers.
3. **Presigned Upload Pipeline**:
   - Client sends `POST /api/uploads/presign` with filename and MIME type.
   - Backend generates a pre-authenticated S3 PUT URL valid for 15 minutes.
   - Client uploads directly to Cloudflare R2.
   - Client calls `POST /api/uploads/confirm` to log the asset.

---

## 6. Curated Homepage Sections Engine

The homepage layout is controlled through database-backed section curation (`curated_sections` table):

```sql
CREATE TABLE IF NOT EXISTS curated_sections (
    section_key TEXT PRIMARY KEY,
    article_ids TEXT[] NOT NULL DEFAULT '{}',
    metadata JSONB DEFAULT '{}'::jsonb,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

### Section Slots
- `hero_lead`: 1 primary lead story.
- `hero_secondary`: Up to 4 featured stories on the hero grid.
- `weekend_dossier`: Up to 3 in-depth weekend cultural features.
- `state_affairs`: Curated stories for State Affairs & Environs.
- `newsbeat`: Curated stories for National Newsbeat.
- `brands_smes`: Curated commercial stories for Brands & SMEs.
- `opinions`: Curated op-eds and guest columns.
- `sports`: Curated sports news and profiles.
- `entertainment_lifestyle`: Curated music, film, and fashion stories.
- `culture_features`: Curated long-form cultural dispatches.

When `GET /api/sections` is invoked:
1. `SectionService` reads all slot assignments from `curated_sections`.
2. Resolves all assigned article IDs in a single batch query (`GetByIDs`).
3. Uses `MediaResolver` to populate responsive image URLs.
4. Returns both an ordered `sections` array and a keyed `byKey` map for instant UI consumption.

---

## 7. Operational Health & Verification

The backend engine is verified with automated tests and live database connectivity:
- **Test Suite**: `go test ./...` passes cleanly across all packages (`handlers`, `services`, `auth`, `migrate`, `seed`, `urlutil`).
- **Production Server**: `go build -o /dev/null ./cmd/server` compiles with zero warnings.
- **Database Connection**: Supabase PostgreSQL connection pool verified with automated migration checks on boot.
