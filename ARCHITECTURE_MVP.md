# Cherryl Media — Product MVP Flow & Architecture

> **Focus:** Lightweight, robust architecture for editorial news publishing, admin authentication, image uploads, and AI-assisted workflows.

---

## 1. System Architecture Diagram

```mermaid
flowchart TD
    subgraph Frontend ["Frontend (Next.js 15)"]
        PublicSite["Public Reader Site\n(/, /article/:slug, /category/:slug)"]
        AdminStudio["Admin Studio Newsroom\n(/admin, /admin/editor, /admin/login)"]
    end

    subgraph Backend ["Backend Engine (Go / Gin)"]
        AuthService["Auth Service (JWT / Sessions)"]
        ArticleService["Article & Editorial Service"]
        UploadService["S3 Presign Upload Service"]
        IngestWorker["News Feeds & Ingestion Worker"]
    end

    subgraph Storage ["Infrastructure & Storage"]
        DB[("Supabase PostgreSQL\n(Auth, Articles, Sources, Distribution)")]
        R2[("Cloudflare R2 (S3 Compatible)\nPublic Images Bucket")]
    end

    PublicSite -->|GET Published Articles| ArticleService
    AdminStudio -->|Admin Login / Session| AuthService
    AdminStudio -->|Autosave & Publish| ArticleService
    AdminStudio -->|Request Presigned Upload URL| UploadService
    UploadService -.->|Direct Client Upload (PUT)| R2
    R2 -.->|Public CDN URL| AdminStudio

    AuthService --> DB
    ArticleService --> DB
    IngestWorker --> DB
```

---

## 2. Core MVP Flows

### A. Admin Authentication Flow
1. **Login Route (`/admin/login`):**
   - Editor submits email + password.
   - Go backend verifies bcrypt hash against the `users` table.
   - Returns a secure, `HttpOnly` JWT cookie (or session token).
2. **Route Protection:**
   - Next.js middleware checks for the session cookie on all `/admin/*` routes (redirecting unauthenticated users to `/admin/login`).
   - Go backend validates JWT on all mutating `/api/articles` and `/api/uploads` endpoints.

### B. Story Editorial & Publishing Flow
1. **Discovery / Drafting:**
   - Editor either clicks **"+ Write New Story"** or picks an item from the **"Story Leads / Discovery Inbox"**.
   - Progressive block-based editor expands vertically as editor types (paragraphs, headings, pull quotes, images, links).
2. **Autosave:**
   - Every 2.5 seconds of inactivity, the editor sends a debounced autosave draft payload to `POST /api/articles` (`status: "draft"`).
3. **Publishing Validation:**
   - Cover photo, headline, and at least 1 verified source citation are required before publishing.
   - Editor clicks **"Publish Story"** -> Backend sets `status: "published"`, records `published_at = NOW()`, and records `published_by = user.id`.
4. **Public Site Rendering:**
   - Public pages query only `status = 'published'` with `published_at <= NOW()`.

### C. Image Upload Flow (Cloudflare R2 / S3 Presigned)
1. Editor selects image in `<ImageUploadField />`.
2. Frontend requests presigned URL: `POST /api/uploads/presign` with `{ fileName, contentType, sizeBytes }`.
3. Backend generates S3 PUT URL for Cloudflare R2 bucket with 15-minute expiration.
4. Browser uploads binary directly to Cloudflare R2 via HTTP `PUT` (bypassing backend server load).
5. Frontend stores public Cloudflare CDN URL (`https://pub-xxx.r2.dev/...`) on the article record.

### D. AI Ingestion & Curation Flow (Phase 2 Ready)
1. Background cron worker fetches news feeds from `sources` table into `ingested_items`.
2. Discovery engine deduplicates and creates `suggested_stories` (shown in Admin Discovery Inbox).
3. Reviewer clicks **"Copy Context for AI"** in the editor: full structured briefing (title, verified facts, sources, drafts) copies to clipboard for vetting in ChatGPT/Claude/Gemini.

---

## 3. Database Schema Audit & Additions

### What Already Exists in `backend/db/schema.sql`
- `sources` (RSS/APIs with reliability score)
- `ingested_items` (raw crawled feed items)
- `suggested_stories` (AI candidate briefs)
- `articles` (canonical published stories with JSONB blocks & verified sources)
- `uploads` (S3/R2 storage tracker)
- `channel_publications` (X, FB, LinkedIn, Newsletter syndication)

### What Needs to Be Added for Admin Login & Publishing

```sql
-- 1. Newsroom Users & Staff
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

-- 2. Auth Sessions (for revokable tokens / multi-device sessions)
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    user_agent TEXT,
    ip_address TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 3. Link Articles to User Author & Publisher
ALTER TABLE articles
    ADD COLUMN IF NOT EXISTS author_id UUID REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS published_by UUID REFERENCES users(id) ON DELETE SET NULL;
```

---

## 4. MVP API Endpoints

| Method | Endpoint | Access | Purpose |
| :--- | :--- | :--- | :--- |
| `POST` | `/api/auth/login` | Public | Verify credentials, set session cookie |
| `POST` | `/api/auth/logout` | Authenticated | Invalidate session cookie |
| `GET` | `/api/auth/me` | Authenticated | Return current logged-in editor profile |
| `GET` | `/api/articles` | Public / Admin | List published articles (or all articles with `?status=all` if admin) |
| `GET` | `/api/articles/:slug`| Public | Fetch single published story |
| `POST` | `/api/articles` | Authenticated | Create / Autosave draft article |
| `PUT` | `/api/articles/:id/publish` | Authenticated | Validate & publish article |
| `POST` | `/api/uploads/presign` | Authenticated | Generate Cloudflare R2 presigned upload URL |
| `GET` | `/api/candidates` | Authenticated | Fetch discovery inbox feed leads |
