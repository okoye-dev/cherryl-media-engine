# Cherryl Media — Product Architecture & System Blueprint

> **The Definitive Product & Engineering Reference for Cherryl Media**
> Combining high-conviction African digital journalism with a sub-10ms Go media engine and edge-rendered Next.js 15 delivery.

---

## 1. Executive Product Vision & Value Proposition

**Cherryl Media** (`cherrylmedia.com`) is an independent, high-standard digital publication and news distribution network focusing on Nigerian current affairs, state government policies, enterprise commerce, cultural chronicles, and youth innovation across West Africa and the diaspora.

### The Problem in African Digital Media
1. **Sensationalist Aggregation**: Ad-bloated blogs aggregating unverified rumors with no primary source citations.
2. **Bureaucratic Stagnation**: Official state websites with inaccessible layouts, broken search, and sluggish server response times.
3. **Rigid Content Systems**: Legacy CMS platforms (e.g. monolithic WordPress) with brittle plugins, slow TTFB, and fragmented media hosting.

### The Cherryl Media Solution
- **Journalistic Verification First**: Every article supports structured inline and footer source citations, linking directly to verified official releases, regulatory gazettes (NERC, CBN, FEC), or direct interviews.
- **Dynamic Block-Based Canvas**: Stories are composed as ordered semantic blocks (`paragraph`, `heading`, `pullquote`, `image`, `video`, `link card`) with explicit order IDs.
- **Clean 10-Pillar Taxonomy**: Ten editorial pillars reflecting real reader interest with responsive navigation ergonomics that never squish across mobile, tablet, or desktop.
- **Enterprise-Grade Decoupled Engine**: Next.js 15 ISR frontend + compiled Go media engine + Supabase PostgreSQL + Cloudflare R2 object storage.

---

## 2. End-to-End System Architecture

```mermaid
flowchart TD
    subgraph Readers ["Readers & Public Clients"]
        WebReaders["Desktop / Tablet / Mobile Readers"]
        SocialFeeds["Syndicated Channels (X, FB, LinkedIn, NL)"]
    end

    subgraph Edge ["Edge Delivery & CDN Layer"]
        CF["Cloudflare Edge Network"]
        NextISR["Next.js 15 App Router (ISR 30s)"]
    end

    subgraph Admin ["Newsroom Staff & Editorial Desk"]
        Editors["Journalists & Desk Editors (/admin/editor)"]
        Curators["Layout Curators (/admin/sections)"]
        Admins["Managing Editors (/admin/settings)"]
    end

    subgraph Engine ["Cherryl Media Engine (Go Backend :8080)"]
        Router["HTTP Router & Middleware (httputil)"]
        subgraph Services ["Application Services Layer (Dependency Inversion)"]
            ArticleSvc["ArticleService (Slugs, Tags, Validation)"]
            SectionSvc["SectionService (Curated Layouts)"]
            SettingsSvc["SettingsService (Site Copy & Defaults)"]
            MediaRes["MediaResolver (URL Egress / Ingress)"]
            AuthSvc["AuthVerifier (Supabase ECC + HMAC)"]
        end
    end

    subgraph StorageLayer ["Persistence & Object Storage"]
        PG[("Supabase PostgreSQL\n- articles (JSONB blocks & tags)\n- curated_sections\n- settings\n- users & leads")]
        R2[("Cloudflare R2 (S3)\n- /media/articles/* (16:9 Photos)\n- /media/videos/* (MP4s)")]
    end

    WebReaders -->|HTTPS| CF
    CF -->|Cache Miss / Revalidate| NextISR
    NextISR -->|REST Read API| Router
    SocialFeeds -.->|Share Links| WebReaders

    Editors & Curators & Admins -->|Authenticated Sessions| NextISR
    NextISR -->|REST Mutating API| Router
    Editors -->|Direct PUT Presigned| R2

    Router --> Services
    ArticleSvc -->|Composite Index Queries| PG
    SectionSvc -->|Batch Fetch Articles| PG
    SettingsSvc -->|Get / Set Settings| PG
    MediaRes <-->|Hydrate Public URLs| R2
    AuthSvc -->|Verify Signatures| PG
```

---

## 3. Official Editorial Taxonomy (10 Pillars)

All publication output is organized under ten canonical pillars, designed to provide comprehensive national coverage while maintaining clear navigation hierarchy:

| # | Pillar Slug | Display Name | Sub-Beats & Scope |
|:---:|---|---|---|
| **01** | `state-affairs-environs` | **STATE AFFAIRS & ENVIRONS** | Sub-national governance, 36 state executive councils, regional rail/power developments, flood & environmental alerts. |
| **02** | `newsbeat` | **NEWSBEAT** | Breaking national alerts, legislative gazettes, judicial rulings, public security, federal policy dispatches. |
| **03** | `brands-smes` | **BRANDS & SMEs** | Enterprise growth, fintech innovations, consumer market pricing, local manufacturing, startups, commercial trade. |
| **04** | `opinions` | **OPINIONS** | Policy critiques, guest columns, academic commentary, economic analysis, philosophical & cultural essays. |
| **05** | `sports` | **SPORTS** | NPFL football, Super Eagles/Super Falcons, African athletics, European leagues, combat sports, grassroots talent. |
| **06** | `entertainment-lifestyle` | **ENTERTAINMENT & LIFESTYLE** | Nollywood features, Afrobeats albums & live concerts, culinary arts, fashion trends, creative industry interviews. |
| **07** | `education` | **EDUCATION** | Tertiary institutions, polytechnic reforms, WAEC/JAMB policy shifts, university research, student scholarships. |
| **08** | `thrills-tricks` | **THRILLS & TRICKS** | Practical consumer advice, life hacks, digital navigation, financial security, internet culture & curiosities. |
| **09** | `youth-corner` | **YOUTH CORNER** | Gen-Z creators, tech hackathons, campus politics, student ventures, emerging youth leadership dispatches. |
| **10** | `tour-hotels` | **TOUR & HOTELS** | Domestic travel destinations, hospitality reviews, eco-tourism resorts, cultural heritage, weekend getaway guides. |

### Responsive Navigation Rules
- **Desktop (Large Screens $\ge$ 1280px)**: Prominent top nav displays the 5 highest-traffic daily beats (`State Affairs`, `Newsbeat`, `Brands & SMEs`, `Opinions`, `Sports`), with secondary pillars in the menu drawer.
- **Laptop / Small Desktop (1024px – 1279px)**: Font sizes scale proportionally to prevent squishing or line wrapping.
- **Tablet & Mobile ($<$ 1024px)**: Top bar renders clean logo, live wire status indicator, and circular navigation icon. Full 10-pillar index expands smoothly in the slide-out menu drawer.

---

## 4. End-to-End User Journeys

### 4.1 The Reader Journey
1. **Landing & Instant Discovery**: Reader opens `cherrylmedia.com`. Page is served immediately from edge cache ($< 50\text{ms}$ TTFB).
2. **Live Wire Awareness**: Continuous ticker notice broadcasts high-priority breaking news alerts.
3. **Immersive Article Consumption**:
   - High-resolution 16:9 cover image with photographer credits.
   - Clear headline, deck, author credentials, publication timestamp, and read-time badge.
   - Text rendered with high legibility typography (65–75 characters per line).
   - Dynamic media blocks (full-bleed images, direct MP4 video, pullquotes, inline link cards).
4. **Trust & Verification**: Reader scrolls to the **Primary Sources** box to inspect verified source documents (NERC orders, gazettes, interviews) with trust badges.
5. **Deeper Exploration**: Curated related stories and "More Stories" cards guide reader deeper into the publication.

### 4.2 The Journalist / Staff Writer Journey
1. **Authentication**: Writer logs into `/admin/login` using email credentials or Supabase Auth.
2. **Story Inception**:
   - Option A: Click **"+ Write New Story"** for original reporting.
   - Option B: Select an ingested lead from the **Discovery Inbox** on the dashboard.
3. **Authoring in Article Studio (`/admin/editor`)**:
   - **Headline & Deck**: Type title; system generates URL slug.
   - **Slug Uniqueness**: Real-time debounced check verifies slug availability against PostgreSQL (displays green check or red conflict warning).
   - **Category & Tags**: Pick from the 10 pillars; typeahead tag search suggests existing legacy tags.
   - **Cover Photo Upload**: Drag-and-drop image directly triggers presigned Cloudflare R2 upload (bypassing backend CPU).
   - **Content Composition**: Add and reorder semantic blocks (paragraphs, section headers, pullquotes with attribution, inline photos, MP4 video).
   - **Internal Story Cards**: Search publication database to embed interactive cards linking related stories.
   - **Multi-Channel Copy**: Prepare pre-formatted blurbs for X (Twitter), Facebook, LinkedIn, and the newsletter.
4. **Pre-Publish Checklist**:
   - Reviewer modal confirms title, slug availability, cover image presence, and word count.
   - Click **"Confirm & Publish Story"** -> Backend sets `status: "published"`, records timestamp, and updates PostgreSQL.
   - Story becomes live globally within seconds.

### 4.3 The Layout Curator Journey (`/admin/sections`)
1. Curator opens `/admin/sections` to manage homepage layout.
2. Views real-time slot fill counts for all layout blocks (Hero Lead, Hero Secondary, Weekend Dossier, State Affairs, Newsbeat, etc.).
3. Uses ↑ and ↓ controls to reorder stories within slots.
4. Clicks **"+ Add Story Slot"** or **"Replace"**:
   - Searchable modal opens.
   - Searches across all published articles by headline, kicker, or category.
   - Selects story -> slot instantly updates.
5. Clicks **"Save Curated Layout"** -> Batch payload persists to `curated_sections` table in PostgreSQL.

### 4.4 The Managing Editor Journey (`/admin/settings`)
1. Editor opens `/admin/settings` to adjust global newsroom announcements.
2. Tabs provide dedicated controls:
   - **Homepage Hero**: Edition badge, title lead, crimson italic accent, and description.
   - **Live Wire / Ticker**: Urgent breaking notice text and ticker label.
   - **Weekend Dossier**: Dossier issue badge and heading.
   - **Distribution Network**: Syndicated readership statistics (subscribers, monthly reach).
   - **Masthead & Footer**: Bureau locations (Abuja, Lagos, London), newsroom email, copyright notices.
3. Clicks **"Save Changes"** -> Settings persist to PostgreSQL, with instant fallback to default canonical values.

### 4.5 The Publish-to-Homepage Pipeline & Media Confirmation Invariant
Before an article transitions from `draft` to `published`, an automated validation sequence guarantees editorial completeness and media integrity:
1. **Media Invariant Verification**:
   - The Go engine scans `featured_image.storage_key` and every `image` and `video` block in the article body for assets with keys in the `media/` namespace.
   - Every key must have `uploads.status = 'confirmed'` in the database.
   - If any asset is unconfirmed or still in-flight, `PublishArticle` rejects the request with HTTP 400 (`cannot publish article: media asset "..." is not confirmed uploaded`).
2. **Draft Persistence & Recovery**:
   - Writers can click **"Save Draft"** at any time. Drafts are safely stored in PostgreSQL and mirrored in browser local storage.
   - The newsroom dashboard (`/admin`) features dedicated status tabs (`All`, `Published`, `Drafts`), visual metric badges, and one-click "Resume Draft →" shortcuts.
3. **Automated Notification Dispatch**:
   - Upon successful publication, the engine queries active subscribers in `post_subscribers` matching the article's category pillar.
   - Dispatch jobs are recorded in `post_email_dispatches` for asynchronous email delivery and audit logging.

### 4.6 The Featured Section Decision Engine & Human Curation UI
The editorial homepage uses an automated decision engine paired with human override capabilities:
1. **Human Override Priority**:
   - Desk editors can flag any article as featured via the Article Studio or through the Section Curator (`POST /api/articles/{id}/featured`).
   - Flagged stories (`is_featured = true`) immediately take precedence as the homepage lead.
2. **Automated Decision Engine Fallback**:
   - If no story is manually pinned, the engine evaluates candidates using a 4-tier heuristic:
     1. Most recent published story with `is_lead = true`.
     2. Highest editorial completeness: most recent published story having both a 16:9 cover photograph and an editorial deck/subtitle.
     3. Most recently published article overall.
3. **Curator Gist Search**:
   - In `/admin/sections`, editors searching for stories can filter by pillar section and sort by recency or relevance.
   - Results show full editorial "gists": 16:9 thumbnail, headline, subtitle/deck, kicker, author, publication date, status badge, and a one-click "★ Set as Featured" toggle.

### 4.7 Data-Backed Homepage Scroll Flow (Top 4 Pillars)
To preserve magazine elegance and prevent vertical visual fatigue, the homepage scroll does not render all 10 pillars consecutively. The scroll hierarchy is informed by analysis of the 28,000+ legacy post archive:
- **State Affairs & Environs**: 10,033 posts (~36% of all content) — Section 1 in homepage scroll.
- **Newsbeat**: 9,946 posts (~35% of all content) — Section 2 in homepage scroll.
- **Brands & SMEs**: 2,063 posts (~7.5% of all content) — Section 3 in homepage scroll.
- **Opinions**: 2,027 posts (~7.5% of all content) — Section 4 in homepage scroll.
Together, these four pillars account for over 86% of the publication's content. The remaining six pillars (`Sports`, `Entertainment & Lifestyle`, `Education`, `Thrills & Tricks`, `Youth Corner`, `Tour & Hotels`) are accessible via the top navigation, mobile drawer, and dedicated category hub pages.

---

## 5. Data Architecture & Hybrid Storage Model

### 5.1 Why a Hybrid Relational + Document Model?
- **Relational Schema**: Columns for primary identifiers (`id`, `slug`, `category`, `status`, `published_at`, `author_name`) enable fast indexing, foreign keys, and sub-10ms sorting.
- **JSONB Document Blocks**: The `body`, `related_stories`, `sources`, and `social_variants` columns store structured content as JSONB. This allows rich block types without complex relational join tables (`article_blocks`, `block_attributes`).

### 5.2 Content Block Specification (`body` JSONB)
Every content block adheres to a strict discriminated union:

```typescript
type ContentBlock =
  | { order_id: number; type: "paragraph"; text: string; isDropCap?: boolean }
  | { order_id: number; type: "heading"; level: 2 | 3 | 4; text: string }
  | { order_id: number; type: "quote"; quote: string; author: string; context?: string }
  | { order_id: number; type: "image"; src: string; storageKey?: string; caption?: string; credit?: string }
  | { order_id: number; type: "video"; src: string; storageKey?: string; poster?: string; caption?: string }
  | { order_id: number; type: "link"; title: string; url: string; publisher?: string; deck?: string; isInternal?: boolean };
```

---

## 6. Security, Authentication & Role Hierarchy

```
┌───────────────────────────────────────────────────────────────┐
│                      Authentication Gateway                   │
├──────────────────────────────┬────────────────────────────────┤
│ Local HMAC JWT Verifier      │ Supabase ECC JWKS Verifier     │
│ (HS256 with JWTSecret)       │ (ES256 with KeyID / JWKS URL)  │
└──────────────────────────────┴────────────────────────────────┘
                               │
                               ▼
┌───────────────────────────────────────────────────────────────┐
│                   Role-Based Authorization                     │
├───────────────┬───────────────────────────────────────────────┤
│ Role          │ Privileges                                    │
├───────────────┼───────────────────────────────────────────────┤
│ Admin         │ Full access: Users, Settings, Sections, Posts  │
│ Editor        │ Publish stories, curate layout slots, settings│
│ Reporter      │ Draft stories, upload media, preview articles  │
└───────────────┴───────────────────────────────────────────────┘
```

- **Token Storage**: JWT tokens are stored in the client's `localStorage` and automatically injected as `Authorization: Bearer <token>` on all API requests via `ApiClient`.
- **401 Interception**: The frontend service container detects unauthorized responses and automatically clears session state to redirect to `/admin/login`.

### 6.1 Security Threat Model & Defense-in-Depth Hardening

Following a comprehensive code security audit, the platform implements multi-layer defense mechanisms:

1. **Anti-DoS Request Bounds**:
   - Every mutating HTTP endpoint wraps `r.Body` in `http.MaxBytesReader` to strictly bound payload size (Articles: 10MB, Media Uploads: 100MB, JSON Settings/Batches: 2–5MB, Auth & Subscribers: 64KB).
   - The Go HTTP server configures explicit connection timeouts (`ReadTimeout: 15s`, `ReadHeaderTimeout: 10s`, `WriteTimeout: 30s`, `IdleTimeout: 60s`) neutralizing Slowloris and connection starvation attacks.

2. **Media Upload & Execution Isolation**:
   - Strict MIME and file extension validation enforces safe editorial formats (`JPEG`, `PNG`, `WebP`, `GIF`, `AVIF`, `MP4`, `WebM`, `MP3`, `PDF`).
   - Dangerous executable formats (`.exe`, `.sh`, `.php`, `.js`, `.html`, `.svg`) are blocked. Raw SVG upload is explicitly barred to eliminate Stored XSS vectors via inline SVG script tags.
   - S3 presigned URLs offload asset uploads directly to Cloudflare R2, bypassing backend CPU and memory.

3. **HTTP Security Response Headers**:
   - `X-Content-Type-Options: nosniff` prevents MIME type sniffing and confusion exploits.
   - `X-Frame-Options: SAMEORIGIN` blocks UI clickjacking inside hostile iframes.
   - `Referrer-Policy: strict-origin-when-cross-origin` protects sensitive referral parameters.
   - `Permissions-Policy: camera=(), microphone=(), geolocation=()` disables unnecessary browser capabilities.

4. **Dynamic Origin CORS**:
   - Replaced wildcard CORS with dynamic origin validation allowing trusted domains (`localhost`, `*.cherrylmedia.com`, `*.vercel.app`) with `Vary: Origin` and `Access-Control-Allow-Credentials: true`.

5. **SQL Injection & Invariant Safety**:
   - 100% of database queries use PostgreSQL parameterized placeholders (`$1`, `$2`, etc.) via `pgxpool`.
   - Slug uniqueness checks validate UUID structure (`uuid.Parse`) before attempting UUID casting in PostgreSQL queries.

6. **Client-Side XSS Protection**:
   - Article body content is rendered through React JSX component structures (`ArticleBody.tsx`), guaranteeing automatic string escaping for all text, headings, captions, and quotes.
   - YouTube embeds validate strict 11-character alphanumeric IDs and render only via `https://www.youtube-nocookie.com/embed/...`.

---

## 7. Operational Health, CI/CD & Deployments

| Component | Target Runtime | Test Command | Build Command | Latency SLA |
|---|---|---|---|---|
| **Frontend** | Node.js / Edge (Next.js 15) | `npm test` | `npm run build` | $< 50\text{ms}$ (ISR) |
| **Backend** | Compiled Go Binary | `go test ./...` | `go build ./cmd/server` | $< 10\text{ms}$ (REST) |
| **Database** | Supabase PostgreSQL | `cmd/verify-db` | `cmd/migrate` | $< 5\text{ms}$ (Indexed) |
| **Media** | Cloudflare R2 / S3 | N/A | Direct Uploads | CDN Edge Cached |

---

## 8. Zero-Mock Production Architecture & Database Synchronization

To guarantee end-to-end operational fidelity, Cherryl Media strictly enforces a **Zero-Mock Policy**: no dummy articles, hardcoded story lists, or simulated payloads remain in the codebase. Every story, section, image asset, and configuration is retrieved live from PostgreSQL and Cloudflare R2.

### 8.1 Complete Excision of In-Memory Mock Datasets
- **1,389-Line Mock Purge**: The legacy in-memory `CANONICAL_ARTICLES` mock constant was removed across the entire frontend repository.
- **Pure Database Resolution**:
  - `getAllArticles()`, `getArticleBySlug()`, `getArticlesByCategory()`, and `getDispatches()` fetch directly from `${BACKEND_URL}/api/articles`.
  - `fetchCuratedSections()` queries the Go engine `/api/sections` endpoint, populating slots directly from `curated_sections` table allocations.
  - Slug availability checking (`/api/articles/check-slug`) and tag typeahead (`/api/tags`) route directly to the backend database.
  - The admin dashboard (`/admin`), article editor (`/admin/editor`), and section curator (`/admin/sections`) initialize with empty states and load dynamic data on mount via service hooks.

### 8.2 Database Verification & Media Integrity Metrics
- **PostgreSQL Database Verification** (`go run cmd/verify-db/main.go`):
  - **Articles**: 24 total articles stored in Supabase PostgreSQL (23 published stories, 1 active draft story).
  - **Media Confirmation**: 100% of published stories have `media_confirmed = true`.
  - **Uploads**: 43 media assets stored and confirmed (`status = 'confirmed'`) in the `uploads` table, linking directly to Cloudflare R2 CDN webp/mp4 files.
  - **Curated Sections**: 8 curated homepage sections persisted in `curated_sections`.
  - **Site Settings**: Active global site copy and editorial announcements in `site_settings`.
- **Static & ISR Build Coverage**:
  - Next.js 15 production build generates **49/49** static and ISR pages cleanly with zero compilation warnings or type errors (`npm run build`).
